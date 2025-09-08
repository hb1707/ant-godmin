package json

import (
	"errors"
	"net/http"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hb1707/ant-godmin/auth"
	"github.com/hb1707/ant-godmin/consts"
	"github.com/hb1707/ant-godmin/model"
	"github.com/hb1707/ant-godmin/pkg/log"
	"github.com/hb1707/ant-godmin/routers/hook"
	"github.com/hb1707/ant-godmin/sdk/upload"
	"github.com/hb1707/ant-godmin/service"
	"github.com/hb1707/ant-godmin/setting"
	"github.com/hb1707/exfun/fun"
	"github.com/xuri/excelize/v2"
)

type UserPathPrams struct {
	Hash     string `uri:"hash" binding:"required"`
	Filepath string `uri:"filepath" binding:"required"`
}

func GetUserFile(c *gin.Context) {
	var params UserPathPrams
	if err := c.ShouldBindUri(&params); err != nil {
		jsonErr(c, http.StatusBadRequest, err)
		return
	}
	hash := params.Hash

	pathStr := params.Filepath
	key := path.Join(setting.Upload.UserPath, hash, pathStr)

	oss := upload.NewUpload(upload.TypeAliyunOss)
	oss.SetBucket(setting.AliyunOSS.BucketNameUser)
	filePath := oss.GetUrl(key, true, 3600)
	response, err := http.Get(filePath)
	if err != nil || response.StatusCode != http.StatusOK {
		c.Status(http.StatusServiceUnavailable)
		return
	}

	reader := response.Body
	contentLength := response.ContentLength
	contentType := response.Header.Get("Content-Type")

	extraHeaders := map[string]string{
		//"Content-Disposition": "image; filename=\"" + name + "\"",
	}

	c.DataFromReader(http.StatusOK, contentLength, contentType, reader, extraHeaders)
	return
}

var (
	ErrUploadFail = errors.New("接收文件失败")
)

func UploadOSS(c *gin.Context) {
	var file model.Files
	var req model.Files
	pathStr := c.Param("path")
	typeId, _ := strconv.Atoi(c.DefaultQuery("type_id", "0"))
	fileType, _ := strconv.Atoi(c.DefaultQuery("file_type", "1"))
	photoId, _ := strconv.Atoi(c.DefaultQuery("photo_id", "0")) //todo: 20220621之后准备废弃
	fileId, _ := strconv.Atoi(c.DefaultQuery("file_id", "0"))
	fileName := c.PostForm("file_name")
	req.TypeId = uint(typeId)
	req.FileType = consts.FileType(fileType)
	uid, _ := auth.Identity(c)
	req.Uid = uint(uid)
	req.From = "T"
	if photoId > 0 {
		req.FileId = uint(photoId)
	} else {
		req.FileId = uint(fileId)
	}
	_, header, err := c.Request.FormFile("file")
	if err != nil {
		log.Error("接收文件失败!", err)
		jsonErr(c, http.StatusInternalServerError, ErrUploadFail)
		return
	}
	hookReq := hook.GenerateFileTag(c)
	req.Tag = hookReq.Tag
	ext := path.Ext(header.Filename)
	if fileName == "" {
		name := strings.TrimSuffix(header.Filename, ext)
		name = fun.MD5(name)
		fileName = time.Now().Format("2006-01-02") + "/" + name + "_" + time.Now().Format("20060102150405")
	}
	req.Name = fileName + ext
	err, file = service.NewFileService(pathStr).UploadToOSS(header, req, false) // 文件上传后拿到文件路径
	if err != nil {
		log.Error("上传文件到OSS失败!", err)
		jsonErr(c, http.StatusInternalServerError, err)
		return
	}
	if file.Id > 0 {
		jsonResult(c, http.StatusOK, map[string]interface{}{
			"file": file,
		})
		return
	} else {
		jsonResult(c, http.StatusOK, map[string]interface{}{})
		return
	}
}
func UploadLocal(c *gin.Context) {
	var file model.Files
	var req model.Files
	pathStr := c.Param("path")
	typeId, _ := strconv.Atoi(c.DefaultQuery("type_id", "0"))
	fileType, _ := strconv.Atoi(c.DefaultQuery("file_type", "1"))
	fileName := c.DefaultQuery("file_name", "")
	req.TypeId = uint(typeId)
	req.FileType = consts.FileType(fileType)
	uid, _ := auth.Identity(c)
	req.Uid = uint(uid)
	_, header, err := c.Request.FormFile("file")
	if err != nil {
		log.Error("接收文件失败!", err)
		jsonErr(c, http.StatusInternalServerError, ErrUploadFail)
		return
	}
	hookReq := hook.GenerateFileTag(c)
	req.Tag = hookReq.Tag
	ext := path.Ext(header.Filename)
	if fileName == "" {
		name := strings.TrimSuffix(header.Filename, ext)
		name = fun.MD5(name)
		fileName = time.Now().Format("2006-01-02") + "/" + name + "_" + time.Now().Format("20060102150405")
	}
	req.Name = fileName + ext
	err, file = service.NewFileService(pathStr).UploadLocal(header, req, false) // 文件上传后拿到文件路径
	if err != nil {
		log.Error("上传文件到服务器失败!", err)
		jsonErr(c, http.StatusInternalServerError, err)
		return
	}
	if file.Id > 0 {
		jsonResult(c, http.StatusOK, map[string]interface{}{
			"file": file,
		})
		return
	} else {
		jsonResult(c, http.StatusOK, map[string]interface{}{})
		return
	}
}
func DownloadFile(c *gin.Context) {
	var file model.Files
	var up model.Files
	var req struct {
		FileName string `json:"file_name"`
		Path     string `json:"path"`
		Url      string `json:"url"`
	}
	err := c.ShouldBindJSON(&req)
	if err != nil {
		jsonErr(c, http.StatusInternalServerError, ErrUploadFail)
		return
	}
	up.FileType = consts.FileTypeJson
	uid, _ := auth.Identity(c)
	up.Uid = uint(uid)
	up.From = req.Url
	hookReq := hook.GenerateFileTag(c)
	up.Tag = hookReq.Tag
	up.Name = req.FileName
	err, file = service.NewFileService(req.Path).DownloadFile(up, true) // 文件上传后拿到文件路径
	if err != nil {
		log.Error("远程文件下载失败!", err)
		jsonErr(c, http.StatusInternalServerError, err)
		return
	}
	if file.Id > 0 {
		jsonResult(c, http.StatusOK, map[string]interface{}{
			"file": file,
		})
		return
	} else {
		jsonResult(c, http.StatusOK, map[string]interface{}{})
		return
	}
}
func UploadSyncIPFS(c *gin.Context) {
	var file model.Files
	var up model.Files
	var req struct {
		FileName  string          `json:"file_name"`
		Path      string          `json:"path"`
		RemoteUrl string          `json:"remote_url"`
		LocalId   uint            `json:"local_id"`
		FileType  consts.FileType `json:"file_type"`
	}
	err := c.ShouldBindJSON(&req)
	if err != nil {
		jsonErr(c, http.StatusInternalServerError, ErrUploadFail)
		return
	}
	up.FileType = req.FileType
	uid, _ := auth.Identity(c)
	up.Uid = uint(uid)
	hookReq := hook.GenerateFileTag(c)
	up.Tag = hookReq.Tag
	up.Name = req.FileName
	if req.RemoteUrl != "" {
		up.From = req.RemoteUrl
		err, fileLocal := service.NewFileService(req.Path).DownloadFile(up, false) // 文件上传后拿到文件路径
		if err != nil {
			log.Error("远程文件下载失败!", err)
			jsonErr(c, http.StatusInternalServerError, err)
			return
		}
		up.From = fileLocal.Url
	} else {
		up.Id = req.LocalId
	}
	err, file = service.NewFileService(req.Path).IPFSAdd(up) //本地文件上传到IPFS
	if err != nil {
		log.Error("本地文件上传到IPFS!", err)
		jsonErr(c, http.StatusInternalServerError, err)
		return
	}
	if file.Id > 0 {
		jsonResult(c, http.StatusOK, map[string]interface{}{
			"file": file,
		})
		return
	} else {
		jsonResult(c, http.StatusOK, map[string]interface{}{})
		return
	}
}
func UploadSyncOSS(c *gin.Context) {
	var file model.Files
	var up model.Files
	var req struct {
		FileName  string          `json:"file_name"`
		Path      string          `json:"path"`
		RemoteUrl string          `json:"remote_url"`
		LocalId   uint            `json:"local_id"`
		FileType  consts.FileType `json:"file_type"`
	}
	err := c.ShouldBindJSON(&req)
	if err != nil {
		jsonErr(c, http.StatusInternalServerError, ErrUploadFail)
		return
	}
	up.FileType = req.FileType
	uid, _ := auth.Identity(c)
	up.Uid = uint(uid)
	hookReq := hook.GenerateFileTag(c)
	up.Tag = hookReq.Tag
	up.Name = req.FileName
	if req.RemoteUrl != "" {
		up.From = req.RemoteUrl
		err, fileLocal := service.NewFileService(req.Path).DownloadFile(up, false) // 文件上传后拿到文件路径
		if err != nil {
			log.Error("远程文件下载失败!", err)
			jsonErr(c, http.StatusInternalServerError, err)
			return
		}
		up.From = fileLocal.Url
	} else {
		up.Id = req.LocalId
	}
	err, file = service.NewFileService(req.Path).OSSAdd(up, false) //本地文件上传到IPFS
	if err != nil {
		log.Error("本地文件上传到OSS!", err)
		jsonErr(c, http.StatusInternalServerError, err)
		return
	}
	if file.Id > 0 {
		jsonResult(c, http.StatusOK, map[string]interface{}{
			"file": file,
		})
		return
	} else {
		jsonResult(c, http.StatusOK, map[string]interface{}{})
		return
	}
}

func UploadSyncWx(c *gin.Context) {
	var file model.Files
	var up model.Files
	var req struct {
		RemoteUrl string          `json:"remote_url"`
		LocalId   uint            `json:"local_id"`
		FileType  consts.FileType `json:"file_type"`
		Appid     string          `json:"appid"`
		TypeId    uint            `json:"type_id"`
	}
	err := c.ShouldBindJSON(&req)
	if err != nil {
		jsonErr(c, http.StatusInternalServerError, ErrUploadFail)
		return
	}
	pathStr := c.Param("path")
	if req.Appid == "" {
		jsonErr(c, http.StatusInternalServerError, errors.New("appid不能为空"))
		return
	}
	up.FileType = req.FileType
	uid, _ := auth.Identity(c)
	up.Uid = uint(uid)
	hookReq := hook.GenerateFileTag(c)
	up.Tag = hookReq.Tag
	up.Name = fun.MD5(req.RemoteUrl) + filepath.Ext(req.RemoteUrl)
	if req.RemoteUrl != "" {
		up.From = req.RemoteUrl
		err, fileLocal := service.NewFileService(pathStr).DownloadFile(up, true) // 文件上传后拿到文件路径
		if err != nil {
			log.Error("远程文件下载失败!", err)
			jsonErr(c, http.StatusInternalServerError, err)
			return
		}
		up.From = fileLocal.Url
	} else {
		up.Id = req.LocalId
	}
	up.TypeId = req.TypeId
	err, file = service.NewFileService(pathStr).WxAdd(req.Appid, up) //本地文件上传到IPFS
	if err != nil {
		log.Error("本地文件上传到OSS!", err)
		jsonErr(c, http.StatusInternalServerError, err)
		return
	}
	if file.Id > 0 {
		jsonResult(c, http.StatusOK, map[string]interface{}{
			"file": file,
		})
		return
	} else {
		jsonResult(c, http.StatusOK, map[string]interface{}{})
		return
	}
}

func ReadXls(c *gin.Context) {
	//uid, _ := auth.Identity(c)
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		jsonErr(c, http.StatusInternalServerError, err)
		return
	}
	defer file.Close()
	f, err := excelize.OpenReader(file)
	if err != nil {
		jsonErr(c, http.StatusInternalServerError, err)
		return
	}
	//xlsTmpName := fmt.Sprintf("%s/%d_%s.xlsx", setting.Upload.LocalPath, uid, time.Now().Format("20060102150405"))
	//f.Path = xlsTmpName
	sheetIndex := f.GetActiveSheetIndex()
	sheetName := f.GetSheetName(sheetIndex)
	rows, err := f.GetRows(sheetName)
	if err != nil {
		jsonErr(c, http.StatusInternalServerError, err)
		return
	}
	jsonResult(c, http.StatusOK, rows)
	return
}
