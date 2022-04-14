package upload

import (
	"errors"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/hb1707/ant-godmin/setting"
	"mime/multipart"
	"regexp"
	"time"
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
	ossPath := setting.AliyunOSS.BasePath + time.Now().Format("2006-01-02") + "/" + newFileName
	match, _ := regexp.MatchString(`^[A-Za-z]+$`, pathType)
	if match && pathType != "" {
		ossPath = setting.AliyunOSS.BasePath + pathType + "/" + time.Now().Format("2006-01-02") + "/" + newFileName
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
