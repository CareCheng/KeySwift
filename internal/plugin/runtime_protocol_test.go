package plugin

import (
	"bufio"
	"bytes"
	"testing"
)

func TestControlFrameEncodeDecode(t *testing.T) {
	frame := ControlFrame{
		Type:      ControlMessageHello,
		MessageID: "msg-1",
		PluginID:  "diagnostic",
		Instance:  "pin-1",
		Payload: PluginHelloPayload{
			PluginID:        "diagnostic",
			Version:         "1.0.0",
			ProtocolVersion: HandshakeProtocol,
			InstanceID:      "pin-1",
		},
	}
	var buffer bytes.Buffer
	if err := WriteControlFrame(&buffer, frame); err != nil {
		t.Fatalf("写入控制帧失败: %v", err)
	}
	decoded, err := ReadControlFrame(bufio.NewReader(&buffer))
	if err != nil {
		t.Fatalf("读取控制帧失败: %v", err)
	}
	if decoded.Type != ControlMessageHello || decoded.PluginID != "diagnostic" {
		t.Fatalf("控制帧解码不符合预期: %#v", decoded)
	}
}
