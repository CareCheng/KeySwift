package plugin

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"time"
)

const (
	ControlMessageHello       = "plugin.hello"
	ControlMessageHostAccept  = "host.accept"
	ControlMessageHostReject  = "host.reject"
	ControlMessageReady       = "plugin.ready"
	ControlMessageHeartbeat   = "plugin.heartbeat"
	ControlMessageFault       = "plugin.fault"
	ControlMessageHostStop    = "host.stop"
	ControlMessagePluginStop  = "plugin.stop_ack"
	ControlMessageInvocation  = "host.invoke"
	ControlMessageInvokeReply = "plugin.invoke_reply"
)

// ControlFrame 是插件控制面统一 JSON 行消息。
type ControlFrame struct {
	Type       string       `json:"type"`
	MessageID  string       `json:"messageId"`
	PluginID   string       `json:"pluginId"`
	Instance   string       `json:"instanceId"`
	Timestamp  time.Time    `json:"timestamp"`
	Payload    any          `json:"payload,omitempty"`
	Extensions ExtensionMap `json:"extensions,omitempty"`
}

// PluginHelloPayload 是插件进程启动后的首包。
type PluginHelloPayload struct {
	PluginID            string   `json:"pluginId"`
	Version             string   `json:"version"`
	ProtocolVersion     string   `json:"protocolVersion"`
	InstanceID          string   `json:"instanceId"`
	CapabilityHash      string   `json:"capabilityHash"`
	SupportedTransports []string `json:"supportedTransports"`
}

// HostAcceptPayload 是宿主接受握手后的确认包。
type HostAcceptPayload struct {
	ProtocolVersion string         `json:"protocolVersion"`
	HeartbeatMs     int            `json:"heartbeatMs"`
	Config          map[string]any `json:"config"`
	Extensions      map[string]any `json:"extensions"`
}

// FaultPayload 是插件上报故障时的结构。
type FaultPayload struct {
	FaultType   string `json:"faultType"`
	FaultReason string `json:"faultReason"`
	StackTrace  string `json:"stackTrace"`
	Retryable   bool   `json:"retryable"`
}

// EncodeControlFrame 将控制帧编码为 JSON 行。
func EncodeControlFrame(frame ControlFrame) ([]byte, error) {
	if strings.TrimSpace(frame.Type) == "" {
		return nil, errors.New("控制消息类型不能为空")
	}
	if frame.Timestamp.IsZero() {
		frame.Timestamp = time.Now()
	}
	data, err := json.Marshal(frame)
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

// DecodeControlFrame 从单行 JSON 解码控制帧。
func DecodeControlFrame(line []byte) (ControlFrame, error) {
	var frame ControlFrame
	if len(strings.TrimSpace(string(line))) == 0 {
		return frame, errors.New("控制消息为空")
	}
	if err := json.Unmarshal(line, &frame); err != nil {
		return frame, err
	}
	if strings.TrimSpace(frame.Type) == "" {
		return frame, errors.New("控制消息类型不能为空")
	}
	return frame, nil
}

// ReadControlFrame 从 reader 读取一条 JSON 行控制帧。
func ReadControlFrame(reader *bufio.Reader) (ControlFrame, error) {
	line, err := reader.ReadBytes('\n')
	if err != nil {
		if errors.Is(err, io.EOF) && len(line) > 0 {
			return DecodeControlFrame(line)
		}
		return ControlFrame{}, err
	}
	return DecodeControlFrame(line)
}

// WriteControlFrame 写入一条 JSON 行控制帧。
func WriteControlFrame(writer io.Writer, frame ControlFrame) error {
	data, err := EncodeControlFrame(frame)
	if err != nil {
		return err
	}
	_, err = writer.Write(data)
	return err
}
