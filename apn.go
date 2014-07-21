package push

import "github.com/timehop/apns"

type APN struct {
	Gateway  string
	KeyFile  string
	CertFile string
}

var _ Service = &APN{}

func (apn *APN) Accepts() []DeviceType {
	return []DeviceType{DeviceTypeIOS}
}

func (apn *APN) Start(io *IO) error {
	go apn.inputPump(io)
	return nil
}

func (apn *APN) inputPump(io *IO) {
	for session := range io.Input {
		client, e := apns.NewClientWithFiles(apn.Gateway, apn.CertFile, apn.KeyFile)
		if e != nil {
			// Drain the queue and only report connection error tied to the first device.
			firstDevice := <-session.Devices
			for _ = range session.Devices {
			}

			io.Output <- &Feedback{firstDevice, "", e}
			continue
		}

		payload := apns.NewPayload()
		payload.APS.Alert.Body = session.Payload.Title
		if badge, ok := session.Payload.Data["badge"]; ok {
			payload.APS.Badge = badge.(int)
		}
		if sound, ok := session.Payload.Data["sound"]; ok {
			payload.APS.Sound = sound.(string)
		}
		for key, value := range session.Payload.Data {
			payload.SetCustomValue(key, value)
		}

		for device := range session.Devices {
			pn := apns.NewNotification()
			pn.Payload = payload
			pn.DeviceToken = device.Token
			pn.Priority = apns.PriorityImmediate

			e := client.Send(pn)
			io.Output <- &Feedback{device, "", e}
		}

		// End of session. We did this per-session since the apns library automatically
		// terminates the connection after all feedbacks have been read.
		go apn.feedbackSink(io)
	}
}

func (apn *APN) feedbackSink(io *IO) {
	feedback, e := apns.NewFeedbackWithFiles(apn.Gateway, apn.CertFile, apn.KeyFile)
	if e != nil {
		// TODO: The expectation might be that all feedbacks have a non-nil devices. We should
		//   keep that expectation, which might require some architectural changes or an extra
		//   channel tied to a session.
		io.Output <- &Feedback{nil, "", e}
		return
	}

	for f := range feedback.Receive() {
		info := &Device{f.DeviceToken, DeviceTypeIOS}
		io.Output <- &Feedback{info, "", ErrInvalidDevice}
	}
}
