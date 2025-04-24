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
	PathType string
}

func NewFileService(pathType string) *FileService {
	var fs = new(FileService)
	fs.PathType = pathType
	return fs
}

// UploadToOSS 从客户端上传到OSS
func (f *FileService) UploadToOSS(header *multipart.FileHeader, req model.Files, isEnc bool) (err error, outFile model.Files) {
	var oss upload.Cloud
	if isEnc {
		oss = upload.NewUpload(upload.TypeAliyunOssEnc)
	} else {
		oss = upload.NewUpload(upload.TypeAliyunOss)
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
	newFileName = f.prevPathType(newFileName)
	key, err := oss.Upload(file, newFileName)
	filePath := setting.AliyunOSS.BucketUrl + "/" + key
	if req.UserSpace != "" {
		filePath = setting.AliyunOSS.BucketUrl + upload.RoutePathUser + "/" + req.UserSpace + "/" + key
	}
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
	if req.FileId > 0 {
		var exist model.FilesTemp
		if key != "" {
			model.NewFileTemp().Where("`key` = ?", key).One(&exist)
		}
		var temp model.FilesTemp
		sql := model.NewFileTemp()
		temp.Id = exist.Id
		temp.FileId = req.FileId
		temp.Url = filePath
		temp.Key = key
		sql.Request(&temp)
		err := sql.AddOrUpdate()
		req.TempExist = true
		req.Url = filePath
		req.Key = key
		return err, req
	} else {
		filePath = strings.ReplaceAll(filePath, req.Domain, "{DOMAIN}")
		newFile := model.Files{
			CloudType: consts.CloudTypeAliyun,
			FileType:  req.FileType,
			TypeId:    req.TypeId,
			From:      req.From,
			Uid:       req.Uid,
			Domain:    req.Domain,
			UserSpace: req.UserSpace,
			Url:       filePath,
			Name:      header.Filename,
			Tag:       req.Tag,
			Key:       key,
			Other:     req.Other,
		}
		var exist model.Files
		if key != "" {
			model.NewFile().Where("`key` = ?", key).One(&exist)
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
	filePath := setting.AliyunOSS.BucketUrl + "/" + key
	if err != nil {
		return err, model.Files{}
	}
	newFile := model.Files{
		CloudType: consts.CloudTypeAliyun,
		FileType:  req.FileType,
		TypeId:    req.TypeId,
		From:      req.From,
		Uid:       req.Uid,
		Url:       filePath,
		Name:      newFileName,
		Tag:       req.Tag,
		Key:       newFileName,
	}
	var exist model.Files
	if filePath != "" {
		model.NewFile().Where("url = ?", filePath).One(&exist)
	}
	sql := model.NewFile()
	newFile.Id = exist.Id
	sql.Request(&newFile)
	err = sql.AddOrUpdate()
	return err, newFile
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
	newFileName = f.prevPathType(newFileName)
	newFileName, err = local.Upload(file, newFileName)
	filePath := setting.App.APIURL + upload.RoutePath + "/" + newFileName
	if req.UserSpace != "" {
		filePath = setting.App.APIURL + upload.RoutePathUser + "/" + req.UserSpace + "/" + newFileName
	}
	if err != nil {
		return err, model.Files{}
	}
	if saveTemp {
		var exist model.FilesTemp
		if filePath != "" {
			model.NewFileTemp().Where("url = ?", filePath).One(&exist)
		}
		var temp model.FilesTemp
		sql := model.NewFileTemp()
		temp.Id = exist.Id
		temp.FileId = req.FileId
		temp.Url = filePath
		temp.Key = newFileName
		sql.Request(&temp)
		err := sql.AddOrUpdate()
		req.Name = newFileName
		req.TempExist = true
		req.Url = filePath
		req.Id = temp.Id
		return err, req
	} else {
		filePath = strings.ReplaceAll(filePath, req.Domain, "{DOMAIN}")
		newFile := model.Files{
			CloudType: consts.CloudTypeAliyun,
			FileType:  req.FileType,
			TypeId:    req.TypeId,
			Domain:    req.Domain,
			UserSpace: req.UserSpace,
			From:      req.From,
			Uid:       req.Uid,
			Url:       filePath,
			Name:      head.Filename,
			Tag:       req.Tag,
			Key:       newFileName,
			Other:     req.Other,
		}
		var exist model.Files
		if filePath != "" {
			model.NewFile().Where("url = ?", filePath).One(&exist)
		}
		sql := model.NewFile()
		newFile.Id = exist.Id
		sql.Request(&newFile)
		err = sql.AddOrUpdate()
		newFile.Url = strings.ReplaceAll(newFile.Url, "{DOMAIN}", req.Domain)
		return err, newFile
	}
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
	if saveSql {
		var exist model.Files
		if filePath != "" {
			model.NewFile().Where("url = ?", filePath).One(&exist)
		}
		sql := model.NewFile()
		newFile.Id = exist.Id
		sql.Request(&newFile)
		err = sql.AddOrUpdate()
	}
	return err, newFile
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
	newFile := model.Files{
		CloudType: consts.CloudTypeIPFS,
		FileType:  req.FileType,
		TypeId:    req.TypeId,
		From:      req.From,
		Uid:       req.Uid,
		Url:       filePath,
		Name:      filepath.Base(newFileName),
		Tag:       req.Tag,
		Key:       key,
	}
	var exist model.Files
	if filePath != "" {
		model.NewFile().Where("url = ?", filePath).One(&exist)
	}
	sql := model.NewFile()
	newFile.Id = exist.Id
	sql.Request(&newFile)
	err = sql.AddOrUpdate()
	return err, newFile
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
	contentType := ""
	if req.FileType == consts.FileTypeJson {
		contentType = "application/json; charset=utf-8"
	} else {
		newFileName = newFileName + path.Ext(localSql.From)
	}
	newFileName = f.prevPathType(newFileName)
	key, err := oss.Upload(file, newFileName, contentType)
	filePath := setting.AliyunOSS.BucketUrl + "/" + newFileName
	if isEnc {
		fileMap, err := oss.GetInfo(key)
		if err != nil {
			return err, model.Files{}
		}
		filePath = fileMap["url"]
	}
	if err != nil {
		return err, model.Files{}
	}
	newFile := model.Files{
		CloudType: consts.CloudTypeAliyun,
		FileType:  req.FileType,
		TypeId:    req.TypeId,
		From:      req.From,
		Uid:       req.Uid,
		Url:       filePath,
		Name:      filepath.Base(newFileName),
		Tag:       req.Tag,
		Key:       key,
	}
	var exist model.Files
	if filePath != "" {
		model.NewFile().Where("url = ?", filePath).One(&exist)
	}
	sql := model.NewFile()
	newFile.Id = exist.Id
	sql.Request(&newFile)
	err = sql.AddOrUpdate()
	return err, newFile
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
	newFile := model.Files{
		CloudType: consts.CloudTypeWxOa,
		FileType:  req.FileType,
		TypeId:    req.TypeId,
		From:      req.From,
		Uid:       req.Uid,
		Url:       res,
		Name:      filepath.Base(newFileName),
		Tag:       req.Tag,
		Key:       appid,
	}
	var exist model.Files
	if res != "" {
		model.NewFile().Where("url = ?", res).One(&exist)
	}
	sql := model.NewFile()
	newFile.Id = exist.Id
	sql.Request(&newFile)
	err = sql.AddOrUpdate()
	return err, newFile

}
func (f *FileService) prevPathType(filename string) (newFileName string) {
	pathType := f.PathType
	match, _ := regexp.MatchString(`^[A-Za-z0-9_]+$`, pathType)
	if match && pathType != "" {
		filename = pathType + "/" + filename
	}
	return filename
}
