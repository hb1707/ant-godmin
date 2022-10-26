package service

import (
	"github.com/hb1707/ant-godmin/consts"
	"github.com/hb1707/ant-godmin/model"
	"github.com/hb1707/ant-godmin/sdk/upload"
	"github.com/hb1707/ant-godmin/setting"
	"github.com/hb1707/exfun/fun"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type FileService struct {
	PathType string
}

func NewFileService(pathType string) *FileService {
	var fs = new(FileService)
	fs.PathType = pathType
	return fs
}
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
	newFileName = f.prevPathType(newFileName)
	key, err := oss.Upload(file, newFileName)
	filePath := setting.AliyunOSS.BucketUrl + "/" + key
	if err != nil {
		return err, model.Files{}
	}
	if req.FileId > 0 {
		var temp model.FilesTemp
		sql := model.NewFileTemp()
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
		newFile := model.Files{
			CloudType: consts.CloudTypeAliyun,
			FileType:  req.FileType,
			TypeId:    req.TypeId,
			From:      req.From,
			Uid:       req.Uid,
			Url:       filePath,
			Name:      header.Filename,
			Tag:       req.Tag,
			Key:       key,
		}
		sql := model.NewFile()
		sql.Request(&newFile)
		err := sql.AddOrUpdate()
		return err, newFile
	}
}
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
	sql := model.NewFile()
	sql.Request(&newFile)
	err = sql.AddOrUpdate()
	return err, newFile
}

func (f *FileService) UploadLocal(head *multipart.FileHeader, req model.Files, saveTemp bool) (err error, outFile model.Files) {
	local := upload.NewUpload(upload.TypeLocal)
	newFileName := req.Name
	if newFileName == "" {
		ext := path.Ext(head.Filename)
		name := strings.TrimSuffix(head.Filename, ext)
		name = fun.MD5(name)
		newFileName = name + "_" + time.Now().Format("20060102150405") + ext
	}
	file, err := head.Open()
	if err != nil {
		return err, model.Files{}
	}
	defer file.Close()
	newFileName = f.prevPathType(newFileName)
	newFileName, err = local.Upload(file, newFileName)
	filePath := setting.App.APIURL + "/upload/" + newFileName
	if err != nil {
		return err, model.Files{}
	}
	if saveTemp {
		var temp model.FilesTemp
		sql := model.NewFileTemp()
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
		newFile := model.Files{
			CloudType: consts.CloudTypeAliyun,
			FileType:  req.FileType,
			TypeId:    req.TypeId,
			From:      req.From,
			Uid:       req.Uid,
			Url:       filePath,
			Name:      newFileName,
			Tag:       req.Tag,
			Key:       "",
		}
		sql := model.NewFile()
		sql.Request(&newFile)
		err = sql.AddOrUpdate()
		return err, newFile
	}
}
func (f *FileService) DownloadFile(req model.Files, saveSql bool) (err error, file model.Files) {
	local := upload.NewUpload(upload.TypeLocal)
	newFileName := req.Name
	newFileName = f.prevPathType(newFileName)
	filePath, err := local.Download(req.From, newFileName)
	if err != nil {
		return err, model.Files{}
	}
	newFile := model.Files{
		CloudType: consts.CloudTypeLocal,
		FileType:  req.FileType,
		TypeId:    req.TypeId,
		From:      req.From,
		Uid:       req.Uid,
		Url:       filePath,
		Name:      newFileName,
		Tag:       req.Tag,
		Key:       req.Key,
	}
	if saveSql {
		sql := model.NewFile()
		sql.Request(&newFile)
		err = sql.AddOrUpdate()
	}
	return err, newFile
}
func (f *FileService) IPFSAdd(req model.Files) (err error, outFile model.Files) {
	ipfs := upload.NewUpload(upload.TypeIpfs)
	var fSql model.Files
	model.NewFile("url = ?", req.From).One(&fSql, "created_at desc")
	localPath := setting.Upload.LocalPath + "/" + fSql.Name
	file, err := os.Open(localPath)
	if err != nil {
		return err, model.Files{}
	}
	defer file.Close()

	newFileName := req.Name
	newFileName = f.prevPathType(newFileName)
	key, err := ipfs.Upload(file, newFileName)
	filePath := setting.Upload.IpfsGateway + "/" + key
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
	sql := model.NewFile()
	sql.Request(&newFile)
	err = sql.AddOrUpdate()
	return err, newFile
}
func (f *FileService) OSSAdd(req model.Files, isEnc bool) (err error, outFile model.Files) {
	var oss upload.Cloud
	if isEnc {
		oss = upload.NewUpload(upload.TypeAliyunOssEnc)
	} else {
		oss = upload.NewUpload(upload.TypeAliyunOss)
	}
	var localSql model.Files
	model.NewFile("url = ?", req.From).One(&localSql, "created_at desc")
	localPath := setting.Upload.LocalPath + "/" + localSql.Name
	file, err := os.Open(localPath)
	if err != nil {
		return err, model.Files{}
	}
	defer file.Close()

	newFileName := req.Name
	contentType := ""
	if req.FileType > consts.FileTypeOther {
		newFileName = newFileName + path.Ext(localSql.From)
	} else {
		contentType = "application/json; charset=utf-8"
	}
	newFileName = f.prevPathType(newFileName)
	key, err := oss.Upload(file, newFileName, contentType)
	filePath := setting.AliyunOSS.BucketUrl + "/" + newFileName
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
	sql := model.NewFile()
	sql.Request(&newFile)
	err = sql.AddOrUpdate()
	return err, newFile
}
func (f *FileService) prevPathType(filename string) (newFileName string) {
	pathType := f.PathType
	match, _ := regexp.MatchString(`^[A-Za-z0-9]+$`, pathType)
	if match && pathType != "" {
		filename = pathType + "/" + filename
	}
	return filename
}
