package upload

import (
	"io"
	"net/http"
	"os"
)

type Cloud interface {
	Upload(file io.Reader, newFileName string, other ...string) (string, error)
	Copy(ori string, new string) error
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

//GetFileExt 获取文件类型
func GetFileExt(f *os.File) string {
	var buf [512]byte
	n, _ := f.Read(buf[:])
	_, err := f.Seek(0, 0)
	if err != nil {
		return ""
	}
	meta := http.DetectContentType(buf[:n])
	switch meta {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	case "image/svg+xml":
		return ".svg"
	case "image/bmp":
		return ".bmp"
	case "video/mp4":
		return ".mp4"
	case "video/avi":
		return ".avi"
	case "video/mpeg":
		return ".mpeg"
	case "video/quicktime":
		return ".mov"
	case "video/x-ms-wmv":
		return ".wmv"
	case "sound/mp3":
		return ".mp3"
	case "sound/wav":
		return ".wav"
	case "sound/ogg":
		return ".ogg"
	case "application/pdf":
		return ".pdf"
	case "application/zip":
		return ".zip"
	case "application/x-rar-compressed":
		return ".rar"
	case "application/x-7z-compressed":
		return ".7z"
	case "application/x-tar":
		return ".tar"
	case "object/txt":
		return ".txt"
	case "object/doc":
		return ".doc"
	case "object/docx":
		return ".docx"
	case "object/xls":
		return ".xls"
	case "object/xlsx":
		return ".xlsx"
	case "object/ppt":
		return ".ppt"
	case "object/pptx":
		return ".pptx"
	case "text/javascript":
		return ".js"
	case "text/css":
		return ".css"
	case "text/html":
		return ".html"
	default:
		return ""
	}
}
