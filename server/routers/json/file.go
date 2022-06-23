package json

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/hb1707/ant-godmin/auth"
	"github.com/hb1707/ant-godmin/consts"
	"github.com/hb1707/ant-godmin/model"
	"github.com/hb1707/ant-godmin/pkg/log"
	"github.com/hb1707/ant-godmin/routers/hook"
	"github.com/hb1707/ant-godmin/service"
	"github.com/hb1707/exfun/fun"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"
)

var (
	ErrUploadFail = errors.New("接收文件失败")
)

func UploadFile(c *gin.Context) {
	var file model.Files
	var req model.Files
	pathStr := c.Param("path")
	typeId, _ := strconv.Atoi(c.DefaultQuery("type_id", "0"))
	fileType, _ := strconv.Atoi(c.DefaultQuery("file_type", "1"))
	photoId, _ := strconv.Atoi(c.DefaultQuery("photo_id", "0")) //todo: 20220621之后准备废弃
	fileId, _ := strconv.Atoi(c.DefaultQuery("file_id", "0"))
	fileName := c.DefaultQuery("file_name", "")
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
	err, file = service.NewFileService(pathStr).UploadToOSS(header, req) // 文件上传后拿到文件路径
	if err != nil {
		log.Error("修改数据库链接失败!", err)
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
	err, file = service.NewFileService(req.Path).DownloadFile(up) // 文件上传后拿到文件路径
	if err != nil {
		log.Error("修改数据库链接失败!", err)
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
func AddIPFS(c *gin.Context) {
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
	up.From = "./temp.json"
	hookReq := hook.GenerateFileTag(c)
	up.Tag = hookReq.Tag
	up.Name = req.FileName
	err, file = service.NewFileService(req.Path).IPFSAdd(up) // 文件上传后拿到文件路径
	if err != nil {
		log.Error("修改数据库链接失败!", err)
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
