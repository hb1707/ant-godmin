package upload

import (
	"errors"
	"github.com/hb1707/ant-godmin/setting"
	"github.com/hb1707/exfun/fun"
	"io"
	"mime/multipart"
	"os"
	"path"
	"regexp"
	"strings"
	"time"
)

type Local struct{}

func (*Local) Upload(file *multipart.FileHeader, pathType string, newFileName string) (string, string, error) {

	if newFileName == "" {
		ext := path.Ext(file.Filename)
		name := strings.TrimSuffix(file.Filename, ext)
		name = fun.MD5(name)
		newFileName = name + "_" + time.Now().Format("20060102150405") + ext
	}
	err := os.MkdirAll(setting.Upload.LocalPath, os.ModePerm)
	if err != nil {
		return "", "", errors.New("Local.Upload().os.MkdirAll() Error:" + err.Error())
	}
	pathNew := setting.Upload.LocalPath + "/" + newFileName
	match, _ := regexp.MatchString(`^[A-Za-z]+$`, pathType)
	if match && pathType != "" {
		pathNew = setting.Upload.LocalPath + "/" + pathType + "/" + newFileName
	}
	f, err := file.Open()
	if err != nil {
		return "", "", errors.New("Local.Upload().file.Open() Error:" + err.Error())
	}
	defer f.Close()
	out, err := os.Create(pathNew)
	if err != nil {
		return "", "", errors.New("Local.Upload().os.Create() Error:" + err.Error())
	}
	defer out.Close()
	_, err = io.Copy(out, f)
	if err != nil {
		return "", "", errors.New("Local.Upload().io.Copy() Error:" + err.Error())
	}
	return pathNew, newFileName, nil
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
