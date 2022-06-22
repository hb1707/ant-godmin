package upload

import (
	"errors"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/hb1707/ant-godmin/setting"
	"io"
	"mime/multipart"
	"net/http"
	"regexp"
)

type AliyunOSS struct{}

func (*AliyunOSS) Upload(file *multipart.FileHeader, pathType string, newFileName string) (string, string, error) {
	bucket, err := NewBucket()
	if err != nil {
		return "", "", errors.New("AliyunOSS.Upload().NewBucket Error:" + err.Error())
	}
	f, err := file.Open()
	if err != nil {
		return "", "", errors.New("AliyunOSS.Upload().NewBucket.file.Open() Error:" + err.Error())
	}
	defer f.Close()
	if newFileName == "" {
		newFileName = file.Filename
	}
	ossPath := setting.AliyunOSS.BasePath + newFileName
	match, _ := regexp.MatchString(`^[A-Za-z0-9]+$`, pathType)
	if match && pathType != "" {
		ossPath = setting.AliyunOSS.BasePath + pathType + "/" + newFileName
	}
	err = bucket.PutObject(ossPath, f)
	if err != nil {
		return "", "", errors.New("AliyunOSS.Upload().bucket.PutObject() Error:" + err.Error())
	}

	return setting.AliyunOSS.BucketUrl + "/" + ossPath, ossPath, nil
}
func (*AliyunOSS) Download(url string, pathType string, newFileName string) (string, string, error) {
	bucket, err := NewBucket()
	if err != nil {
		return "", "", errors.New("AliyunOSS.Download().NewBucket Error:" + err.Error())
	}
	res, _ := http.Get(url)
	f := io.Reader(res.Body)
	ossPath := setting.AliyunOSS.BasePath + newFileName
	match, _ := regexp.MatchString(`^[A-Za-z0-9]+$`, pathType)
	if match && pathType != "" {
		ossPath = setting.AliyunOSS.BasePath + pathType + "/" + newFileName
	}
	err = bucket.PutObject(ossPath, f)
	if err != nil {
		return "", "", errors.New("AliyunOSS.Upload().bucket.PutObject() Error:" + err.Error())
	}
	return setting.AliyunOSS.BucketUrl + "/" + ossPath, ossPath, nil
}

func (*AliyunOSS) Delete(key string) error {
	bucket, err := NewBucket()
	if err != nil {
		return errors.New("AliyunOSS.Delete().NewBucket() Error:" + err.Error())
	}
	err = bucket.DeleteObject(key)
	if err != nil {
		return errors.New("AliyunOSS.Delete().bucket.DeleteObject() Error:" + err.Error())
	}

	return nil
}

func NewBucket() (*oss.Bucket, error) {
	client, err := oss.New(setting.AliyunOSS.Endpoint, setting.AliyunOSS.AccessKeyId, setting.AliyunOSS.AccessKeySecret)
	if err != nil {
		return nil, err
	}
	bucket, err := client.Bucket(setting.AliyunOSS.BucketName)
	if err != nil {
		return nil, err
	}

	return bucket, nil
}
