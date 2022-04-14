package upload

import "mime/multipart"

type Cloud interface {
	Upload(file *multipart.FileHeader, pathType string, newFileName string) (string, string, error)
	Delete(key string) error
}

const (
	AliyunOss = "aliyun_oss"
)

func NewUpload(cloudType string) Cloud {
	switch cloudType {
	case AliyunOss:
		return &AliyunOSS{}
	default:
		return &Local{}
	}
}
