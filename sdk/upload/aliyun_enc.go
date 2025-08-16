package upload

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/hb1707/ant-godmin/setting"
	"io"
	"strconv"
)

type ObjectInfoEnc struct {
	FileSize       string `json:"FileSize"`
	Format         string `json:"Format"`
	FrameCount     int    `json:"FrameCount"`
	ImageHeight    int    `json:"ImageHeight"`
	ImageWidth     int    `json:"ImageWidth"`
	ResolutionUnit int    `json:"ResolutionUnit"`
	XResolution    int    `json:"XResolution"`
	YResolution    int    `json:"YResolution"`
}

type AliyunOSSEnc struct {
	BasePath   string
	BucketName string
}

func (c *AliyunOSSEnc) SetPath(path string) {
	c.BasePath = path
}
func (c *AliyunOSSEnc) SetBucket(bucketName string) {
	c.BucketName = bucketName
}

// AllObjects 列举所有文件的信息
func (c *AliyunOSSEnc) AllObjects(path string, continuation string) (pathList []map[string]string, next string, err error) {
	bucket, err := NewBucketEnc(c.BucketName)
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
func (c *AliyunOSSEnc) GetUrl(key string, isPrivate bool) string {
	bucket, err := NewBucket(c.BucketName)
	if err != nil {
		return ""
	}
	//exist, err := bucket.IsObjectExist(key)
	//if err != nil || !exist {
	//	return ""
	//}
	// 生成一个临时的访问URL，过期时间为1小时
	signedURL, err := bucket.SignURL(key, oss.HTTPGet, 3600)
	if err != nil {
		return ""
	}
	return signedURL
}

// GetInfo 文件的信息
func (c *AliyunOSSEnc) GetInfo(key string) (info map[string]any, err error) {
	bucket, err := NewBucket(c.BucketName)
	if err != nil {
		return nil, errors.New("AliyunOSSEnc.GetInfo().NewBucket() Error:" + err.Error())
	}
	// 构建图片信息处理指令
	process := "image/info"
	result, err := bucket.GetObject(key, oss.Process(process))
	if err != nil {
		return nil, fmt.Errorf("获取图片信息失败: %w", err)
	}
	defer result.Close()

	// 读取并解析图片信息
	infoB, err := io.ReadAll(result)
	if err != nil {
		return nil, fmt.Errorf("读取图片信息失败: %w", err)
	}
	var resp ObjectInfoEnc
	err = json.Unmarshal(infoB, &resp)
	if err != nil {
		return nil, fmt.Errorf("解析图片信息失败: %w", err)
	}

	info = map[string]any{
		"key":             key,
		"size":            resp.FileSize,
		"format":          resp.Format,
		"frame_count":     resp.FrameCount,
		"image_height":    resp.ImageHeight,
		"image_width":     resp.ImageWidth,
		"resolution_unit": resp.ResolutionUnit,
		"x_resolution":    resp.XResolution,
		"y_resolution":    resp.YResolution,
	}
	return
}
func (c *AliyunOSSEnc) Upload(file io.Reader, newFileName string, other ...string) (string, error) {
	bucket, err := NewBucketEnc(c.BucketName)
	if err != nil {
		return "", errors.New("AliyunOSSEnc.Upload().NewBucket Error:" + err.Error())
	}
	ossPath := c.BasePath + newFileName
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

func (c *AliyunOSSEnc) AsyncProcessObject(sourceKey, process string) (map[string]string, error) {
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
func (*AliyunOSSEnc) Copy(ori string, new string) error {
	return errors.New("AliyunOSSEnc.Copy() Not Support")
}
func (*AliyunOSSEnc) Download(url string, localFileName string) (string, error) {
	return "", errors.New("AliyunOSSEnc.Download() Not Support")
}

func (c *AliyunOSSEnc) Delete(key string) error {
	bucket, err := NewBucketEnc(c.BucketName)
	if err != nil {
		return errors.New("AliyunOSSEnc.Delete().NewBucket() Error:" + err.Error())
	}
	err = bucket.DeleteObject(key)
	if err != nil {
		return errors.New("AliyunOSSEnc.Delete().bucket.DeleteObject() Error:" + err.Error())
	}

	return nil
}

func NewBucketEnc(bucketName string) (*oss.Bucket, error) {
	client, err := oss.New(setting.AliyunOSSEnc.Endpoint, setting.AliyunOSSEnc.AccessKeyId, setting.AliyunOSSEnc.AccessKeySecret)
	if err != nil {
		return nil, err
	}
	bucket, err := client.Bucket(bucketName)
	if err != nil {
		return nil, err
	}

	return bucket, nil
}
