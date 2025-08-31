package upload

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/hb1707/ant-godmin/setting"
	"github.com/hb1707/exfun/fun/curl"
)

type ObjectInfo struct {
	FileSize struct {
		Value string `json:"value"`
	} `json:"FileSize"`
	Format struct {
		Value string `json:"value"`
	} `json:"Format"`
	FrameCount struct {
		Value string `json:"value"`
	} `json:"FrameCount"`
	ImageHeight struct {
		Value string `json:"value"`
	} `json:"ImageHeight"`
	ImageWidth struct {
		Value string `json:"value"`
	} `json:"ImageWidth"`
	ResolutionUnit struct {
		Value string `json:"value"`
	} `json:"ResolutionUnit"`
	XResolution struct {
		Value string `json:"value"`
	} `json:"XResolution"`
	YResolution struct {
		Value string `json:"value"`
	} `json:"YResolution"`
}

type AliyunOSS struct {
	BasePath   string
	BucketName string
}

func (c *AliyunOSS) SetPath(path string) {
	c.BasePath = path
}
func (c *AliyunOSS) SetBucket(bucketName string) {
	c.BucketName = bucketName
}

// AllObjects 列举所有文件的信息
func (c *AliyunOSS) AllObjects(path string, continuation string) (pathList []map[string]string, next string, err error) {
	bucket, err := NewBucket(c.BucketName)
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

// GetUrl 获取文件的访问地址
func (c *AliyunOSS) GetUrl(key string, isPrivate bool, expire int64) string {
	if expire <= 0 {
		expire = 3600 // 默认过期时间为1小时
	}
	if isPrivate {
		bucket, err := NewBucket(c.BucketName)
		if err != nil {
			return ""
		}
		// 生成一个临时的访问URL，过期时间为1小时
		keyParams := strings.Split(key, "?")
		if len(keyParams) > 1 {
			key = keyParams[0]
			signedURL, err := bucket.SignURL(key, oss.HTTPGet, expire, oss.Process(keyParams[1]))
			if err != nil {
				return ""
			}
			return signedURL
		} else {
			signedURL, err := bucket.SignURL(key, oss.HTTPGet, expire)
			if err != nil {
				return ""
			}
			return signedURL
		}
	}
	return setting.AliyunOSS.BucketUrl + "/" + c.BasePath + key
}

// GetInfo 文件的信息
func (*AliyunOSS) GetInfo(key string) (info map[string]any, err error) {
	var resp ObjectInfo
	body, status, err := curl.GET(setting.AliyunOSS.BucketUrl+"/"+key, map[string]string{
		"x-oss-process": "image/info",
	})
	if err != nil {
		return
	}
	if status != 200 {
		err = errors.New("AliyunOSS.GetInfo().curl.GET() Error:" + string(body))
		return
	}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return
	}
	info = map[string]any{
		"key":             key,
		"size":            resp.FileSize.Value,
		"format":          resp.Format.Value,
		"frame_count":     resp.FrameCount.Value,
		"image_height":    resp.ImageHeight.Value,
		"image_width":     resp.ImageWidth.Value,
		"resolution_unit": resp.ResolutionUnit.Value,
		"x_resolution":    resp.XResolution.Value,
		"y_resolution":    resp.YResolution.Value,
	}
	return
}

func (c *AliyunOSS) Upload(file io.Reader, newFileName string, other ...string) (string, error) {
	bucket, err := NewBucket(c.BucketName)
	if err != nil {
		return "", errors.New("AliyunOSS.Upload().NewBucket Error:" + err.Error())
	}
	ossPath := c.BasePath + newFileName
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

func (c *AliyunOSS) AsyncProcessObject(sourceKey, process string) (map[string]string, error) {
	bucket, err := NewBucket(c.BucketName)
	if err != nil {
		return nil, errors.New("AliyunOSS.Upload().NewBucket Error:" + err.Error())
	}
	result, err := bucket.AsyncProcessObject(sourceKey, process)
	if err != nil {
		return nil, fmt.Errorf("转换失败:%w", err)
	}
	return map[string]string{
		"EventId":   result.EventId,
		"RequestId": result.RequestId,
		"TaskId":    result.TaskId,
	}, nil
}

func (c *AliyunOSS) Copy(ori string, new string) error {
	bucket, err := NewBucket(c.BucketName)
	if err != nil {
		return errors.New("AliyunOSS.Copy().NewBucket() Error:" + err.Error())
	}
	options := []oss.Option{
		oss.MetadataDirective(oss.MetaReplace),
		//oss.Expires(expires),
		//oss.SetTagging(taggingInfo),
		// 指定复制源Object的对象标签到目标 Object。
		// oss.TaggingDirective(oss.TaggingCopy),
		// 指定创建目标Object时的访问权限ACL为私有。
		// oss.ObjectACL(oss.ACLPrivate),
		// 指定KMS托管的用户主密钥，该参数仅在x-oss-server-side-encryption为KMS时有效。
		//oss.ServerSideEncryptionKeyID("9468da86-3509-4f8d-a61e-6eab1eac****"),
		// 指定OSS创建目标Object时使用的服务器端加密算法。
		// oss.ServerSideEncryption("AES256"),
		// 指定复制源Object的元数据到目标Object。
		oss.MetadataDirective(oss.MetaCopy),
		// 指定CopyObject操作时是否覆盖同名目标Object。此处设置为true，表示禁止覆盖同名Object。
		// oss.ForbidOverWrite(true),
		// 如果源Object的ETag值和您提供的ETag相等，则执行拷贝操作，并返回200 OK。
		//oss.CopySourceIfMatch("5B3C1A2E053D763E1B002CC607C5****"),
		// 如果源Object的ETag值和您提供的ETag不相等，则执行拷贝操作，并返回200 OK。
		//oss.CopySourceIfNoneMatch("5B3C1A2E053D763E1B002CC607C5****"),
		// 如果指定的时间早于文件实际修改时间，则正常拷贝文件，并返回200 OK。
		//oss.CopySourceIfModifiedSince(2021-12-09T07:01:56.000Z),
		// 如果指定的时间等于或者晚于文件实际修改时间，则正常拷贝文件，并返回200 OK。
		//oss.CopySourceIfUnmodifiedSince(2021-12-09T07:01:56.000Z),
		// 指定Object的存储类型。此处设置为Standard，表示标准存储类型。
		//oss.StorageClass("Standard"),
	}

	// 使用指定的元信息覆盖源文件的元信息。
	_, err = bucket.CopyObject(ori, new, options...)
	if err != nil {
		return errors.New("AliyunOSS.Copy().CopyObject() Error:" + err.Error())
	}
	return nil
}
func (c *AliyunOSS) Download(url string, localFileName string) (string, error) {
	return "", errors.New("AliyunOSS.Download() Not Support")
}

func (c *AliyunOSS) Delete(key string) error {
	bucket, err := NewBucket(c.BucketName)
	if err != nil {
		return errors.New("AliyunOSS.Delete().NewBucket() Error:" + err.Error())
	}
	err = bucket.DeleteObject(key)
	if err != nil {
		return errors.New("AliyunOSS.Delete().bucket.DeleteObject() Error:" + err.Error())
	}

	return nil
}

func NewBucket(bucketName string) (*oss.Bucket, error) {
	client, err := oss.New(setting.AliyunOSS.Endpoint, setting.AliyunOSS.AccessKeyId, setting.AliyunOSS.AccessKeySecret)
	if err != nil {
		return nil, err
	}
	bucket, err := client.Bucket(bucketName)
	if err != nil {
		return nil, err
	}

	return bucket, nil
}
