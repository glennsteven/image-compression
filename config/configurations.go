package config

type RabbitMq struct {
	Username                    string
	Password                    string
	Port                        string
	Host                        string
	Topic                       string
	PathOriginalFile            string
	PathCompressed              string
	SubPathOriginalInvtrypht    string
	SubPathOriginalAdjdmgpht    string
	SubPathCompressionInvtrypht string
	SubPathCompressionAdjdmgpht string
}

type Discord struct {
	Url string
}

type Server struct {
	Name string
}

type ImageSetting struct {
	Quality int
}
