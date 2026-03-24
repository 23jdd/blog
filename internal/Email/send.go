// Package email 提供基于 SMTP 的简单发信（依赖 viper 读取配置）。
//
// 在 config.yaml 中可配置（字段名与 viper 键一致，注意 yaml 嵌套）：
//
//	smtp:
//	  host: smtp.example.com
//	  port: 587
//	  username: "your_user"
//	  password: "your_pass"
//	  from: "noreply@example.com"
//
// 常见：587 端口 + STARTTLS（net/smtp.SendMail 会自动协商）。
package email

import (
	"blog/internal/config"
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"net/smtp"
	"strings"
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
	cfg := config.Get()
	host := strings.TrimSpace(cfg.SMTP.Host)
	from := strings.TrimSpace(cfg.SMTP.From)
	if host == "" || from == "" {
		return errors.New("smtp 未配置：请在配置中设置 smtp.host 与 smtp.from")
	}
	if len(to) == 0 {
		return errors.New("收件人不能为空")
	}
	port := cfg.SMTP.Port
	if port <= 0 {
		port = 587
	}
	user := strings.TrimSpace(cfg.SMTP.Username)
	pass := cfg.SMTP.Password

	addr := fmt.Sprintf("%s:%d", host, port)

	var auth smtp.Auth
	if user != "" {
		auth = smtp.PlainAuth("", user, pass, host)
	}

	var buf bytes.Buffer
	fmt.Fprintf(&buf, "From: %s\r\n", from)
	fmt.Fprintf(&buf, "To: %s\r\n", strings.Join(to, ","))
	fmt.Fprintf(&buf, "Subject: %s\r\n", mimeEncodeSubject(subject))
	fmt.Fprintf(&buf, "MIME-Version: 1.0\r\n")
	fmt.Fprintf(&buf, "Content-Type: text/plain; charset=UTF-8\r\n")
	fmt.Fprintf(&buf, "Content-Transfer-Encoding: 8bit\r\n")
	fmt.Fprintf(&buf, "\r\n%s", body)

	return smtp.SendMail(addr, auth, from, to, buf.Bytes())
}

// 简单 RFC 2047 编码主题中的非 ASCII 字符（仅处理整段为 UTF-8 的常见情况）。
func mimeEncodeSubject(s string) string {
	if s == "" {
		return ""
	}
	ascii := true
	for _, r := range s {
		if r > 127 {
			ascii = false
			break
		}
	}
	if ascii {
		return s
	}
	return fmt.Sprintf("=?UTF-8?B?%s?=", base64.StdEncoding.EncodeToString([]byte(s)))
}
