package types

type ContentType string

const (
	TextPlain ContentType = "text/plain; charset=utf-8"
	TextHTML  ContentType = "text/html; charset=utf-8"
	AppJSON   ContentType = "application/json"
	AppXML    ContentType = "application/xml"
	AppOctet  ContentType = "application/octet-stream"
	ImagePNG  ContentType = "image/png"
	ImageJPEG ContentType = "image/jpeg"
	ImageGIF  ContentType = "image/gif"
)
