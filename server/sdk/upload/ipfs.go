package upload

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/hb1707/ant-godmin/setting"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
)

type IPFS struct{}
type IpfsRespBSN struct {
	Name string `json:"Name"`
	Hash string `json:"Hash"`
	Size string `json:"Size"`
}

func (*IPFS) Upload(file io.Reader, newFileName string, other ...string) (string, error) {
	extraParams := map[string]string{}
	request, err := uploadRequest(setting.Upload.IpfsEndpoint+"/api/v0/add", file, "arg", filepath.Base(newFileName), extraParams)
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
			return "", fmt.Errorf("Upload failed: %s", body.String())
		}
		var respBody IpfsRespBSN
		err = json.Unmarshal(body.Bytes(), &respBody)
		if err != nil {
			return "", err
		}
		return respBody.Hash, nil
	}
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
