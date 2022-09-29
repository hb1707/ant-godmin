package upload

import (
	"io"
)

type Cloud interface {
	Upload(file io.Reader, newFileName string, other ...string) (string, error)
	Download(url string, localPath string) (string, error)
	Delete(key string) error
}

const (
	TypeLocal        = "local"
	TypeAliyunOss    = "aliyun_oss"
	TypeAliyunOssEnc = "aliyun_oss_enc"
	TypeIpfs         = "ipfs"
)

func NewUpload(cloudType string) Cloud {
	switch cloudType {
	case TypeAliyunOss:
		return &AliyunOSS{}
	case TypeAliyunOssEnc:
		return &AliyunOSSEnc{}
	case TypeIpfs:
		return &IPFS{}
	default:
		return &Local{}
	}
}
