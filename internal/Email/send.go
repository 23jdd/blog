package send

import (
	"blog/internal/config"
	"crypto/tls"
	"net/smtp"
	"strconv"
	"strings"

	"github.com/jordan-wright/email"
)

// Configured 是否已配置可发送邮件的最小项（host、from）。
func Configured() bool {
	cfg := config.Get()
	return strings.TrimSpace(cfg.SMTP.Host) != "" &&
		strings.TrimSpace(cfg.SMTP.From) != ""
}

// SendPlain 发送纯文本邮件（UTF-8）。适用于验证码、通知等。
// to：收件人列表；subject、body：主题与正文。
func SendPlain(to []string, subject, body string) error {
	e := email.NewEmail()
	e.From = config.Get().SMTP.From
	e.To = to
	e.Subject = subject
	e.HTML = []byte(body)
	auth := smtp.PlainAuth("", e.From, config.Get().SMTP.Password, config.Get().SMTP.Host)
	tlsConfig := &tls.Config{ServerName: config.Get().SMTP.Host}
	return e.SendWithTLS(config.Get().SMTP.Host+":"+strconv.Itoa(config.Get().SMTP.Port), auth, tlsConfig)
}
