package request

type ConsumerRequest struct {
	FileName string `json:"filename"`
	MimeType string `json:"mimetype"`
}
