package aliyun

import (
	"encoding/json"
	"github.com/hb1707/ant-godmin/pkg/log"
	"github.com/hb1707/ant-godmin/sdk/upload"
	"github.com/hb1707/ant-godmin/setting"
	"github.com/hb1707/exfun/fun"
	"github.com/hb1707/exfun/fun/curl"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func GetImgColor(imgUrl string) string {
	var Resp struct {
		RGB string `json:"RGB"`
	}
	if fun.Stripos(imgUrl, setting.AliyunOSS.BucketUrl) == -1 {
		return ""
	}
	imgUrlArr := strings.Split(imgUrl, "?")
	if len(imgUrlArr) > 1 {
		imgUrl = imgUrlArr[0]
	}
	get, _, err := curl.GET(imgUrl, map[string]string{
		"x-oss-process": "image/average-hue",
	})
	if err != nil {
		log.Error("获取图片主色调出错：", imgUrl, err)
		return ""
	}
	err = json.Unmarshal(get, &Resp)
	if err != nil {
		log.Error("获取图片主色调JSON出错：", imgUrl, err)
		return ""
	}
	return "#" + Resp.RGB[2:]
}

func RemoteSync(imgUrl string, path string, pre string, newFileName string) string {
	if imgUrl == "" {
		return ""
	}
	if fun.Stripos(imgUrl, setting.AliyunOSS.BucketUrl) == -1 {
		ext := filepath.Ext(imgUrl)
		if newFileName == "" {
			newFileName = pre + fun.MD5(imgUrl)
		} else {
			newFileName = pre + newFileName
		}
		if ext != "." && ext != "" {
			newFileName = newFileName + ext
		}
		localPath, err := upload.NewUpload(upload.TypeLocal).Download(imgUrl, "tmp/"+newFileName)
		if err != nil {
			return imgUrl
		}
		file, err := os.Open(localPath)
		if err != nil {
			return imgUrl
		}
		defer file.Close()
		ext2 := filepath.Ext(localPath)
		if ext2 != "." && ext2 != "" && (ext == "" || ext == ".") {
			newFileName = newFileName + ext2
		}
		//上传文件
		ossPath, err := upload.NewUpload(upload.TypeAliyunOss).Upload(file, path+"/"+newFileName)
		if err != nil {
			return imgUrl
		}
		imgUrl = setting.AliyunOSS.BucketUrl + "/" + ossPath
	}
	return imgUrl
}

func XOssProcess(url string, wh string) string {
	if url == "" {
		return ""
	}
	if fun.Stripos(url, setting.AliyunOSS.BucketUrl) == -1 || fun.Stripos(url, "?") > -1 {
		return url
	}
	ext := strings.ToLower(filepath.Ext(url))
	if ext == ".gif" {
		wh = wh + "g"
	}
	if ext == ".mp4" || ext == ".mov" {
		url = url + "?x-oss-process=video/snapshot,t_1000,f_jpg,w_300,h_300,m_fast"
	} else if ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".bmp" || ext == ".webp" || ext == ".gif" {
		url = url + "?x-oss-process=style/" + wh
	}
	return url
}

type ImgInfo struct {
	FileSize int    `json:"fileSize"`
	Format   string `json:"format"`
	Height   int    `json:"height"`
	Width    int    `json:"width"`
}

func GetImgInfo(imgUrl string) *ImgInfo {
	var resp struct {
		FileSize struct {
			Value string `json:"value"`
		} `json:"FileSize"`
		Format struct {
			Value string `json:"value"`
		} `json:"Format"`
		ImageHeight struct {
			Value string `json:"value"`
		} `json:"ImageHeight"`
		ImageWidth struct {
			Value string `json:"value"`
		} `json:"ImageWidth"`
	}
	if fun.Stripos(imgUrl, setting.AliyunOSS.BucketUrl) == -1 {
		return nil
	}
	get, _, err := curl.GET(imgUrl, map[string]string{
		"x-oss-process": "image/info",
	})
	if err != nil {
		log.Error("获取图片信息出错：", imgUrl, err)
		return nil
	}
	err = json.Unmarshal(get, &resp)
	if err != nil {
		log.Error("获取图片信息JSON出错：", imgUrl, err)
		return nil
	}
	var imgInfo ImgInfo
	imgInfo.FileSize, _ = strconv.Atoi(resp.FileSize.Value)
	imgInfo.Format = resp.Format.Value
	imgInfo.Height, _ = strconv.Atoi(resp.ImageHeight.Value)
	imgInfo.Width, _ = strconv.Atoi(resp.ImageWidth.Value)
	return &imgInfo
}
