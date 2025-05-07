package aliyun

import (
	"errors"
	"fmt"
	openapi "github.com/alibabacloud-go/darabonba-openapi/client"
	dysmsapi20170525 "github.com/alibabacloud-go/dysmsapi-20170525/v2/client"
	util "github.com/alibabacloud-go/tea-utils/service"
	"github.com/hb1707/ant-godmin/pkg/log"
)

var Endpoint = "sms.aliyuncs.com"

// CreateClient 创建短信客户端
func CreateClient(accessKeyId *string, accessKeySecret *string) (_result *dysmsapi20170525.Client, _err error) {
	config := &openapi.Config{
		// 您的 AccessKey ID
		AccessKeyId: accessKeyId,
		// 您的 AccessKey Secret
		AccessKeySecret: accessKeySecret,
	}
	// 访问的域名
	config.Endpoint = &Endpoint
	_result = &dysmsapi20170525.Client{}
	_result, _err = dysmsapi20170525.NewClient(config)
	return _result, _err
}

func Send(phoneNumbers string, templateParam, signName, templateCode string) (resp *dysmsapi20170525.SendSmsResponse, _err error) {
	client, _err := CreateClient(&AccessKeyId, &AccessKeySecret)
	if _err != nil {
		return nil, _err
	}

	sendSmsRequest := &dysmsapi20170525.SendSmsRequest{
		SignName:      &signName,
		TemplateCode:  &templateCode,
		PhoneNumbers:  &phoneNumbers,
		TemplateParam: &templateParam,
	}
	runtime := &util.RuntimeOptions{}
	// 复制代码运行请自行打印 API 的返回值
	resp, _err = client.SendSmsWithOptions(sendSmsRequest, runtime)
	if _err != nil {
		log.Error("SendSmsWithOptions error:", _err)
		return nil, errors.New("短信发送失败，请联系客服")
	}
	if *resp.Body.Code != "OK" {
		if *resp.Body.Code == "isv.BUSINESS_LIMIT_CONTROL" {
			return nil, errors.New("您的操作太频繁了，请稍后再试")
		}
		log.Error("短信发送失败！", fmt.Sprintf("%+v", *resp.Body), phoneNumbers, templateCode)
		return resp, errors.New("短信发送失败！" + *resp.Body.Message)
	}
	return resp, nil

}
