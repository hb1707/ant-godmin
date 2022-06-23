package upload

import (
	"errors"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/hb1707/ant-godmin/setting"
	"io"
)

type AliyunOSS struct{}

func (*AliyunOSS) Upload(file io.Reader, newFileName string) (string, error) {
	bucket, err := NewBucket()
	if err != nil {
		return "", errors.New("AliyunOSS.Upload().NewBucket Error:" + err.Error())
	}
	ossPath := setting.AliyunOSS.BasePath + newFileName
	err = bucket.PutObject(ossPath, file)
	if err != nil {
		return "", errors.New("AliyunOSS.Upload().bucket.PutObject() Error:" + err.Error())
	}

	return ossPath, nil
}
func (*AliyunOSS) Download(url string, localFileName string) (string, error) {
	return "", errors.New("AliyunOSS.Download() Not Support")
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
