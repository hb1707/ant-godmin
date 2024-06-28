package utils

import (
	"fmt"
	"github.com/hb1707/ant-godmin/consts"
	"github.com/hb1707/ant-godmin/model"
	"github.com/hb1707/ant-godmin/pkg/log"
	"github.com/hb1707/ant-godmin/sdk/upload"
	"github.com/hb1707/ant-godmin/service"
	"github.com/hb1707/ant-godmin/setting"
	"github.com/hb1707/exfun/fun"
	"path/filepath"
	"regexp"
	"strings"
)

// MarkdownImage2Local html中的图片转换为本地图片
func MarkdownImage2Local(body string, callback func(thLocal string) string, path string, saveSql bool) (string, []string) {
	srcParams := regexp.MustCompile(`!\[.+\]\((.+?)\)`).FindAllStringSubmatch(body, -1)
	//var urls []string
	var localPath []string
	for _, v := range srcParams {
		if len(v) < 2 {
			continue
		}
		urlRemote := v[1]
		if !strings.HasPrefix(urlRemote, setting.App.APIURL) && (strings.HasPrefix(urlRemote, "http") || strings.HasPrefix(urlRemote, "https")) {
			local, url := RemoteImage2Local(urlRemote, path, saveSql)
			urlNew := callback(local)
			if urlNew != "" {
				url = urlNew
			}
			body = strings.Replace(body, v[1], url, -1)
			//urls = append(urls, url)
			localPath = append(localPath, local)
		}
	}
	return body, localPath
}

// HtmlImage2Local html中的图片转换为本地图片
func HtmlImage2Local(body string, callback func(thLocal string) string, path string, saveSql bool) (string, []string) {
	srcParams := regexp.MustCompile(`<img.+?src="(.+?)".*?>`).FindAllStringSubmatch(body, -1)
	//var urls []string
	var localPath []string
	for _, v := range srcParams {
		if len(v) < 2 {
			continue
		}
		urlRemote := v[1]
		if !strings.HasPrefix(urlRemote, setting.App.APIURL) && (strings.HasPrefix(urlRemote, "http") || strings.HasPrefix(urlRemote, "https")) {
			local, url := RemoteImage2Local(urlRemote, path, saveSql)
			urlNew := callback(local)
			if urlNew != "" {
				url = urlNew
			}
			body = strings.Replace(body, v[1], url, -1)
			//urls = append(urls, url)
			localPath = append(localPath, local)
		}
	}
	return body, localPath
}

// RemoteImage2Local 远程图片转换为本地图片
func RemoteImage2Local(url string, path string, saveSql bool) (string, string) {
	if strings.HasPrefix(url, setting.App.APIURL) {
		return url, strings.Replace(url, fmt.Sprintf("%s%s", setting.App.APIURL, upload.UploadPath), "", -1)
	}
	var up model.Files
	up.FileType = consts.FileTypeImage
	up.From = url
	up.Name = fun.SHA256(url) + filepath.Ext(url)
	err, f := service.NewFileService(path).DownloadFile(up, saveSql) // 文件上传后拿到文件路径
	if err != nil {
		log.Error("远程文件下载失败!", err)
		return "", ""
	}
	return f.Url, fmt.Sprintf("%s%s%s", setting.App.APIURL, upload.UploadPath, f.Name)
}
