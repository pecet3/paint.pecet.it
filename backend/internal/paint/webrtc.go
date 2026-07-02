package paint

import (
	"context"
	"encoding/json"

	"paint.pecet.it/pkg/ward/wardsocket"
)

type SignalPayload struct {
	TargetUUID string          `json:"targetUuid"`
	SenderUUID string          `json:"senderUuid"`
	SignalType string          `json:"signalType"`
	Data       json.RawMessage `json:"data"`
}

func (p *PaintRoom) handleSignal(ctx context.Context, e *wardsocket.Event) {

	var payload SignalPayload
	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		e.Client.Request.Log("Invalid WebRTC signal payload:", err)
		return
	}

	payload.SenderUUID = e.Client.Request.User.Uuid()

	outgoingPayloadBytes, err := json.Marshal(payload)
	if err != nil {
		e.Client.Request.Log("Failed to marshal WebRTC payload:", err)
		return
	}

	outgoingEvent := []byte(`{"type":"webrtc_signal","payload":` + string(outgoingPayloadBytes) + `}`)
	p.Channel.Log("webrtc", payload)
	p.Channel.Broadcast(outgoingEvent)
}
