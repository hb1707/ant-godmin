package upload

import (
	"errors"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/hb1707/ant-godmin/setting"
	"io"
	"strconv"
)

type AliyunOSSEnc struct{}

// AllObjects 列举所有文件的信息
func (*AliyunOSSEnc) AllObjects(path string, continuation string) (pathList []map[string]string, next string, err error) {
	bucket, err := NewBucket()
	if err != nil {
		return
	}
	continueToken := continuation
	if continuation == "all" {
		continueToken = ""
	}
	for {
		var lsRes oss.ListObjectsResultV2
		lsRes, err = bucket.ListObjectsV2(oss.ContinuationToken(continueToken), oss.Prefix(path))
		if err != nil {
			return
		}
		// 打印列举结果。默认情况下，一次返回100条记录。
		for _, object := range lsRes.Objects {
			pathList = append(pathList, map[string]string{
				"key":           object.Key,
				"type":          object.Type,
				"size":          strconv.FormatInt(object.Size, 10),
				"etag":          object.ETag,
				"last_modified": object.LastModified.Format("2006-01-02 15:04:05"),
				"storage_class": object.StorageClass,
			})
		}
		if lsRes.IsTruncated {
			continueToken = lsRes.NextContinuationToken
			if continuation != "all" {
				next = continueToken
				break
			}
		} else {
			next = ""
			break
		}
	}
	return
}
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
