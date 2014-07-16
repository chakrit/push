package push

import "fmt"
import "github.com/anachronistic/apns"

var ErrInvalidAPNConfiguration = fmt.Errorf("Invalid APN configuration.")

type APN struct {
	Gateway  string
	KeyFile  string
	CertFile string
}

var _ Service = &APN{}

func (apn *APN) Accepts() []DeviceType {
	return []DeviceType{DeviceTypeIOS}
}

func (apn *APN) Start(io *IO) (e error) {
	defer func() {
		var ok bool

		if r := recover(); r != nil {
			e, ok = r.(error)
			if !ok {
				e = fmt.Errorf("%s", r)
			}
		}
	}()

	client := apns.NewClient(apn.Gateway, apn.CertFile, apn.KeyFile)
	go apn.inputPump(client, io)
	go apn.feedbackSink(client, io)

	return nil
}

func (apn *APN) inputPump(client *apns.Client, io *IO) {
	for session := range io.Input {
		payload := apns.NewPayload()
		payload.Alert = session.Payload.Title
		if badge, ok := session.Payload.Data["badge"]; ok {
			payload.Badge = badge.(int)
		}
		if sound, ok := session.Payload.Data["sound"]; ok {
			payload.Sound = sound.(string)
		}

		for device := range session.Devices {
			pn := apns.NewPushNotification()
			pn.DeviceToken = device.Token
			pn.AddPayload(payload)
			for key, value := range session.Payload.Data {
				pn.Set(key, value)
			}

			resp := client.Send(pn)
			if resp.Error == nil && resp.AppleResponse != "" {
				resp.Error = fmt.Errorf(resp.AppleResponse)
			}

			io.Output <- &Feedback{device, "", resp.Error}
		}
	}
}

func (apn *APN) feedbackSink(client *apns.Client, io *IO) {
	go client.ListenForFeedback()

	for {
		select {
		case feedback := <-apns.FeedbackChannel:
			info := &Device{feedback.DeviceToken, DeviceTypeIOS}
			io.Output <- &Feedback{info, "", ErrInvalidDevice}

		case <-apns.ShutdownChannel:
			// TODO: Warn?
			return
		}
	}
}
