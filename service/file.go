package service

import (
	"errors"
	"fmt"
	"html"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/hb1707/ant-godmin/consts"
	"github.com/hb1707/ant-godmin/model"
	"github.com/hb1707/ant-godmin/sdk/upload"
	"github.com/hb1707/ant-godmin/sdk/wx"
	"github.com/hb1707/ant-godmin/setting"
	"github.com/hb1707/exfun/fun"
)

type FileService struct {
	PathType           string
	LocalOutputUrl     string
	CloudOutputUrl     string
	CloudOutputUserUrl string
}

func NewFileService(pathType string) *FileService {
	var fs = new(FileService)
	fs.PathType = pathType
	fs.LocalOutputUrl = setting.App.APIURL
	fs.CloudOutputUrl = setting.AliyunOSS.BucketUrl
	fs.CloudOutputUserUrl = setting.AliyunOSS.BucketUrlUser
	return fs
}

func (f *FileService) SaveSql(req model.Files, key string, originalName string) (error, model.Files) {
	if req.FileType == consts.FileTypeOther {
		req.FileType = Ext2FileType(req.Url)
	}
	fileUrl := req.Url
	if req.FileId > 0 {
		var exist model.FilesTemp
		if fileUrl != "" {
			model.NewFileTemp().Where("`url` = ?", fileUrl).One(&exist)
		}
		var temp model.FilesTemp
		sql := model.NewFileTemp()
		sql.Id = exist.Id
		temp.FileId = req.FileId
		temp.Url = fileUrl
		temp.Key = key
		sql.Request(&temp)
		err := sql.AddOrUpdate()
		req.TempExist = true
		req.Url = fileUrl
		req.Path = req.Path
		req.Key = key
		return err, req
	} else {
		if req.Domain == "" {
			// 从url中提取domain，注意要带上https://
			if strings.Contains(fileUrl, "://") {
				domainArr := strings.Split(fileUrl, "/")
				if len(domainArr) > 2 {
					req.Domain = domainArr[0] + "//" + domainArr[2]
				}
			}
		}
		if req.Domain != "" {
			fileUrl = strings.ReplaceAll(fileUrl, req.Domain, "{DOMAIN}")
		}
		newFile := model.Files{
			UUID:      req.UUID,
			CloudType: req.CloudType,
			FileType:  req.FileType,
			TypeId:    req.TypeId,
			From:      req.From,
			Uid:       req.Uid,
			Domain:    req.Domain,
			UserSpace: req.UserSpace,
			Url:       fileUrl,
			Path:      req.Path,
			Name:      filepath.Base(originalName),
			Tag:       req.Tag,
			Key:       key,
			Other:     req.Other,
		}
		var exist model.Files
		if fileUrl != "" {
			model.NewFile().Where("url = ?", fileUrl).One(&exist)
			newFile.Id = exist.Id
		}
		sql := model.NewFile()
		sql.Id = exist.Id
		sql.Request(&newFile)
		err := sql.AddOrUpdate()
		newFile.Url = strings.ReplaceAll(newFile.Url, "{DOMAIN}", req.Domain)
		return err, newFile
	}
}

// UploadToOSS 从客户端上传到OSS
func (f *FileService) UploadToOSS(header *multipart.FileHeader, req model.Files, isEnc bool) (err error, outFile model.Files) {
	var oss upload.Cloud
	if isEnc {
		oss = upload.NewUpload(upload.TypeAliyunOssEnc)
		if req.UserSpace != "" {
			oss.SetBucket(setting.AliyunOSSEnc.BucketNameUser)
		}
	} else {
		oss = upload.NewUpload(upload.TypeAliyunOss)
		if req.UserSpace != "" {
			oss.SetBucket(setting.AliyunOSS.BucketNameUser)
		}
	}
	newFileName := req.Name
	file, err := header.Open()
	if err != nil {
		return err, model.Files{}
	}
	defer file.Close()
	if newFileName == "" {
		newFileName = header.Filename
	}
	size := header.Size
	if req.UserSpace != "" {
		newFileName = req.UserSpace + "/" + newFileName
	}
	newFileName = f.prevPathType(newFileName)
	key, err := oss.Upload(file, newFileName)
	if err != nil {
		return err, model.Files{}
	}
	if req.FileType == consts.FileTypeImage {
		if req.Other.Width == 0 || req.Other.Height == 0 {
			info, _ := oss.GetInfo(key)
			req.Other.Width, _ = strconv.Atoi(fmt.Sprintf("%v", info["image_width"]))
			req.Other.Height, _ = strconv.Atoi(fmt.Sprintf("%v", info["image_height"]))
			req.Other.Size = int(size)
			req.Other.Ext = path.Ext(newFileName)
		}
	}
	fileUrl := f.CloudOutputUrl + "/" + key
	if req.UserSpace != "" {
		fileUrl = f.LocalOutputUrl + "/" + key
	}
	req.CloudType = consts.CloudTypeAliyun
	req.Url = fileUrl
	err, req = f.SaveSql(req, key, header.Filename)
	req.UrlEnc = oss.GetUrl(key, isEnc || req.UserSpace != "", 3600)
	return err, req
}

// UploadRemote 从远程同步到OSS
func (f *FileService) UploadRemote(req model.Files, isEnc bool) (err error, outFile model.Files) {
	var oss upload.Cloud
	if isEnc {
		oss = upload.NewUpload(upload.TypeAliyunOssEnc)
		if req.UserSpace != "" {
			oss.SetBucket(setting.AliyunOSSEnc.BucketNameUser)
		}
	} else {
		oss = upload.NewUpload(upload.TypeAliyunOss)
		if req.UserSpace != "" {
			oss.SetBucket(setting.AliyunOSS.BucketNameUser)
		}
	}

	fileUrl := req.From
	fileUrlArr := strings.Split(fileUrl, "?")
	ext := filepath.Ext(fileUrl)
	if len(fileUrlArr) > 1 {
		ext = filepath.Ext(fileUrlArr[0])
		if req.FileType == consts.FileTypeOther {
			req.FileType = Ext2FileType(fileUrlArr[0])
		}
	} else {
		if req.FileType == consts.FileTypeOther {
			req.FileType = Ext2FileType(fileUrl)
		}
	}
	// 处理特殊字符
	fileUrl = html.UnescapeString(fileUrl)
	res, err := http.Get(fileUrl)
	if err != nil {
		return err, model.Files{}
	}
	if res == nil || res.StatusCode > 400 {
		return errors.New(fmt.Sprintf("StatusCode: %d", res.StatusCode)), model.Files{}
	}
	// 检查是否是html，json，xml等内容
	if strings.Contains(res.Header.Get("Content-Type"), "text/html") || strings.Contains(res.Header.Get("Content-Type"), "application/xml") {
		return errors.New("不支持的文件格式"), model.Files{}
	}
	nameHead := ""
	if cd := res.Header.Get("Content-Disposition"); cd != "" && ext == "" {
		var extMatch = new([]string)
		fun.PregMatch(`filename=(?:"([^"]+)"|([^;\s]+))`, cd, extMatch)
		if len(*extMatch) > 0 {
			for i := 0; i < len(*extMatch); i++ {
				if i == 0 {
					continue
				}
				if (*extMatch)[i] != "" {
					ext = filepath.Ext((*extMatch)[i])
					if ext != "" {
						nameHead = (*extMatch)[i]
						ext = strings.ToLower(ext)
						break
					}
				}
			}
		}
	}
	if ext == "" && strings.Contains(res.Header.Get("Content-Type"), "/") {
		ext = "." + strings.Split(res.Header.Get("Content-Type"), "/")[1]
		if strings.Contains(ext, ";") {
			ext = strings.Split(ext, ";")[0]
		}
		ext = strings.ToLower(ext)
		if ext == ".jpeg" {
			ext = ".jpg"
		}
		if ext == ".svg+xml" {
			ext = ".svg"
		}
		if ext == ".plain" {
			ext = ".txt"
		}
		if ext == ".x-icon" {
			ext = ".ico"
		}
		if ext == ".vnd.microsoft.icon" {
			ext = ".ico"
		}
		if ext == ".javascript" {
			ext = ".js"
		}
		if ext == ".json" {
			req.FileType = consts.FileTypeJson
		}
	}
	extPath := strings.TrimPrefix(ext, ".")
	if extPath == "" {
		extPath = "other"
	}
	file := io.Reader(res.Body)
	newFileName := req.Name
	if req.Name == "" {
		if nameHead != "" {
			req.Name = nameHead
		} else {
			req.Name = filepath.Base(fileUrl)
		}
		newFileName = fmt.Sprintf("%s%s", fun.MD5(fileUrl), ext)
	}

	if req.UserSpace != "" {
		newFileName = req.UserSpace + "/" + extPath + "/" + newFileName
	}
	newFileName = f.prevPathType(newFileName)
	key, err := oss.Upload(file, newFileName)
	if err != nil {
		return err, model.Files{}
	}

	if req.FileType == consts.FileTypeImage {
		if req.Other.Width == 0 || req.Other.Height == 0 {
			info, _ := oss.GetInfo(key)
			req.Other.Width, _ = strconv.Atoi(fmt.Sprintf("%v", info["image_width"]))
			req.Other.Height, _ = strconv.Atoi(fmt.Sprintf("%v", info["image_height"]))
			req.Other.Size = int(res.ContentLength)
			req.Other.Ext = path.Ext(newFileName)
		}
	}
	fileUrlNew := f.CloudOutputUrl + "/" + key
	if req.UserSpace != "" {
		fileUrlNew = f.LocalOutputUrl + "/" + key
	}
	req.CloudType = consts.CloudTypeAliyun
	req.Url = fileUrlNew
	err, req = f.SaveSql(req, key, req.Name)
	req.UrlEnc = oss.GetUrl(key, isEnc || req.UserSpace != "", 3600)
	return err, req
}

// UploadLocal 客户端上传到服务器本地
func (f *FileService) UploadLocal(head *multipart.FileHeader, req model.Files, saveTemp bool) (err error, outFile model.Files) {
	local := upload.NewUpload(upload.TypeLocal)
	newFileName := req.Name
	if newFileName == "" {
		newFileName = head.Filename
	}
	file, err := head.Open()
	if err != nil {
		return err, model.Files{}
	}
	defer file.Close()
	if req.UserSpace != "" {
		newFileName = req.UserSpace + "/" + newFileName
	}
	newFileName = f.prevPathType(newFileName)
	newFileName, err = local.Upload(file, newFileName)
	if err != nil {
		return err, model.Files{}
	}
	key := upload.RoutePath + "/" + newFileName
	if req.UserSpace != "" {
		key = upload.RoutePathUser + "/" + newFileName
	}
	fileUrl := f.LocalOutputUrl + "/" + key
	req.CloudType = consts.CloudTypeLocal
	req.Url = fileUrl
	return f.SaveSql(req, key, head.Filename)
}

// DownloadFile 下载文件到服务器本地
func (f *FileService) DownloadFile(req model.Files, saveSql bool) (err error, file model.Files) {
	local := upload.NewUpload(upload.TypeLocal)
	newFileName := req.Name
	newFileName = f.prevPathType(newFileName)
	filePath, err := local.Download(req.From, newFileName)
	if err != nil {
		return err, model.Files{}
	}
	newFileNamePath := strings.Replace(filePath, setting.Upload.LocalPath, "", -1)
	newFile := model.Files{
		CloudType: consts.CloudTypeLocal,
		FileType:  req.FileType,
		TypeId:    req.TypeId,
		From:      req.From,
		Uid:       req.Uid,
		Url:       filePath,
		Name:      newFileNamePath,
		Tag:       req.Tag,
		Key:       req.Key,
		Path:      req.Path,
	}
	if !saveSql {
		return nil, newFile
	}
	return f.SaveSql(newFile, req.Key, newFileNamePath)
}

// IPFSAdd 服务器本地同步到IPFS
func (f *FileService) IPFSAdd(req model.Files) (err error, outFile model.Files) {
	ipfs := upload.NewUpload(upload.TypeIpfs)
	var localSql model.Files
	var sq = model.NewFile()
	localPath := ""
	if req.Id > 0 {
		sq.Where("id = ?", req.Id)
		sq.One(&localSql, "created_at desc")
		localPath = setting.Upload.LocalPath + "/" + localSql.Name
	} else {
		localPath = req.From
	}
	if fun.Stripos(localPath, setting.Upload.LocalPath) == -1 {
		return errors.New("非本地文件无法处理：" + localPath), model.Files{}
	}
	file, err := os.Open(localPath)
	if err != nil {
		return err, model.Files{}
	}
	defer file.Close()

	newFileName := req.Name
	newFileName = f.prevPathType(newFileName)
	key, err := ipfs.Upload(file, newFileName)
	filePath := setting.IPFS.IpfsGateway + "/" + key
	if err != nil {
		return err, model.Files{}
	}
	req.CloudType = consts.CloudTypeIPFS
	req.Url = filePath
	return f.SaveSql(req, key, newFileName)
}

// OSSAdd 服务器本地同步到OSS
func (f *FileService) OSSAdd(req model.Files, isEnc bool) (err error, outFile model.Files) {
	var oss upload.Cloud
	if isEnc {
		oss = upload.NewUpload(upload.TypeAliyunOssEnc)
		if req.UserSpace != "" {
			oss.SetBucket(setting.AliyunOSSEnc.BucketNameUser)
		}
	} else {
		oss = upload.NewUpload(upload.TypeAliyunOss)
		if req.UserSpace != "" {
			oss.SetBucket(setting.AliyunOSS.BucketNameUser)
		}
	}
	var localSql model.Files
	var sq = model.NewFile()
	localPath := ""
	if req.Id > 0 {
		sq.Where("id = ?", req.Id)
		sq.One(&localSql, "created_at desc")
		localPath = setting.Upload.LocalPath + "/" + localSql.Key
		if req.UserSpace != "" {
			localPath = setting.Upload.UserPath + "/" + localSql.Key
		}
	} else {
		localPath = req.From
	}
	if fun.Stripos(localPath, setting.Upload.LocalPath) == -1 && fun.Stripos(localPath, setting.Upload.UserPath) == -1 {
		return errors.New("非本地文件无法处理：" + localPath), model.Files{}
	}
	file, err := os.Open(localPath)
	if err != nil {
		return err, model.Files{}
	}
	defer file.Close()
	newFileName := req.Name
	contentType := ""
	if req.FileType == consts.FileTypeJson {
		contentType = "application/json; charset=utf-8"
	} else {
		newFileName = newFileName + path.Ext(localSql.From)
	}
	newFileName = f.prevPathType(newFileName)
	key, err := oss.Upload(file, newFileName, contentType)
	if err != nil {
		return err, model.Files{}
	}
	fileUrl := f.CloudOutputUrl + "/" + newFileName
	if isEnc {
		fileUrl = f.LocalOutputUrl + "/" + newFileName
	}
	req.CloudType = consts.CloudTypeAliyun
	req.Url = fileUrl
	err, req = f.SaveSql(req, key, newFileName)
	req.UrlEnc = oss.GetUrl(key, isEnc || req.UserSpace != "", 3600)
	return err, req
}

// WxAdd 服务器本地同步到微信公众号
func (f *FileService) WxAdd(appid string, req model.Files) (err error, outFile model.Files) {

	var localSql model.Files
	var sq = model.NewFile()
	localPath := ""
	if req.Id > 0 {
		sq.Where("id = ?", req.Id)
		sq.One(&localSql, "created_at desc")
		localPath = setting.Upload.LocalPath + "/" + localSql.Name
	} else {
		localPath = req.From
	}
	if fun.Stripos(localPath, setting.Upload.LocalPath) == -1 {
		return errors.New("非本地文件无法处理：" + localPath), model.Files{}
	}
	file, err := os.Open(localPath)
	if err != nil {
		return err, model.Files{}
	}
	defer file.Close()
	newFileName := req.Name
	newFileName = f.prevPathType(newFileName)
	res, err := wx.WxOaUploadImg(appid, localPath)
	if err != nil {
		return err, model.Files{}

	}

	req.CloudType = consts.CloudTypeWxOa
	req.Url = res
	return f.SaveSql(req, appid, newFileName)
}
func (f *FileService) Clear(req model.Files, isEnc bool) error {
	var oss upload.Cloud
	if isEnc {
		oss = upload.NewUpload(upload.TypeAliyunOssEnc)
		if req.UserSpace != "" {
			oss.SetBucket(setting.AliyunOSSEnc.BucketNameUser)
		}
	} else {
		oss = upload.NewUpload(upload.TypeAliyunOss)
		if req.UserSpace != "" {
			oss.SetBucket(setting.AliyunOSS.BucketNameUser)
		}
	}
	var localSql model.Files
	var sq = model.NewFile()
	fileUrl := req.Url
	if req.Domain == "" {
		// 从url中提取domain，注意要带上https://
		if strings.Contains(fileUrl, "://") {
			domainArr := strings.Split(fileUrl, "/")
			if len(domainArr) > 2 {
				req.Domain = domainArr[0] + "//" + domainArr[2]
			}
		}
	}
	if req.Domain != "" {
		fileUrl = strings.ReplaceAll(fileUrl, req.Domain, "{DOMAIN}")
	}
	sq.Where("url = ? AND cloud_type = ?", fileUrl, req.CloudType)
	sq.One(&localSql, "created_at desc")
	if localSql.Id > 0 {
		sql := model.NewFile()
		return sql.DB.Where("id = ?", localSql.Id).Delete(&sql).Error
	}
	err := oss.Delete(localSql.Key)
	if err != nil {
		return err
	}
	return nil
}

func (f *FileService) prevPathType(filename string) (newFileName string) {
	pathType := f.PathType
	match, _ := regexp.MatchString(`^[A-Za-z0-9_\/]+$`, pathType)
	if match && pathType != "" {
		filename = pathType + "/" + filename
	}
	return filename
}

//ext := filepath.Ext(header.Filename)
//	if ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".gif" || ext == ".bmp" || ext == ".tiff" || ext == ".webp" {
//		req.FileType = consts.FileTypeImage
//	} else if ext == ".mp4" || ext == ".avi" || ext == ".mov" || ext == ".wmv" || ext == ".mkv" || ext == ".flv" || ext == ".webm" {
//		req.FileType = consts.FileTypeVideo
//	} else if ext == ".mp3" || ext == ".wav" || ext == ".ogg" || ext == ".flac" || ext == ".aac" || ext == ".wma" || ext == ".m4a" {
//		req.FileType = consts.FileTypeAudio
//	} else if ext == ".pdf" || ext == ".docx" || ext == ".doc" || ext == ".pptx" || ext == ".ppt" || ext == ".xls" || ext == ".xlsx" || ext == ".txt" || ext == ".csv" {
//		req.FileType = consts.FileTypeDocument
//	} else if ext == ".md" || ext == ".markdown" {
//		req.FileType = consts.FileTypeMarkdown
//	} else {
//		req.FileType = consts.FileTypeOther
//	}

func Ext2FileType(url string) consts.FileType {
	ext := strings.ToLower(filepath.Ext(url))
	switch ext {
	case ".png", ".jpg", ".jpeg", ".gif", ".bmp", ".tiff", ".webp":
		return consts.FileTypeImage
	case ".mp4", ".avi", ".mov", ".wmv", ".mkv", ".flv", ".webm":
		return consts.FileTypeVideo
	case ".mp3", ".wav", ".ogg", ".flac", ".aac", ".wma", ".m4a":
		return consts.FileTypeAudio
	case ".pdf", ".docx", ".doc", ".pptx", ".ppt", ".xls", ".xlsx", ".txt", ".csv":
		return consts.FileTypeDocument
	case ".md", ".markdown":
		return consts.FileTypeMarkdown
	case ".json":
		return consts.FileTypeJson
	case ".zip", ".rar", ".tar", ".gz", ".7z":
		return consts.FileTypeArchive
	default:
		return consts.FileTypeOther
	}
}
