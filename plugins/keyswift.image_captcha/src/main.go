// Package main 是 KeySwift 官方图片人机验证独立插件入口。
package main

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	mathrand "math/rand"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	providerID   = "keyswift.image_captcha"
	providerType = "image_captcha"
)

var challengeIDPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]{8,128}$`)

type request struct {
	Action         string         `json:"action"`
	Scope          string         `json:"scope"`
	ProviderID     string         `json:"provider_id"`
	ProviderType   string         `json:"provider_type"`
	Payload        *payload       `json:"payload"`
	Config         map[string]any `json:"config"`
	RequestContext map[string]any `json:"request_context"`
}

type payload struct {
	ChallengeID string `json:"challenge_id"`
	Answer      string `json:"answer"`
}

type response struct {
	Success      bool               `json:"success"`
	Error        string             `json:"error,omitempty"`
	ConfigStatus string             `json:"config_status,omitempty"`
	HealthStatus string             `json:"health_status,omitempty"`
	PublicConfig map[string]any     `json:"public_config,omitempty"`
	Challenge    *challengeResponse `json:"challenge,omitempty"`
	Verified     bool               `json:"verified,omitempty"`
	Reason       string             `json:"reason,omitempty"`
}

type challengeResponse struct {
	ProviderID   string         `json:"provider_id"`
	ProviderType string         `json:"provider_type"`
	Scope        string         `json:"scope"`
	ChallengeID  string         `json:"challenge_id,omitempty"`
	Image        string         `json:"image,omitempty"`
	ExpiresAt    *time.Time     `json:"expires_at,omitempty"`
	PublicConfig map[string]any `json:"public_config,omitempty"`
}

type challengeState struct {
	AnswerHash string    `json:"answer_hash"`
	Salt       string    `json:"salt"`
	ExpiresAt  time.Time `json:"expires_at"`
	CreatedAt  time.Time `json:"created_at"`
}

func main() {
	var req request
	if err := json.NewDecoder(os.Stdin).Decode(&req); err != nil {
		write(response{Success: false, Error: "请求解析失败"})
		return
	}
	resp := handle(req)
	write(resp)
}

func handle(req request) response {
	switch strings.TrimSpace(req.Action) {
	case "public_config":
		return response{Success: true, PublicConfig: publicConfig(req.Config)}
	case "config_status":
		return response{Success: true, ConfigStatus: "ready"}
	case "health":
		if err := ensureChallengeDir(); err != nil {
			return response{Success: true, HealthStatus: "unhealthy", Error: err.Error()}
		}
		return response{Success: true, HealthStatus: "ready"}
	case "create_challenge":
		return createChallenge(req)
	case "verify":
		return verify(req)
	default:
		return response{Success: false, Error: "不支持的人机验证插件动作"}
	}
}

func createChallenge(req request) response {
	if err := cleanupExpiredChallenges(); err != nil {
		return response{Success: false, Error: err.Error()}
	}
	cfg := captchaConfig(req.Config)
	id, err := randomHex(16)
	if err != nil {
		return response{Success: false, Error: "生成挑战 ID 失败"}
	}
	answer, err := randomDigits(cfg.length)
	if err != nil {
		return response{Success: false, Error: "生成人机验证答案失败"}
	}
	imageData, err := renderCaptchaImage(answer, cfg)
	if err != nil {
		return response{Success: false, Error: "生成人机验证挑战失败"}
	}
	expiresAt := time.Now().Add(time.Duration(cfg.ttlSeconds) * time.Second)
	state, err := newChallengeState(answer, expiresAt)
	if err != nil {
		return response{Success: false, Error: err.Error()}
	}
	if err := saveChallengeState(id, state); err != nil {
		return response{Success: false, Error: err.Error()}
	}
	return response{
		Success: true,
		Challenge: &challengeResponse{
			ProviderID:   providerID,
			ProviderType: providerType,
			Scope:        strings.TrimSpace(req.Scope),
			ChallengeID:  id,
			Image:        imageData,
			ExpiresAt:    &expiresAt,
			PublicConfig: publicConfig(req.Config),
		},
	}
}

func verify(req request) response {
	if req.Payload == nil || strings.TrimSpace(req.Payload.ChallengeID) == "" || strings.TrimSpace(req.Payload.Answer) == "" {
		return response{Success: true, Verified: false, Reason: "请完成人机验证"}
	}
	state, err := consumeChallengeState(req.Payload.ChallengeID)
	if err != nil {
		return response{Success: true, Verified: false, Reason: "人机验证失败"}
	}
	if time.Now().After(state.ExpiresAt) {
		return response{Success: true, Verified: false, Reason: "人机验证已过期"}
	}
	if state.AnswerHash != hashAnswer(state.Salt, req.Payload.Answer) {
		return response{Success: true, Verified: false, Reason: "人机验证失败"}
	}
	return response{Success: true, Verified: true}
}

type runtimeConfig struct {
	width      int
	height     int
	length     int
	dotCount   int
	ttlSeconds int
}

func captchaConfig(values map[string]any) runtimeConfig {
	return runtimeConfig{
		width:      intFromAny(values["width"], 240, 120, 480),
		height:     intFromAny(values["height"], 80, 40, 200),
		length:     intFromAny(values["length"], 4, 4, 8),
		dotCount:   intFromAny(values["dot_count"], 80, 0, 300),
		ttlSeconds: intFromAny(values["ttl_seconds"], 300, 60, 1800),
	}
}

func publicConfig(values map[string]any) map[string]any {
	cfg := captchaConfig(values)
	return map[string]any{
		"width":       cfg.width,
		"height":      cfg.height,
		"length":      cfg.length,
		"ttl_seconds": cfg.ttlSeconds,
	}
}

func newChallengeState(answer string, expiresAt time.Time) (challengeState, error) {
	salt, err := randomHex(16)
	if err != nil {
		return challengeState{}, errors.New("生成挑战状态失败")
	}
	return challengeState{
		AnswerHash: hashAnswer(salt, answer),
		Salt:       salt,
		ExpiresAt:  expiresAt,
		CreatedAt:  time.Now(),
	}, nil
}

func saveChallengeState(id string, state challengeState) error {
	path, err := challengePath(id)
	if err != nil {
		return err
	}
	if err := ensureChallengeDir(); err != nil {
		return err
	}
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func consumeChallengeState(id string) (challengeState, error) {
	path, err := challengePath(id)
	if err != nil {
		return challengeState{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return challengeState{}, err
	}
	_ = os.Remove(path)
	var state challengeState
	if err := json.Unmarshal(data, &state); err != nil {
		return challengeState{}, err
	}
	return state, nil
}

func cleanupExpiredChallenges() error {
	dir := challengeDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	now := time.Now()
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var state challengeState
		if json.Unmarshal(data, &state) != nil || now.After(state.ExpiresAt) {
			_ = os.Remove(path)
		}
	}
	return nil
}

func challengePath(id string) (string, error) {
	if !challengeIDPattern.MatchString(id) {
		return "", errors.New("挑战 ID 不合法")
	}
	return filepath.Join(challengeDir(), id+".json"), nil
}

func challengeDir() string {
	dataDir := strings.TrimSpace(os.Getenv("KEYSWIFT_PLUGIN_DATA_DIR"))
	if dataDir == "" {
		dataDir = filepath.Join(".", "data")
	}
	return filepath.Join(dataDir, "challenges")
}

func ensureChallengeDir() error {
	return os.MkdirAll(challengeDir(), 0755)
}

func hashAnswer(salt string, answer string) string {
	normalized := strings.ToLower(strings.TrimSpace(answer))
	sum := sha256.Sum256([]byte(salt + ":" + normalized))
	return hex.EncodeToString(sum[:])
}

func randomHex(size int) (string, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func randomDigits(length int) (string, error) {
	buf := make([]byte, length)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	var builder strings.Builder
	for _, item := range buf {
		builder.WriteByte(byte('0' + item%10))
	}
	return builder.String(), nil
}

func renderCaptchaImage(answer string, cfg runtimeConfig) (string, error) {
	img := image.NewRGBA(image.Rect(0, 0, cfg.width, cfg.height))
	draw.Draw(img, img.Bounds(), &image.Uniform{C: color.RGBA{R: 246, G: 248, B: 252, A: 255}}, image.Point{}, draw.Src)

	rnd := mathrand.New(mathrand.NewSource(time.Now().UnixNano()))
	for i := 0; i < cfg.dotCount; i++ {
		x := rnd.Intn(cfg.width)
		y := rnd.Intn(cfg.height)
		fillCircle(img, x, y, rnd.Intn(2)+1, color.RGBA{R: uint8(120 + rnd.Intn(100)), G: uint8(120 + rnd.Intn(100)), B: uint8(120 + rnd.Intn(100)), A: 120})
	}
	for i := 0; i < 5; i++ {
		drawLine(img, rnd.Intn(cfg.width), rnd.Intn(cfg.height), rnd.Intn(cfg.width), rnd.Intn(cfg.height), color.RGBA{R: 80, G: 110, B: 150, A: 90})
	}

	slotWidth := cfg.width / len(answer)
	digitWidth := maxInt(18, slotWidth-12)
	digitHeight := maxInt(34, cfg.height-24)
	thick := maxInt(4, digitWidth/7)
	for index, digit := range answer {
		x := index*slotWidth + (slotWidth-digitWidth)/2 + rnd.Intn(5) - 2
		y := (cfg.height-digitHeight)/2 + rnd.Intn(7) - 3
		c := color.RGBA{R: uint8(30 + rnd.Intn(70)), G: uint8(55 + rnd.Intn(70)), B: uint8(90 + rnd.Intn(90)), A: 255}
		drawSevenSegmentDigit(img, digit, x, y, digitWidth, digitHeight, thick, c)
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", err
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

func drawSevenSegmentDigit(img *image.RGBA, digit rune, x int, y int, w int, h int, t int, c color.Color) {
	segments := map[rune]string{
		'0': "abcedf",
		'1': "bc",
		'2': "abged",
		'3': "abgcd",
		'4': "fgbc",
		'5': "afgcd",
		'6': "afgecd",
		'7': "abc",
		'8': "abcdefg",
		'9': "abfgcd",
	}
	for _, segment := range segments[digit] {
		switch segment {
		case 'a':
			fillRect(img, x+t, y, w-2*t, t, c)
		case 'b':
			fillRect(img, x+w-t, y+t, t, h/2-t, c)
		case 'c':
			fillRect(img, x+w-t, y+h/2, t, h/2-t, c)
		case 'd':
			fillRect(img, x+t, y+h-t, w-2*t, t, c)
		case 'e':
			fillRect(img, x, y+h/2, t, h/2-t, c)
		case 'f':
			fillRect(img, x, y+t, t, h/2-t, c)
		case 'g':
			fillRect(img, x+t, y+h/2-t/2, w-2*t, t, c)
		}
	}
}

func fillRect(img *image.RGBA, x int, y int, w int, h int, c color.Color) {
	draw.Draw(img, image.Rect(x, y, x+w, y+h).Intersect(img.Bounds()), &image.Uniform{C: c}, image.Point{}, draw.Src)
}

func fillCircle(img *image.RGBA, cx int, cy int, r int, c color.Color) {
	for x := cx - r; x <= cx+r; x++ {
		for y := cy - r; y <= cy+r; y++ {
			if (x-cx)*(x-cx)+(y-cy)*(y-cy) <= r*r && image.Pt(x, y).In(img.Bounds()) {
				img.Set(x, y, c)
			}
		}
	}
}

func drawLine(img *image.RGBA, x0 int, y0 int, x1 int, y1 int, c color.Color) {
	dx := absInt(x1 - x0)
	dy := -absInt(y1 - y0)
	sx := -1
	if x0 < x1 {
		sx = 1
	}
	sy := -1
	if y0 < y1 {
		sy = 1
	}
	err := dx + dy
	for {
		if image.Pt(x0, y0).In(img.Bounds()) {
			img.Set(x0, y0, c)
		}
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 >= dy {
			err += dy
			x0 += sx
		}
		if e2 <= dx {
			err += dx
			y0 += sy
		}
	}
}

func absInt(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

func maxInt(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

func intFromAny(value any, fallback int, minValue int, maxValue int) int {
	result := fallback
	switch item := value.(type) {
	case float64:
		result = int(item)
	case int:
		result = item
	case string:
		if _, err := fmt.Sscanf(item, "%d", &result); err != nil {
			result = fallback
		}
	}
	if result < minValue {
		return minValue
	}
	if result > maxValue {
		return maxValue
	}
	return result
}

func write(resp response) {
	_ = json.NewEncoder(os.Stdout).Encode(resp)
}
