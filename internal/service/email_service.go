package service

import (
	"crypto/rand"
	"crypto/tls"
	"errors"
	"fmt"
	"math/big"
	"net"
	"net/smtp"
	"strings"
	"time"

	"user-frontend/internal/config"
	"user-frontend/internal/model"
	"user-frontend/internal/repository"
)

type EmailService struct {
	repo *repository.Repository
	cfg  *config.EmailConfig
}

func NewEmailService(repo *repository.Repository, cfg *config.EmailConfig) *EmailService {
	return &EmailService{repo: repo, cfg: cfg}
}

// UpdateConfig 更新邮箱配置（不重建服务实例）
// 参数：
//   - cfg: 新的邮箱配置
func (s *EmailService) UpdateConfig(cfg *config.EmailConfig) {
	s.cfg = cfg
}

// GenerateCode 生成验证码（长度可配置）
func (s *EmailService) GenerateCode() string {
	const digits = "0123456789"
	codeLen := s.cfg.CodeLength
	if codeLen <= 0 {
		codeLen = 6 // 默认6位
	}
	code := make([]byte, codeLen)
	for i := range code {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		code[i] = digits[n.Int64()]
	}
	return string(code)
}

// SendVerifyCode 发送验证码邮件
func (s *EmailService) SendVerifyCode(email, codeType string) error {
	if !s.cfg.Enabled {
		return errors.New("邮箱服务未启用")
	}

	if email == "" {
		return errors.New("邮箱地址不能为空")
	}

	// 生成验证码
	code := s.GenerateCode()

	// 保存到数据库
	verifyCode := &model.EmailVerifyCode{
		Email:     email,
		Code:      code,
		Type:      codeType,
		ExpiresAt: time.Now().Add(10 * time.Minute), // 10分钟有效
	}
	if err := s.repo.CreateEmailVerifyCode(verifyCode); err != nil {
		return err
	}

	// 发送邮件
	subject := s.getSubject(codeType)
	body := s.getEmailBody(code, codeType)

	return s.SendEmail(email, subject, body)
}

// VerifyCode 验证验证码
func (s *EmailService) VerifyCode(email, code, codeType string) bool {
	verifyCode, err := s.repo.GetLatestEmailVerifyCode(email, codeType)
	if err != nil {
		return false
	}

	if verifyCode.Used {
		return false
	}

	if time.Now().After(verifyCode.ExpiresAt) {
		return false
	}

	if verifyCode.Code != code {
		return false
	}

	// 标记为已使用
	s.repo.MarkEmailVerifyCodeUsed(verifyCode.ID)
	return true
}

// CheckCodeValid 检查验证码是否有效（不消耗验证码）
// 用于实时验证，不会标记验证码为已使用
func (s *EmailService) CheckCodeValid(email, code, codeType string) bool {
	verifyCode, err := s.repo.GetLatestEmailVerifyCode(email, codeType)
	if err != nil {
		return false
	}

	if verifyCode.Used {
		return false
	}

	if time.Now().After(verifyCode.ExpiresAt) {
		return false
	}

	return verifyCode.Code == code
}

func (s *EmailService) getSubject(codeType string) string {
	systemTitle := config.GlobalConfig.ServerConfig.SystemTitle
	switch codeType {
	case "register":
		return fmt.Sprintf("[%s] 注册验证码", systemTitle)
	case "login":
		return fmt.Sprintf("[%s] 登录验证码", systemTitle)
	case "reset_password":
		return fmt.Sprintf("[%s] 重置密码验证码", systemTitle)
	case "enable_2fa":
		return fmt.Sprintf("[%s] 安全验证码", systemTitle)
	default:
		return fmt.Sprintf("[%s] 验证码", systemTitle)
	}
}

func (s *EmailService) getEmailBody(code, codeType string) string {
	systemTitle := config.GlobalConfig.ServerConfig.SystemTitle
	action := "操作"
	switch codeType {
	case "register":
		action = "注册账号"
	case "login":
		action = "登录账号"
	case "reset_password":
		action = "重置密码"
	case "enable_2fa":
		action = "安全设置"
	}

	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h2 style="color: #667eea;">%s</h2>
        <p>您好，</p>
        <p>您正在进行<strong>%s</strong>操作，验证码为：</p>
        <div style="background: #f5f5f5; padding: 20px; text-align: center; margin: 20px 0; border-radius: 8px;">
            <span style="font-size: 32px; font-weight: bold; letter-spacing: 8px; color: #667eea;">%s</span>
        </div>
        <p>验证码有效期为 <strong>10分钟</strong>，请尽快使用。</p>
        <p style="color: #999; font-size: 12px;">如果这不是您本人的操作，请忽略此邮件。</p>
        <hr style="border: none; border-top: 1px solid #eee; margin: 20px 0;">
        <p style="color: #999; font-size: 12px;">此邮件由系统自动发送，请勿回复。</p>
    </div>
</body>
</html>
`, systemTitle, action, code)
}

// SendEmail 发送邮件（公开方法，供其他服务调用）
// 参数：
//   - to: 收件人邮箱
//   - subject: 邮件主题
//   - body: 邮件内容（HTML格式）
// 返回：
//   - 错误信息
func (s *EmailService) SendEmail(to, subject, body string) error {
	from := s.cfg.FromEmail
	if from == "" {
		from = s.cfg.SMTPUser
	}

	headers := make(map[string]string)
	headers["From"] = fmt.Sprintf("%s <%s>", s.cfg.FromName, from)
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"

	var msg strings.Builder
	for k, v := range headers {
		msg.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	msg.WriteString("\r\n")
	msg.WriteString(body)

	addr := fmt.Sprintf("%s:%d", s.cfg.SMTPHost, s.cfg.SMTPPort)
	auth := smtp.PlainAuth("", s.cfg.SMTPUser, s.cfg.SMTPPassword, s.cfg.SMTPHost)

	switch s.cfg.Encryption {
	case "ssl":
		return s.sendEmailSSL(addr, auth, from, to, msg.String())
	case "starttls":
		return s.sendEmailSTARTTLS(addr, auth, from, to, msg.String())
	default:
		// 无加密
		return smtp.SendMail(addr, auth, from, []string{to}, []byte(msg.String()))
	}
}

// sendEmailSSL 使用 SSL/TLS 发送邮件（端口 465）
func (s *EmailService) sendEmailSSL(addr string, auth smtp.Auth, from, to, msg string) error {
	tlsConfig := &tls.Config{
		ServerName: s.cfg.SMTPHost,
	}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return err
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, s.cfg.SMTPHost)
	if err != nil {
		return err
	}
	defer client.Close()

	if err = client.Auth(auth); err != nil {
		return err
	}

	if err = client.Mail(from); err != nil {
		return err
	}

	if err = client.Rcpt(to); err != nil {
		return err
	}

	w, err := client.Data()
	if err != nil {
		return err
	}

	_, err = w.Write([]byte(msg))
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	return client.Quit()
}

// sendEmailSTARTTLS 使用 STARTTLS 发送邮件（端口 587）
func (s *EmailService) sendEmailSTARTTLS(addr string, auth smtp.Auth, from, to, msg string) error {
	// 先建立普通 TCP 连接
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, s.cfg.SMTPHost)
	if err != nil {
		return err
	}
	defer client.Close()

	// 发送 EHLO
	if err = client.Hello("localhost"); err != nil {
		return err
	}

	// 升级到 TLS
	tlsConfig := &tls.Config{
		ServerName: s.cfg.SMTPHost,
	}
	if err = client.StartTLS(tlsConfig); err != nil {
		return err
	}

	// 认证
	if err = client.Auth(auth); err != nil {
		return err
	}

	if err = client.Mail(from); err != nil {
		return err
	}

	if err = client.Rcpt(to); err != nil {
		return err
	}

	w, err := client.Data()
	if err != nil {
		return err
	}

	_, err = w.Write([]byte(msg))
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	return client.Quit()
}
