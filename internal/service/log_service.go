package service

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
	"time"

	"user-frontend/internal/utils"
)

// LogEntry 日志条目结构
type LogEntry struct {
	ID        uint      `json:"id"`
	UserType  string    `json:"user_type"` // user, admin, security
	UserID    uint      `json:"user_id"`
	Username  string    `json:"username"`
	Action    string    `json:"action"`
	Category  string    `json:"category"` // 操作分类：auth, product, order, user, system, payment
	Target    string    `json:"target"`
	TargetID  string    `json:"target_id"`
	Detail    string    `json:"detail"`
	IP        string    `json:"ip"`
	UserAgent string    `json:"user_agent"`
	CreatedAt time.Time `json:"created_at"`
}

// LogConfig 日志配置
type LogConfig struct {
	EnableUserLog  bool `json:"enable_user_log"`  // 是否启用用户端日志
	EnableAdminLog bool `json:"enable_admin_log"` // 是否启用管理端日志
}

// LogService 日志服务（文件存储版本）
type LogService struct {
	logDir     string     // 日志目录
	mu         sync.Mutex // 写入锁
	config     LogConfig  // 日志配置
	configPath string     // 配置文件路径
}

// NewLogService 创建日志服务
// 日志文件保存在程序根目录的 server_log 文件夹下
func NewLogService() *LogService {
	// 获取程序根目录
	execPath, err := os.Executable()
	if err != nil {
		execPath = "."
	}
	rootDir := filepath.Dir(execPath)

	// 如果是开发环境，使用当前目录
	if _, err := os.Stat(filepath.Join(rootDir, "go.mod")); err != nil {
		// 生产环境
		rootDir = filepath.Dir(execPath)
	} else {
		// 开发环境，使用当前工作目录
		rootDir, _ = os.Getwd()
	}

	logDir := filepath.Join(rootDir, "server_log")
	configPath := filepath.Join(rootDir, "server_log", "log_config.json")

	// 确保日志目录存在
	if err := os.MkdirAll(logDir, 0755); err != nil {
		fmt.Printf("创建日志目录失败: %v\n", err)
	}

	svc := &LogService{
		logDir:     logDir,
		configPath: configPath,
		config: LogConfig{
			EnableUserLog:  false, // 默认关闭用户端日志
			EnableAdminLog: false, // 默认关闭管理端日志
		},
	}

	// 加载配置
	svc.loadConfig()

	return svc
}

// loadConfig 加载日志配置
func (s *LogService) loadConfig() {
	data, err := os.ReadFile(s.configPath)
	if err != nil {
		// 配置文件不存在，使用默认配置
		return
	}

	var config LogConfig
	if err := json.Unmarshal(data, &config); err != nil {
		fmt.Printf("解析日志配置失败: %v\n", err)
		return
	}

	s.config = config
}

// saveConfig 保存日志配置
func (s *LogService) saveConfig() error {
	data, err := json.MarshalIndent(s.config, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}

	if err := os.WriteFile(s.configPath, data, 0644); err != nil {
		return fmt.Errorf("保存配置失败: %v", err)
	}

	return nil
}

// GetLogConfig 获取日志配置
func (s *LogService) GetLogConfig() LogConfig {
	return s.config
}

// UpdateLogConfig 更新日志配置
func (s *LogService) UpdateLogConfig(config LogConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.config = config
	return s.saveConfig()
}

// getLogFilePath 获取指定日期的日志文件路径
func (s *LogService) getLogFilePath(date time.Time) string {
	filename := date.Format("2006-01-02") + ".csv"
	return filepath.Join(s.logDir, filename)
}

// getTodayLogFilePath 获取今天的日志文件路径
func (s *LogService) getTodayLogFilePath() string {
	return s.getLogFilePath(time.Now())
}

// LogOperation 记录操作日志
// 日志使用AES-256-GCM加密后保存到CSV文件
func (s *LogService) LogOperation(userType string, userID uint, username, action, category, target, targetID string, detail interface{}, ip, userAgent string) {
	// 检查日志开关
	if userType == "user" && !s.config.EnableUserLog {
		return
	}
	if userType == "admin" && !s.config.EnableAdminLog {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// 构建详情字符串
	detailStr := ""
	if detail != nil {
		if str, ok := detail.(string); ok {
			detailStr = str
		} else {
			bytes, _ := json.Marshal(detail)
			detailStr = string(bytes)
		}
	}

	// 创建日志条目
	entry := &LogEntry{
		UserType:  userType,
		UserID:    userID,
		Username:  username,
		Action:    action,
		Category:  category,
		Target:    target,
		TargetID:  targetID,
		Detail:    detailStr,
		IP:        ip,
		UserAgent: userAgent,
		CreatedAt: time.Now(),
	}

	// 写入日志文件
	if err := s.writeLogEntry(entry); err != nil {
		fmt.Printf("写入日志失败: %v\n", err)
	}
}

// writeLogEntry 将日志条目写入CSV文件
func (s *LogService) writeLogEntry(entry *LogEntry) error {
	filePath := s.getTodayLogFilePath()

	// 检查文件是否存在，不存在则创建并写入表头
	isNewFile := false
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		isNewFile = true
	}

	// 打开文件（追加模式）
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("打开日志文件失败: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 如果是新文件，写入表头（表头不加密）
	if isNewFile {
		header := []string{"user_type", "user_id", "username", "action", "category", "target", "target_id", "detail", "ip", "user_agent", "created_at"}
		if err := writer.Write(header); err != nil {
			return fmt.Errorf("写入表头失败: %v", err)
		}
	}

	// 构建日志记录
	record := []string{
		entry.UserType,
		strconv.FormatUint(uint64(entry.UserID), 10),
		entry.Username,
		entry.Action,
		entry.Category,
		entry.Target,
		entry.TargetID,
		entry.Detail,
		entry.IP,
		entry.UserAgent,
		entry.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	// 加密每个字段
	encryptedRecord := make([]string, len(record))
	for i, field := range record {
		encrypted, err := utils.AESEncrypt(field)
		if err != nil {
			// 加密失败时使用原文（不应该发生）
			encryptedRecord[i] = field
		} else {
			encryptedRecord[i] = encrypted
		}
	}

	// 写入加密后的记录
	if err := writer.Write(encryptedRecord); err != nil {
		return fmt.Errorf("写入日志记录失败: %v", err)
	}

	return nil
}

// LogUserAction 记录用户操作
func (s *LogService) LogUserAction(userID uint, username, action, category, target, targetID string, detail interface{}, ip, userAgent string) {
	s.LogOperation("user", userID, username, action, category, target, targetID, detail, ip, userAgent)
}

// LogUserActionSimple 记录用户操作（简化版，自动推断分类）。
func (s *LogService) LogUserActionSimple(userID uint, username, action, target, targetID string, detail interface{}, ip, userAgent string) {
	category := inferCategory(target)
	s.LogOperation("user", userID, username, action, category, target, targetID, detail, ip, userAgent)
}

// LogAdminAction 记录管理员操作
func (s *LogService) LogAdminAction(username, action, category, target, targetID string, detail interface{}, ip, userAgent string) {
	s.LogOperation("admin", 0, username, action, category, target, targetID, detail, ip, userAgent)
}

// LogAdminActionSimple 记录管理员操作（简化版，自动推断分类）。
func (s *LogService) LogAdminActionSimple(username, action, target, targetID string, detail interface{}, ip, userAgent string) {
	category := inferCategory(target)
	s.LogOperation("admin", 0, username, action, category, target, targetID, detail, ip, userAgent)
}

// inferCategory 根据 target 推断分类
func inferCategory(target string) string {
	switch target {
	case "user", "account":
		return "user"
	case "product", "category":
		return "product"
	case "order":
		return "order"
	case "payment", "balance":
		return "payment"
	case "whitelist", "blacklist", "security", "encryption_key", "role", "admin":
		return "system"
	default:
		return "system"
	}
}

// LogSecurityEvent 记录安全事件
func (s *LogService) LogSecurityEvent(action, ip, userAgent string, detail interface{}) {
	// 安全事件始终记录，不受开关控制
	s.mu.Lock()
	defer s.mu.Unlock()

	detailStr := ""
	if detail != nil {
		if str, ok := detail.(string); ok {
			detailStr = str
		} else {
			bytes, _ := json.Marshal(detail)
			detailStr = string(bytes)
		}
	}

	entry := &LogEntry{
		UserType:  "security",
		UserID:    0,
		Username:  "system",
		Action:    action,
		Category:  "security",
		Target:    "security",
		TargetID:  "",
		Detail:    detailStr,
		IP:        ip,
		UserAgent: userAgent,
		CreatedAt: time.Now(),
	}

	if err := s.writeLogEntry(entry); err != nil {
		fmt.Printf("写入日志失败: %v\n", err)
	}
}

// GetOperationLogs 获取操作日志（从文件读取并解密）
// date: 日期字符串，格式为 YYYY-MM-DD，为空则使用今天
// page, pageSize: 分页参数
// userType, action, category: 过滤条件
func (s *LogService) GetOperationLogs(date string, page, pageSize int, userType, action, category string) ([]LogEntry, int64, error) {
	// 解析日期
	var logDate time.Time
	if date == "" {
		logDate = time.Now()
	} else {
		var err error
		logDate, err = time.Parse("2006-01-02", date)
		if err != nil {
			return nil, 0, fmt.Errorf("日期格式错误: %v", err)
		}
	}

	filePath := s.getLogFilePath(logDate)

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return []LogEntry{}, 0, nil
	}

	// 读取并解密日志
	entries, err := s.readLogFile(filePath)
	if err != nil {
		return nil, 0, err
	}

	// 过滤
	var filtered []LogEntry
	for _, entry := range entries {
		if userType != "" && entry.UserType != userType {
			continue
		}
		if action != "" && entry.Action != action {
			continue
		}
		if category != "" && entry.Category != category {
			continue
		}
		filtered = append(filtered, entry)
	}

	// 按时间倒序排序
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].CreatedAt.After(filtered[j].CreatedAt)
	})

	total := int64(len(filtered))

	// 分页
	start := (page - 1) * pageSize
	end := start + pageSize
	if start >= len(filtered) {
		return []LogEntry{}, total, nil
	}
	if end > len(filtered) {
		end = len(filtered)
	}

	// 为返回的日志添加ID（用于前端显示）
	result := filtered[start:end]
	for i := range result {
		result[i].ID = uint(total - int64(start) - int64(i))
	}

	return result, total, nil
}

// readLogFile 读取并解密日志文件
func (s *LogService) readLogFile(filePath string) ([]LogEntry, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("打开日志文件失败: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("读取CSV失败: %v", err)
	}

	if len(records) <= 1 {
		return []LogEntry{}, nil
	}

	var entries []LogEntry
	// 跳过表头
	for i := 1; i < len(records); i++ {
		record := records[i]
		if len(record) < 11 {
			continue
		}

		// 解密每个字段
		decrypted := make([]string, len(record))
		for j, field := range record {
			dec, err := utils.AESDecrypt(field)
			if err != nil {
				return nil, fmt.Errorf("日志字段解密失败: %v", err)
			}
			decrypted[j] = dec
		}

		// 解析用户ID
		userID, _ := strconv.ParseUint(decrypted[1], 10, 32)

		// 解析时间
		createdAt, _ := time.Parse("2006-01-02 15:04:05", decrypted[10])

		entry := LogEntry{
			UserType:  decrypted[0],
			UserID:    uint(userID),
			Username:  decrypted[2],
			Action:    decrypted[3],
			Category:  decrypted[4],
			Target:    decrypted[5],
			TargetID:  decrypted[6],
			Detail:    decrypted[7],
			IP:        decrypted[8],
			UserAgent: decrypted[9],
			CreatedAt: createdAt,
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

// GetAvailableLogDates 获取可用的日志日期列表
func (s *LogService) GetAvailableLogDates() ([]string, error) {
	files, err := os.ReadDir(s.logDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("读取日志目录失败: %v", err)
	}

	var dates []string
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		name := file.Name()
		if len(name) == 14 && name[10:] == ".csv" {
			// 格式: YYYY-MM-DD.csv
			dates = append(dates, name[:10])
		}
	}

	// 按日期倒序排序
	sort.Sort(sort.Reverse(sort.StringSlice(dates)))

	return dates, nil
}

// GetUserOperationLogs 获取用户操作日志。
// 注意：此方法需要遍历所有日志文件，性能较低
func (s *LogService) GetUserOperationLogs(userID uint, page, pageSize int) ([]LogEntry, int64, error) {
	// 获取所有可用日期
	dates, err := s.GetAvailableLogDates()
	if err != nil {
		return nil, 0, err
	}

	var allEntries []LogEntry
	for _, date := range dates {
		logDate, _ := time.Parse("2006-01-02", date)
		filePath := s.getLogFilePath(logDate)

		entries, err := s.readLogFile(filePath)
		if err != nil {
			continue
		}

		// 过滤指定用户的日志
		for _, entry := range entries {
			if entry.UserType == "user" && entry.UserID == userID {
				allEntries = append(allEntries, entry)
			}
		}
	}

	// 按时间倒序排序
	sort.Slice(allEntries, func(i, j int) bool {
		return allEntries[i].CreatedAt.After(allEntries[j].CreatedAt)
	})

	total := int64(len(allEntries))

	// 分页
	start := (page - 1) * pageSize
	end := start + pageSize
	if start >= len(allEntries) {
		return []LogEntry{}, total, nil
	}
	if end > len(allEntries) {
		end = len(allEntries)
	}

	result := allEntries[start:end]
	for i := range result {
		result[i].ID = uint(total - int64(start) - int64(i))
	}

	return result, total, nil
}
