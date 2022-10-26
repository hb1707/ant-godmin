package upload

import (
	"errors"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/hb1707/ant-godmin/setting"
	"io"
)

type AliyunOSS struct{}

func (*AliyunOSS) Upload(file io.Reader, newFileName string, other ...string) (string, error) {
	bucket, err := NewBucket()
	if err != nil {
		return "", errors.New("AliyunOSS.Upload().NewBucket Error:" + err.Error())
	}
	ossPath := setting.AliyunOSS.BasePath + newFileName
	if len(other) > 0 {
		err = bucket.PutObject(ossPath, file, oss.ContentType(other[0]))
	} else {
		err = bucket.PutObject(ossPath, file)
	}
	if err != nil {
		return "", errors.New("AliyunOSS.Upload().bucket.PutObject() Error:" + err.Error())
	}

	return ossPath, nil
}
func (*AliyunOSS) Copy(ori string, new string) error {
	bucket, err := NewBucket()
	if err != nil {
		return errors.New("AliyunOSS.Copy().NewBucket() Error:" + err.Error())
	}
	_, err = bucket.CopyObject(new, ori)
	if err != nil {
		return errors.New("AliyunOSS.Copy().CopyObject() Error:" + err.Error())
	}
	return nil
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
