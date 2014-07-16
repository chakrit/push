package push

type Session struct {
	*Payload
	Devices chan *Device
}

func NewSession(payload *Payload) *Session {
	return &Session{
		Payload: payload,
		Devices: make(chan *Device),
	}
}

func (session *Session) Close() {
	close(session.Devices)
}
