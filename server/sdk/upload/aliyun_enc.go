package upload

import (
	"errors"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/hb1707/ant-godmin/setting"
	"io"
)

type AliyunOSSEnc struct{}

func (*AliyunOSSEnc) Upload(file io.Reader, newFileName string, other ...string) (string, error) {
	bucket, err := NewBucketEnc()
	if err != nil {
		return "", errors.New("AliyunOSSEnc.Upload().NewBucket Error:" + err.Error())
	}
	ossPath := setting.AliyunOSSEnc.BasePath + newFileName
	if len(other) > 0 {
		err = bucket.PutObject(ossPath, file, oss.ContentType(other[0]))
	} else {
		err = bucket.PutObject(ossPath, file)
	}
	if err != nil {
		return "", errors.New("AliyunOSSEnc.Upload().bucket.PutObject() Error:" + err.Error())
	}

	return ossPath, nil
}
func (*AliyunOSSEnc) Copy(ori string, new string) error {
	return errors.New("AliyunOSSEnc.Copy() Not Support")
}
func (*AliyunOSSEnc) Download(url string, localFileName string) (string, error) {
	return "", errors.New("AliyunOSSEnc.Download() Not Support")
}

func (*AliyunOSSEnc) Delete(key string) error {
	bucket, err := NewBucketEnc()
	if err != nil {
		return errors.New("AliyunOSSEnc.Delete().NewBucket() Error:" + err.Error())
	}
	err = bucket.DeleteObject(key)
	if err != nil {
		return errors.New("AliyunOSSEnc.Delete().bucket.DeleteObject() Error:" + err.Error())
	}

	return nil
}

func NewBucketEnc() (*oss.Bucket, error) {
	client, err := oss.New(setting.AliyunOSSEnc.Endpoint, setting.AliyunOSSEnc.AccessKeyId, setting.AliyunOSSEnc.AccessKeySecret)
	if err != nil {
		return nil, err
	}
	bucket, err := client.Bucket(setting.AliyunOSSEnc.BucketName)
	if err != nil {
		return nil, err
	}

	return bucket, nil
}
