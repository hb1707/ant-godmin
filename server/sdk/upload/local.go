package upload

import (
	"errors"
	"github.com/hb1707/ant-godmin/setting"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type Local struct{}

func (*Local) Upload(file io.Reader, newFileName string) (string, error) {
	pathNew := setting.Upload.LocalPath + "/" + newFileName
	out, err := os.Create(pathNew)
	if err != nil {
		return "", errors.New("Local.Upload().os.Create() Error:" + err.Error())
	}
	defer out.Close()
	_, err = io.Copy(out, file)
	if err != nil {
		return "", errors.New("Local.Upload().io.Copy() Error:" + err.Error())
	}
	return pathNew, nil
}
func (*Local) Download(url string, localFileName string) (string, error) {
	if localFileName == "" {
		localFileName = time.Now().Format("20060102150405")
	}
	err := os.MkdirAll(setting.Upload.LocalPath, os.ModePerm)
	if err != nil {
		return "", errors.New("Local.Download().os.MkdirAll() Error:" + err.Error())
	}
	pathNew := setting.Upload.LocalPath + "/" + localFileName
	out, err := os.Create(pathNew)
	if err != nil {
		return "", errors.New("Local.Download().os.Create() Error:" + err.Error())
	}
	defer out.Close()
	res, _ := http.Get(url)
	defer res.Body.Close()
	//f := io.Reader(res.Body)
	_, err = io.Copy(out, res.Body)
	if err != nil {
		return "", errors.New("Local.Download().io.Copy() Error:" + err.Error())
	}
	return pathNew, nil
}
func (*Local) Delete(key string) error {
	filePath := setting.Upload.LocalPath + "/" + key
	if strings.Contains(filePath, setting.Upload.LocalPath) {
		if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
			return errors.New("Local.Delete().Remove() Error:" + err.Error())
		}
	}
	return nil
}
