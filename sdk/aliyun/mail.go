package aliyun

import (
	"crypto/tls"
	"fmt"
	"github.com/hb1707/ant-godmin/pkg/log"
	"github.com/hb1707/exfun/fun"
	"net"
	"net/smtp"
)

var MailPWD = ""
var MailFromName = ""

func SendEmail(fromEmail string, toEmail string, subject string, body string) {
	host := "smtpdm.aliyun.com"
	port := 465
	email := fromEmail
	password := MailPWD

	header := make(map[string]string)
	header["From"] = fmt.Sprintf(`"=?utf-8?B?%s?="`, fun.Base64Encode(MailFromName, false)) + " <" + email + ">"
	header["To"] = toEmail
	header["Subject"] = subject
	header["Content-Type"] = "text/html; charset=UTF-8"
	message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body

	auth := smtp.PlainAuth(
		"",
		email,
		password,
		host,
	)

	err := sendMailUsingTLS(
		fmt.Sprintf("%s:%d", host, port),
		auth,
		email,
		[]string{toEmail},
		[]byte(message),
	)

	if err != nil {
		log.Error(err)
	}
}

// dial
// return a smtp client
func dial(addr string) (*smtp.Client, error) {
	conn, err := tls.Dial("tcp", addr, nil)
	if err != nil {
		return nil, err
	}
	//分解主机端口字符串
	host, _, _ := net.SplitHostPort(addr)
	return smtp.NewClient(conn, host)
}

// sendMailUsingTLS
// 参考net/smtp的func SendMail()
// 使用net.Dial连接tls（SSL）端口时，smtp.NewClient()会卡住且不提示err
// len(to)>1时，to[1]开始提示是密送
func sendMailUsingTLS(addr string, auth smtp.Auth, from string,
	to []string, msg []byte) (err error) {

	//create smtp client
	c, err := dial(addr)
	if err != nil {
		return err
	}
	defer c.Close()

	if auth != nil {
		if ok, _ := c.Extension("AUTH"); ok {
			if err = c.Auth(auth); err != nil {
				return err
			}
		}
	}

	if err = c.Mail(from); err != nil {
		return err
	}

	for _, addr := range to {
		if err = c.Rcpt(addr); err != nil {
			return err
		}
	}

	w, err := c.Data()
	if err != nil {
		return err
	}

	_, err = w.Write(msg)
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	return c.Quit()
}
