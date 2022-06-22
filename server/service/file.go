package service

import (
	"github.com/hb1707/ant-godmin/model"
	"github.com/hb1707/ant-godmin/sdk/upload"
	"mime/multipart"
)

type FileService struct {
}

func NewFileService() *FileService {
	return new(FileService)
}
func (f *FileService) UploadFile(header *multipart.FileHeader, pathType string, req model.Files) (err error, file model.Files) {
	oss := upload.NewUpload(upload.AliyunOss)
	newFileName := req.Name
	filePath, key, err := oss.Upload(header, pathType, newFileName)
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
			TypeId: req.TypeId,
			From:   req.From,
			Uid:    req.Uid,
			Url:    filePath,
			Name:   header.Filename,
			Tag:    req.Tag,
			Key:    key,
		}
		sql := model.NewFile()
		sql.Request(&newFile)
		err := sql.AddOrUpdate()
		return err, newFile
	}
}
