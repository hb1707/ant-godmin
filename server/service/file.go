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
func (f *FileService) UploadToOSS(header *multipart.FileHeader, req model.Files) (err error, outFile model.Files) {
	oss := upload.NewUpload(upload.TypeAliyunOss)
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
func (f *FileService) UploadRemote(req model.Files) (err error, outFile model.Files) {
	oss := upload.NewUpload(upload.TypeAliyunOss)
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

func (f *FileService) UploadLocal(head *multipart.FileHeader, req model.Files) (err error, outFile model.Files) {
	local := upload.NewUpload(upload.TypeLocal)
	newFileName := req.Name
	if newFileName == "" {
		ext := path.Ext(head.Filename)
		name := strings.TrimSuffix(head.Filename, ext)
		name = fun.MD5(name)
		newFileName = name + "_" + time.Now().Format("20060102150405") + ext
	}
	err = os.MkdirAll(setting.Upload.LocalPath, os.ModePerm)
	if err != nil {
		return err, model.Files{}
	}
	file, err := head.Open()
	if err != nil {
		return err, model.Files{}
	}
	defer file.Close()
	newFileName = f.prevPathType(newFileName)
	filePath, err := local.Upload(file, newFileName)
	if err != nil {
		return err, model.Files{}
	}
	newFile := model.Files{
		CloudType: consts.CloudTypeAliyun,
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
func (f *FileService) DownloadFile(req model.Files) (err error, file model.Files) {
	local := upload.NewUpload(upload.TypeLocal)
	newFileName := req.Name
	newFileName = f.prevPathType(newFileName)
	filePath, err := local.Download(req.From, newFileName)
	if err != nil {
		return err, model.Files{}
	}
	newFile := model.Files{
		CloudType: consts.CloudTypeLocal,
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
func (f *FileService) IPFSAdd(req model.Files) (err error, outFile model.Files) {
	ipfs := upload.NewUpload(upload.TypeIpfs)
	localPath := req.From
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
		TypeId:    req.TypeId,
		From:      req.From,
		Uid:       req.Uid,
		Url:       filePath,
		Name:      newFileName,
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
