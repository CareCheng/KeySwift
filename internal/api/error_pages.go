package api

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
)

// ErrorPageConfig 错误页面配置
type ErrorPageConfig struct {
	StatusCode  int
	Icon        string
	Title       string
	Message     string
	ShowRetry   bool
	RetryAfter  int    // 重试倒计时（秒），0 表示立即可重试
	GradientFrom string
	GradientTo   string
}

// 预定义错误页面配置
var errorPageConfigs = map[int]ErrorPageConfig{
	400: {
		StatusCode:   400,
		Icon:         "❌",
		Title:        "请求无效",
		Message:      "您的请求格式不正确，请检查后重试。",
		ShowRetry:    true,
		GradientFrom: "#ef4444",
		GradientTo:   "#f97316",
	},
	401: {
		StatusCode:   401,
		Icon:         "🔐",
		Title:        "请先登录",
		Message:      "您需要登录后才能访问此页面。",
		ShowRetry:    true,
		GradientFrom: "#8b5cf6",
		GradientTo:   "#6366f1",
	},
	403: {
		StatusCode:   403,
		Icon:         "🚫",
		Title:        "访问被拒绝",
		Message:      "您没有权限访问此页面。",
		ShowRetry:    true,
		GradientFrom: "#ef4444",
		GradientTo:   "#dc2626",
	},
	404: {
		StatusCode:   404,
		Icon:         "🔍",
		Title:        "页面未找到",
		Message:      "您访问的页面不存在或已被移除。",
		ShowRetry:    true,
		GradientFrom: "#6366f1",
		GradientTo:   "#8b5cf6",
	},
	429: {
		StatusCode:   429,
		Icon:         "⏳",
		Title:        "请求过于频繁",
		Message:      "您的操作太快了，服务器需要休息一下。",
		ShowRetry:    true,
		RetryAfter:   60,
		GradientFrom: "#f59e0b",
		GradientTo:   "#ef4444",
	},
	500: {
		StatusCode:   500,
		Icon:         "⚙️",
		Title:        "服务器错误",
		Message:      "服务器遇到了一些问题，请稍后再试。",
		ShowRetry:    true,
		GradientFrom: "#64748b",
		GradientTo:   "#475569",
	},
	502: {
		StatusCode:   502,
		Icon:         "🔌",
		Title:        "网关错误",
		Message:      "服务器暂时无法处理您的请求。",
		ShowRetry:    true,
		GradientFrom: "#64748b",
		GradientTo:   "#475569",
	},
	503: {
		StatusCode:   503,
		Icon:         "🔧",
		Title:        "服务不可用",
		Message:      "服务器正在维护中，请稍后再试。",
		ShowRetry:    true,
		GradientFrom: "#f59e0b",
		GradientTo:   "#d97706",
	},
}

// RenderErrorPage 渲染错误页面
// 参数：
//   - c: Gin 上下文
//   - statusCode: HTTP 状态码
//   - customMessage: 自定义错误消息（可选，为空则使用默认消息）
//   - retryAfter: 重试等待时间（秒），仅对 429 有效
func RenderErrorPage(c *gin.Context, statusCode int, customMessage string, retryAfter int) {
	// 检查是否是 API 请求
	accept := c.GetHeader("Accept")
	path := c.Request.URL.Path
	isAPIRequest := strings.HasPrefix(path, "/api/") || strings.Contains(accept, "application/json")

	if isAPIRequest {
		// API 请求返回 JSON
		response := gin.H{
			"success": false,
			"error":   customMessage,
		}
		if retryAfter > 0 {
			response["retry_after"] = retryAfter
		}
		c.JSON(statusCode, response)
		return
	}

	// 获取错误页面配置
	config, exists := errorPageConfigs[statusCode]
	if !exists {
		// 默认配置
		config = ErrorPageConfig{
			StatusCode:   statusCode,
			Icon:         "⚠️",
			Title:        "发生错误",
			Message:      "请求处理过程中发生了错误。",
			ShowRetry:    true,
			GradientFrom: "#64748b",
			GradientTo:   "#475569",
		}
	}

	// 使用自定义消息
	if customMessage != "" {
		config.Message = customMessage
	}

	// 使用自定义重试时间
	if retryAfter > 0 {
		config.RetryAfter = retryAfter
	}

	// 渲染 HTML 页面
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(statusCode, generateErrorHTML(config))
}

// generateErrorHTML 生成错误页面 HTML
func generateErrorHTML(config ErrorPageConfig) string {
	// 倒计时相关的 HTML 和 JS
	countdownHTML := ""
	countdownJS := ""
	btnInitState := ""
	btnInitText := "重新加载"

	if config.RetryAfter > 0 {
		countdownHTML = fmt.Sprintf(`
        <div class="countdown">
            <span class="countdown-icon">⏱️</span>
            <div>
                <div class="countdown-text">请等待</div>
                <div class="countdown-time" id="countdown">%d</div>
            </div>
            <span class="countdown-text">秒</span>
        </div>
        <br>`, config.RetryAfter)

		countdownJS = fmt.Sprintf(`
        let seconds = %d;
        const countdownEl = document.getElementById('countdown');
        const retryBtn = document.getElementById('retryBtn');
        const btnText = document.getElementById('btnText');
        
        const timer = setInterval(() => {
            seconds--;
            countdownEl.textContent = seconds;
            
            if (seconds <= 0) {
                clearInterval(timer);
                retryBtn.disabled = false;
                btnText.textContent = '重新加载';
                countdownEl.textContent = '0';
            }
        }, 1000);`, config.RetryAfter)

		btnInitState = "disabled"
		btnInitText = "请等待..."
	}

	// 重试按钮
	retryBtnHTML := ""
	if config.ShowRetry {
		retryBtnHTML = fmt.Sprintf(`
        <button class="btn" id="retryBtn" %s onclick="location.reload()">
            <span>🔄</span>
            <span id="btnText">%s</span>
        </button>`, btnInitState, btnInitText)
	}

	// 401 特殊处理：显示登录按钮
	loginBtnHTML := ""
	if config.StatusCode == 401 {
		loginBtnHTML = `
        <a href="/login" class="btn btn-primary" style="margin-right: 12px;">
            <span>🔑</span>
            <span>去登录</span>
        </a>`
		retryBtnHTML = `
        <a href="/" class="btn btn-secondary">
            <span>🏠</span>
            <span>返回首页</span>
        </a>`
	}

	// 404 特殊处理：显示返回首页按钮
	if config.StatusCode == 404 {
		retryBtnHTML = `
        <a href="/" class="btn">
            <span>🏠</span>
            <span>返回首页</span>
        </a>`
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s - %d</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            background: linear-gradient(135deg, #1a1a2e 0%%, #16213e 50%%, #0f3460 100%%);
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            color: #e2e8f0;
            padding: 20px;
        }
        .container {
            text-align: center;
            max-width: 500px;
            padding: 40px;
            background: rgba(255, 255, 255, 0.05);
            border-radius: 24px;
            border: 1px solid rgba(255, 255, 255, 0.1);
            backdrop-filter: blur(10px);
            box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.5);
        }
        .status-code {
            font-size: 14px;
            color: #64748b;
            margin-bottom: 16px;
            font-weight: 500;
        }
        .icon {
            font-size: 80px;
            margin-bottom: 24px;
            animation: pulse 2s ease-in-out infinite;
        }
        @keyframes pulse {
            0%%, 100%% { transform: scale(1); }
            50%% { transform: scale(1.1); }
        }
        h1 {
            font-size: 28px;
            font-weight: 600;
            margin-bottom: 16px;
            background: linear-gradient(135deg, %s, %s);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            background-clip: text;
        }
        .message {
            font-size: 16px;
            color: #94a3b8;
            margin-bottom: 32px;
            line-height: 1.6;
        }
        .countdown {
            display: inline-flex;
            align-items: center;
            gap: 12px;
            padding: 16px 32px;
            background: rgba(245, 158, 11, 0.1);
            border: 1px solid rgba(245, 158, 11, 0.3);
            border-radius: 12px;
            margin-bottom: 32px;
        }
        .countdown-icon {
            font-size: 24px;
        }
        .countdown-text {
            font-size: 14px;
            color: #94a3b8;
        }
        .countdown-time {
            font-size: 32px;
            font-weight: 700;
            color: #f59e0b;
            font-variant-numeric: tabular-nums;
        }
        .btn-group {
            display: flex;
            justify-content: center;
            gap: 12px;
            flex-wrap: wrap;
        }
        .btn {
            display: inline-flex;
            align-items: center;
            gap: 8px;
            padding: 14px 28px;
            background: linear-gradient(135deg, #3b82f6, #8b5cf6);
            color: white;
            text-decoration: none;
            border-radius: 12px;
            font-weight: 500;
            font-size: 16px;
            transition: all 0.3s ease;
            border: none;
            cursor: pointer;
        }
        .btn:hover {
            transform: translateY(-2px);
            box-shadow: 0 10px 20px -10px rgba(59, 130, 246, 0.5);
        }
        .btn:disabled {
            opacity: 0.5;
            cursor: not-allowed;
            transform: none;
        }
        .btn-primary {
            background: linear-gradient(135deg, #10b981, #059669);
        }
        .btn-primary:hover {
            box-shadow: 0 10px 20px -10px rgba(16, 185, 129, 0.5);
        }
        .btn-secondary {
            background: linear-gradient(135deg, #64748b, #475569);
        }
        .btn-secondary:hover {
            box-shadow: 0 10px 20px -10px rgba(100, 116, 139, 0.5);
        }
        .tips {
            margin-top: 32px;
            padding-top: 24px;
            border-top: 1px solid rgba(255, 255, 255, 0.1);
        }
        .tips-title {
            font-size: 14px;
            color: #64748b;
            margin-bottom: 12px;
        }
        .tips-list {
            list-style: none;
            font-size: 13px;
            color: #475569;
        }
        .tips-list li {
            padding: 4px 0;
        }
        .tips-list li::before {
            content: "•";
            color: #3b82f6;
            margin-right: 8px;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="status-code">错误代码: %d</div>
        <div class="icon">%s</div>
        <h1>%s</h1>
        <p class="message">%s</p>
        %s
        <div class="btn-group">
            %s
            %s
        </div>
        <div class="tips">
            <div class="tips-title">温馨提示</div>
            <ul class="tips-list">
                <li>检查网址是否正确</li>
                <li>尝试刷新页面</li>
                <li>如持续出现此问题，请联系管理员</li>
            </ul>
        </div>
    </div>
    <script>%s</script>
</body>
</html>`,
		config.Title, config.StatusCode,
		config.GradientFrom, config.GradientTo,
		config.StatusCode, config.Icon, config.Title, config.Message,
		countdownHTML, loginBtnHTML, retryBtnHTML, countdownJS)
}
