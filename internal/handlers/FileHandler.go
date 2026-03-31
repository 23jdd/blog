package handlers

import (
	"blog/internal/Log"
	"blog/internal/config"
	"blog/internal/types"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"blog/internal/sql"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const maxAvatarBytes = 2 * 1024 * 1024           // 头像最大 2MB
const maxArticleMarkdownBytes = 10 * 1024 * 1024 // 文章 Markdown 正文最大 10MB

// ErrEmptyArticleContent 表示上传的正文为空（仅空白）。
var ErrEmptyArticleContent = errors.New("文章内容不能为空")

// @Summary		SetPersonImage
// @Description	SetPersonImage
// @Tags			file
// @Accept			multipart/form-data
// @Produce		json
// @Param			image	formData	file	true	"图片文件"
// @Success		200		{object}	types.SuccessResponse
// @Failure		400		{object}	types.ErrorResponse
// @Failure		500		{object}	types.ErrorResponse
// @router			/file/setPersonImage [post]
func SetPersonImage(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, types.ErrorResponse{Message: "未登录"})
		return
	} // 如果用户ID不存在，返回 401 错误
	ID, ok := userID.(int)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, types.ErrorResponse{Message: "用户身份无效"})
		return
	} // 如果用户ID无效，返回 401 错误

	file, err := ctx.FormFile("image") // 获取图片文件
	if err != nil {
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Message: "请选择图片文件"})
		return
	} // 如果获取图片文件失败，返回 400 错误
	if file.Size > maxAvatarBytes {
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{
			Message: "图片大小不能超过 2MB",
		})
		return
	} // 如果图片大小超过 2MB，返回 400 错误

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{
			Message: "仅支持 JPG、JPEG、PNG 格式",
		})
		return
	} // 如果文件格式不支持，返回 400 错误

	src, err := file.Open() // 打开图片文件
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Message: "读取文件失败"})
		return
	}
	defer src.Close() // 关闭图片文件

	// 单次读取：限制实际字节数（防止声明大小与真实流不一致），再用魔数校验真实类型
	data, err := io.ReadAll(io.LimitReader(src, maxAvatarBytes+1)) // 读取图片文件
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Message: "读取文件失败"})
		return
	}
	if len(data) > maxAvatarBytes { // 如果图片大小超过 2MB，返回 400 错误
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{
			Message: "图片大小不能超过 2MB",
		})
		return
	}
	if len(data) == 0 { // 如果图片为空，返回 400 错误
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Message: "文件为空"})
		return
	}
	mime := http.DetectContentType(data)             // 检测图片类型
	if mime != "image/jpeg" && mime != "image/png" { // 如果图片类型不是 JPG/PNG，返回 400 错误
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Message: "文件内容不是有效的 JPG/PNG 图片"})
		return
	}

	fileName := RandomFileName() + ext // 生成文件名
	uploadRoot := config.Get().Upload.Dir
	if uploadRoot == "" { // 如果上传目录为空，设置默认目录
		uploadRoot = "uploads"
	}
	userDir := filepath.Join(uploadRoot, fmt.Sprintf("%d", ID))
	if err = os.MkdirAll(userDir, 0755); err != nil {
		ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Message: "创建目录失败"})
		return
	}
	dst := filepath.Join(userDir, fileName)

	if err = os.WriteFile(dst, data, 0644); err != nil {
		ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Message: "保存文件失败"})
		return
	}
	url := config.Get().Upload.URLPrefix + "/" + fmt.Sprintf("%d", ID) + "/" + fileName
	m := sql.NewUserMapper()
	if err = m.UpdateImage(ID, url); err != nil {
		_ = os.Remove(dst)
		ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Message: "更新头像失败",
		})
		return
	}

	Log.ZLog.Info("upload_image_success", zap.Int("user_id", ID), zap.String("file", dst))
	ctx.JSON(http.StatusOK, types.SuccessResponse{
		Message: "头像上传成功",
		Data: gin.H{
			"url": url,
		},
	})
}

// @Summary		UpLoadArticleFile
// @Description	UpLoadArticleFile
// @Tags			file
// @Accept			multipart/form-data
// @Produce		json
// @Param			article	formData	file	true	"文章文件"
// @Success		200		{object}	types.SuccessResponse
// @Failure		400		{object}	types.ErrorResponse
// @Failure		500		{object}	types.ErrorResponse
// @router			/file/uploadArticle [post]
func UpLoadArticleFile(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, types.ErrorResponse{Message: "未登录"})
		return
	}
	id, ok := userID.(int)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, types.ErrorResponse{Message: "用户身份无效"})
		return
	}

	var body []byte
	var err error
	ct := ctx.GetHeader("Content-Type")
	if strings.Contains(ct, "multipart/form-data") {
		file, ferr := ctx.FormFile("article")
		if ferr != nil {
			ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Message: "请通过表单字段 article 上传文件"})
			return
		}
		if file.Size > maxArticleMarkdownBytes {
			ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Message: "文章内容不能超过 10MB"})
			return
		}
		src, ferr := file.Open()
		if ferr != nil {
			ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Message: "读取上传文件失败"})
			return
		}
		body, err = io.ReadAll(io.LimitReader(src, maxArticleMarkdownBytes+1))
		_ = src.Close()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Message: "读取上传文件失败"})
			return
		}
	} else {
		body, err = io.ReadAll(io.LimitReader(ctx.Request.Body, maxArticleMarkdownBytes+1))
		if err != nil {
			ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Message: "读取请求体失败"})
			return
		}
	}

	if len(body) > maxArticleMarkdownBytes {
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Message: "文章内容不能超过 10MB"})
		return
	}
	prefix := ctx.DefaultQuery("type", "md")
	url, err := saveMarkdownContentForUser(id, prefix, string(body))
	if err != nil {
		if errors.Is(err, ErrEmptyArticleContent) {
			ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Message: "文章内容不能为空"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Message: "保存文章文件失败"})
		return
	}
	ctx.JSON(http.StatusOK, types.SuccessResponse{
		Message: "文章上传成功",
		Data: gin.H{
			"url": url,
		},
	})
}

// saveMarkdownContentForUser 保存 Markdown 内容到用户目录
func saveMarkdownContentForUser(userID int, prefix string, content string) (string, error) {
	if strings.TrimSpace(content) == "" {
		return "", ErrEmptyArticleContent
	}
	uploadRoot := config.Get().Upload.Dir
	if uploadRoot == "" {
		uploadRoot = "uploads"
	}
	userDir := filepath.Join(uploadRoot, fmt.Sprintf("%d", userID))
	if err := os.MkdirAll(userDir, 0755); err != nil {
		return "", err // 如果创建用户目录失败，返回错误
	}
	filename := RandomFileName() + "." + prefix
	dst := filepath.Join(userDir, filename)
	if err := os.WriteFile(dst, []byte(content), 0644); err != nil {
		return "", err // 如果写入文件失败，返回错误
	}
	return config.Get().Upload.URLPrefix + "/" + fmt.Sprintf("%d", userID) + "/" + filename, nil
}

func RandomFileName() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b) // 返回随机文件名
}

// @Summary		GetFileHandler
// @Description	GetFileHandler
// @Tags			file
// @Produce		octet-stream
// @Param			filename	path		string	true	"文件名"
// @Success		200			{file}		file
// @Failure		404			{object}	types.ErrorResponse
// @router			/file/getFile/:filename [get]
func GetFileHandler(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, types.ErrorResponse{Message: "未登录"})
		return
	}
	id, ok := userID.(int)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, types.ErrorResponse{Message: "用户身份无效"})
		return
	}
	filename := ctx.Param("filename")
	filepath := filepath.Join(config.Get().Upload.Dir, strconv.Itoa(id), filename)
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		ctx.JSON(http.StatusNotFound, types.ErrorResponse{Message: "File not found"})
		return
	}
	ctx.File(filepath) // 返回文件内容
}
