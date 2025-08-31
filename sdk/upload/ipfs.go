package upload

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"

	"github.com/hb1707/ant-godmin/setting"
)

type IPFS struct{}
type IpfsRespBSN struct {
	Name string `json:"Name"`
	Hash string `json:"Hash"`
	Size string `json:"Size"`
}

func (c *IPFS) SetPath(path string) {

}

func (c *IPFS) SetBucket(bucketName string) {}

// AllObjects 列举所有文件的信息
func (*IPFS) AllObjects(path string, continuation string) (pathList []map[string]string, next string, err error) {

	return
}

// GetUrl 获取文件的访问地址
func (*IPFS) GetUrl(key string, isPrivate bool, expire int64) string {
	if isPrivate {
		return setting.IPFS.IpfsEndpoint + "/ipfs/" + key
	}
	return setting.IPFS.IpfsEndpoint + "/ipfs/" + key
}

// GetInfo 文件的信息
func (*IPFS) GetInfo(key string) (info map[string]any, err error) {
	return
}
func (*IPFS) Upload(file io.Reader, newFileName string, other ...string) (string, error) {
	extraParams := map[string]string{}
	request, err := uploadRequest(setting.IPFS.IpfsEndpoint+"/api/v0/add", file, "arg", filepath.Base(newFileName), extraParams)
	if err != nil {
		return "", err
	}
	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return "", err
	} else {
		body := &bytes.Buffer{}
		_, err := body.ReadFrom(resp.Body)
		if err != nil {
			return "", err
		}
		resp.Body.Close()
		if resp.StatusCode != 200 {
			return "", fmt.Errorf("upload failed: %s", body.String())
		}
		var respBody IpfsRespBSN
		err = json.Unmarshal(body.Bytes(), &respBody)
		if err != nil {
			return "", err
		}
		return respBody.Hash, nil
	}
}
func (c *IPFS) AsyncProcessObject(sourceKey, process string) (map[string]string, error) {
	return nil, nil
}
func (*IPFS) Copy(ori string, new string) error {
	return nil
}
func (*IPFS) Download(path string, localFileName string) (string, error) {
	return "", nil
}

func (*IPFS) Delete(key string) error {
	return nil
}

func uploadRequest(uri string, file io.Reader, paramName string, fileName string, params map[string]string) (*http.Request, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(paramName, fileName)
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, file)

	for key, val := range params {
		_ = writer.WriteField(key, val)
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", uri, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, err
}
