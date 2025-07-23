package service

import (
	"errors"
	"github.com/hb1707/ant-godmin/consts"
	"github.com/hb1707/ant-godmin/model"
	"github.com/hb1707/ant-godmin/sdk/upload"
	"github.com/hb1707/ant-godmin/sdk/wx"
	"github.com/hb1707/ant-godmin/setting"
	"github.com/hb1707/exfun/fun"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
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
			req.Other.Width, _ = strconv.Atoi(info["image_width"])
			req.Other.Height, _ = strconv.Atoi(info["image_height"])
			req.Other.Size = int(size)
			req.Other.Ext = path.Ext(newFileName)
		}
	}
	fileUrl := f.CloudOutputUrl + "/" + key
	if req.UserSpace != "" {
		fileUrl = f.CloudOutputUserUrl + "/" + key
	}
	req.CloudType = consts.CloudTypeAliyun
	req.Url = fileUrl
	err, req = f.SaveSql(req, key, header.Filename)
	return err, req
}

func (f *FileService) SaveSql(req model.Files, key string, originalName string) (error, model.Files) {
	fileUrl := req.Url
	if req.FileId > 0 {
		var exist model.FilesTemp
		if fileUrl != "" {
			model.NewFileTemp().Where("`url` = ?", fileUrl).One(&exist)
		}
		var temp model.FilesTemp
		sql := model.NewFileTemp()
		temp.Id = exist.Id
		temp.FileId = req.FileId
		temp.Url = fileUrl
		temp.Key = key
		sql.Request(&temp)
		err := sql.AddOrUpdate()
		req.TempExist = true
		req.Url = fileUrl
		req.Key = key
		return err, req
	} else {
		fileUrl = strings.ReplaceAll(fileUrl, req.Domain, "{DOMAIN}")
		newFile := model.Files{
			CloudType: req.CloudType,
			FileType:  req.FileType,
			TypeId:    req.TypeId,
			From:      req.From,
			Uid:       req.Uid,
			Domain:    req.Domain,
			UserSpace: req.UserSpace,
			Url:       fileUrl,
			Name:      filepath.Base(originalName),
			Tag:       req.Tag,
			Key:       key,
			Other:     req.Other,
		}
		var exist model.Files
		if fileUrl != "" {
			model.NewFile().Where("url = ?", fileUrl).One(&exist)
		}
		sql := model.NewFile()
		newFile.Id = exist.Id
		sql.Request(&newFile)
		err := sql.AddOrUpdate()
		newFile.Url = strings.ReplaceAll(newFile.Url, "{DOMAIN}", req.Domain)
		return err, newFile
	}
}

// UploadRemote 从远程同步到OSS
func (f *FileService) UploadRemote(req model.Files, isEnc bool) (err error, outFile model.Files) {
	var oss upload.Cloud
	if isEnc {
		oss = upload.NewUpload(upload.TypeAliyunOssEnc)
	} else {
		oss = upload.NewUpload(upload.TypeAliyunOss)
	}
	newFileName := req.Name
	res, _ := http.Get(req.From)
	file := io.Reader(res.Body)
	newFileName = f.prevPathType(newFileName)
	key, err := oss.Upload(file, newFileName)
	if err != nil {
		return err, model.Files{}
	}
	if req.FileType == consts.FileTypeImage {
		if req.Other.Width == 0 || req.Other.Height == 0 {
			info, _ := oss.GetInfo(key)
			req.Other.Width, _ = strconv.Atoi(info["image_width"])
			req.Other.Height, _ = strconv.Atoi(info["image_height"])
			req.Other.Size = int(res.ContentLength)
			req.Other.Ext = path.Ext(newFileName)
		}
	}
	fileUrl := f.CloudOutputUrl + "/" + key
	if req.UserSpace != "" {
		fileUrl = f.CloudOutputUserUrl + "/" + key
	}
	req.CloudType = consts.CloudTypeAliyun
	req.Url = fileUrl
	return f.SaveSql(req, key, req.Name)
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
	} else {
		oss = upload.NewUpload(upload.TypeAliyunOss)
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
		fileMap, err := oss.GetInfo(key)
		if err != nil {
			return err, model.Files{}
		}
		fileUrl = fileMap["url"]
	}
	req.CloudType = consts.CloudTypeAliyun
	req.Url = fileUrl
	return f.SaveSql(req, key, newFileName)
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
func (f *FileService) prevPathType(filename string) (newFileName string) {
	pathType := f.PathType
	match, _ := regexp.MatchString(`^[A-Za-z0-9_]+$`, pathType)
	if match && pathType != "" {
		filename = pathType + "/" + filename
	}
	return filename
}
