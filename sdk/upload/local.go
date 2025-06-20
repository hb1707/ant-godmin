package upload

import (
	"errors"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var RoutePath = "/upload"
var RoutePathUser = "/udata"

type Local struct {
	SavePath string
}

func (c *Local) SetPath(path string) {
	c.SavePath = path
}
func (c *Local) SetBucket(bucketName string) {
	// c.BucketName = bucketName
}

// AllObjects 列举所有文件的信息
func (c *Local) AllObjects(path string, continuation string) (pathList []map[string]string, next string, err error) {

	return
}

// GetInfo 文件的信息
func (*Local) GetInfo(key string) (info map[string]string, err error) {
	return
}
func (c *Local) Upload(file io.Reader, localFileName string, other ...string) (string, error) {
	localPath := c.SavePath
	localFilePath := strings.Split(localFileName, "/")
	err := os.MkdirAll(localPath+"/"+strings.Join(localFilePath[0:len(localFilePath)-1], "/"), os.ModePerm)
	if err != nil {
		return "", errors.New("Local.Upload().os.MkdirAll() Error:" + err.Error())
	}
	pathNew := localPath + "/" + localFileName
	out, err := os.Create(pathNew)
	if err != nil {
		return "", errors.New("Local.Upload().os.Create() Error:" + err.Error())
	}
	defer out.Close()
	_, err = io.Copy(out, file)
	if err != nil {
		return "", errors.New("Local.Upload().io.Copy() Error:" + err.Error())
	}
	return localFileName, nil
}
func (c *Local) AsyncProcessObject(sourceKey, process string) (map[string]string, error) {
	return nil, nil
}
func (*Local) Copy(ori string, new string) error {
	return nil
}
func (c *Local) Download(url string, localFileName string) (string, error) {
	if localFileName == "" {
		localFileName = time.Now().Format("20060102150405")
	}
	//获取文件目录路径
	localPath := c.SavePath
	localFilePath := strings.Split(localFileName, "/")
	err := os.MkdirAll(localPath+"/"+strings.Join(localFilePath[0:len(localFilePath)-1], "/"), os.ModePerm)
	if err != nil {
		return "", errors.New("Local.Download().os.MkdirAll() Error:" + err.Error())
	}
	pathNew := localPath + "/" + localFileName

	res, err := http.Get(url)
	if err != nil {
		return "", errors.New("Local.Download().http.Get() Error:" + err.Error())
	}
	defer res.Body.Close()

	ext := filepath.Ext(pathNew)
	if ext == "" {
		extArr, err := mime.ExtensionsByType(res.Header.Get("Content-Type"))
		if err != nil {
			return "", err
		}
		if len(extArr) > 0 {
			pathNewExt := pathNew + extArr[0]
			pathNew = pathNewExt
		}
	}
	out, err := os.Create(pathNew)
	if err != nil {
		return "", errors.New("Local.Download().os.Create() Error:" + err.Error())
	}
	defer out.Close()
	//f := io.Reader(res.Body)
	_, err = io.Copy(out, res.Body)
	if err != nil {
		return "", errors.New("Local.Download().io.Copy() Error:" + err.Error())
	}

	return pathNew, nil
}
func (c *Local) Delete(key string) error {
	localPath := c.SavePath
	filePath := localPath + "/" + key
	if strings.Contains(filePath, localPath) {
		if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
			return errors.New("Local.Delete().Remove() Error:" + err.Error())
		}
	}
	return nil
}
