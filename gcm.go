package push

import "fmt"
import "github.com/alexjlockwood/gcm"

type GCM struct {
	ApiKey    string
	BatchSize int
}

var _ Service = &GCM{}

func (g *GCM) Accepts() []DeviceType {
	return []DeviceType{DeviceTypeAndroid}
}

func (g *GCM) Start(io *IO) (e error) {
	defer func() {
		var ok bool
		if r := recover(); r != nil {
			e, ok = r.(error)
			if !ok {
				e = fmt.Errorf("%s", r)
			}
		}
	}()

	sender := &gcm.Sender{ApiKey: g.ApiKey}
	go g.pump(sender, io)

	return nil
}

func (g *GCM) pump(sender *gcm.Sender, io *IO) {
	for session := range io.Input {
		batches := g.batchDevices(session.Devices)

		for batch := range batches {
			fmt.Println("BATCH")
			tokens := make([]string, len(batch))
			for i, device := range batch {
				tokens[i] = device.Token
			}

			payload := map[string]interface{}{
				"Title":   session.Payload.Title,
				"Content": session.Payload.Description,
			}

			for key, value := range session.Payload.Data {
				payload[key] = value
			}

			msg := gcm.NewMessage(payload, tokens...)
			resp, err := sender.Send(msg, 1)
			if err != nil {
				for _, device := range batch {
					io.Output <- &Feedback{device, "", err}
				}
				continue
			}

			// Correlating GCM results.
			for i, result := range resp.Results {
				device := batch[i]
				feedback := &Feedback{Device: device}

				switch {
				case result.Error != "":
					feedback.Error = fmt.Errorf("gcm: %s", result.Error)
				case result.RegistrationID != "" && result.RegistrationID != device.Token:
					feedback.NewToken = result.RegistrationID
				default: // no-op
				}

				io.Output <- feedback
			}
		} // batch
	} // session
}

func (g *GCM) batchDevices(devices chan *Device) chan []*Device {
	buffer := make([]*Device, 0, g.BatchSize)
	output := make(chan []*Device)

	flush := func() {
		if len(buffer) > 0 {
			output <- buffer
			buffer = make([]*Device, 0, cap(buffer))
		}
	}

	go func() {
		for device := range devices {
			buffer = append(buffer, device)
			if len(buffer) >= cap(buffer) {
				flush()
			}
		}

		flush()
		close(output)
	}()

	return output
}
