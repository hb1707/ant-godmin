package upload

import (
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/hb1707/ant-godmin/setting"
	"io"
	"net/http"
	"os"
)

type Cloud interface {
	AllObjects(path string, next string) ([]map[string]string, string, error)
	GetInfo(key string) (map[string]string, error)
	SetPath(path string)
	SetBucket(bucketName string)
	Upload(file io.Reader, newFileName string, other ...string) (string, error)
	AsyncProcessObject(sourceKey, process string) (map[string]string, error)
	Copy(ori string, new string) error
	Download(url string, localPath string) (string, error)
	Delete(key string) error
}

type cloudType string

const (
	TypeLocal        cloudType = "local"
	TypeAliyunOss    cloudType = "aliyun_oss"
	TypeAliyunOssEnc cloudType = "aliyun_oss_enc"
	TypeIpfs         cloudType = "ipfs"
)

func NewUpload(cloudType cloudType) Cloud {
	switch cloudType {
	case TypeAliyunOss:
		return &AliyunOSS{
			BasePath:   setting.AliyunOSS.BasePath,
			BucketName: setting.AliyunOSS.BucketName,
		}
	case TypeAliyunOssEnc:
		return &AliyunOSSEnc{
			BasePath:   setting.AliyunOSSEnc.BasePath,
			BucketName: setting.AliyunOSSEnc.BucketName,
		}
	case TypeIpfs:
		return &IPFS{}
	default:
		return &Local{
			SavePath: setting.Upload.LocalPath,
		}
	}
}

// GetFileExt 获取文件类型
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
