package push

type Client struct {
	services []Service
	accepts  map[DeviceType][]Service
	channels map[Service]*IO

	Feedbacks chan *Feedback
}

func NewClient() *Client {
	return &Client{
		services: []Service{},
		accepts:  map[DeviceType][]Service{},
		channels: map[Service]*IO{},

		Feedbacks: make(chan *Feedback),
	}
}

func (client *Client) Add(service Service) {
	client.services = append(client.services, service)
	client.channels[service] = NewIO()
	for _, t := range service.Accepts() {
		client.accepts[t] = append(client.accepts[t], service)
	}
}

func (client *Client) Start() error {
	for _, service := range client.services {
		e := service.Start(client.channels[service])
		if e != nil {
			return e
		}
	}

	// Ensure we have a clean start before spawning aggregators.
	for _, service := range client.services {
		io := client.channels[service]
		go func() {
			for feedback := range io.Output {
				client.Feedbacks <- feedback
			}
		}()
	}

	return nil
}

func (client *Client) Close() {
	for _, io := range client.channels {
		io.Close()
	}

	close(client.Feedbacks)
	client.services = []Service{}
	client.accepts = nil
	client.channels = nil
}

func (client *Client) Send(payload *Payload) *Session {
	session := NewSession(payload)

	go func() {
		subsessions := map[Service]*Session{}

		for device := range session.Devices {
			for _, service := range client.accepts[device.Type] {
				if subsessions[service] == nil {
					subsession := NewSession(payload)
					subsessions[service] = subsession
					client.channels[service].Input <- subsession
				}

				subsessions[service].Devices <- device
			}
		}

		for _, subsession := range subsessions {
			subsession.Close()
		}
	}()

	return session
}

