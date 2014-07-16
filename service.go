package push

type Service interface{
	Accepts() []DeviceType
	Start(io *IO) error
}
