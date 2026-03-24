package handlers

import (
	"blog/internal/Log"
	"blog/internal/config"
	"blog/internal/types"
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"blog/internal/sql"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

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
	file, err := ctx.FormFile("image") // 获取图片文件
	if err != nil {                    // 如果获取图片文件失败，返回 400 错误
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Message: "??????"})
		return
	}
	if file.Size > 1024*1024*2 {
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{
			Message: "????????2MB",
		})
		return
	} // 如果图片文件大小大于 2MB，返回 400 错误

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{
			Message: "???????JPG?JPEG?PNG",
		})
		return
	}

	src, err := file.Open()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Message: "??????"})
		return
	}
	defer src.Close()
	header := make([]byte, 512)
	n, _ := io.ReadFull(src, header)
	mime := http.DetectContentType(header[:n])
	if mime != "image/jpeg" && mime != "image/png" {
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Message: "??????"})
		return
	}

	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, types.ErrorResponse{
			Message: "???",
		})
		return
	}
	m := sql.NewUserMapper()
	ID, ok := userID.(int)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, types.ErrorResponse{
			Message: "???",
		})
		return
	}

	fileName := RandomFileName() + ext
	uploadRoot := config.Get().Upload.Dir
	if uploadRoot == "" {
		uploadRoot = "uploads"
	}
	userDir := filepath.Join(uploadRoot, fmt.Sprintf("%d", ID))
	if err = os.MkdirAll(userDir, 0755); err != nil {
		ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Message: "??????"})
		return
	}
	dst := filepath.Join(userDir, fileName)

	// ????????????????
	src2, err := file.Open()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Message: "??????"})
		return
	}
	defer src2.Close()
	buf := bytes.NewBuffer(nil)
	if _, err = io.Copy(buf, src2); err != nil {
		ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Message: "??????"})
		return
	}
	if err = os.WriteFile(dst, buf.Bytes(), 0644); err != nil {
		ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Message: "??????"})
		return
	}

	err = m.UpdateImage(ID, fileName)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Message: "??????",
		})
		return
	}

	Log.ZLog.Info("upload_image_success", zap.Int("user_id", ID), zap.String("file", dst))
	ctx.JSON(http.StatusOK, types.SuccessResponse{
		Message: "??????",
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
		ctx.JSON(http.StatusUnauthorized, types.ErrorResponse{Message: "???"})
		return
	}
	id, ok := userID.(int)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, types.ErrorResponse{Message: "???"})
		return
	}
	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Message: "参数错误"})
		return
	}
	url, err := saveMarkdownContentForUser(id, string(body))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Message: "??????"})
		return
	}
	ctx.JSON(http.StatusOK, types.SuccessResponse{
		Message: "??????",
		Data: gin.H{
			"url": url,
		},
	})
}

func saveMarkdownContentForUser(userID int, content string) (string, error) {
	if strings.TrimSpace(content) == "" {
		return "", errors.New("content is empty")
	}
	uploadRoot := config.Get().Upload.Dir
	if uploadRoot == "" {
		uploadRoot = "uploads"
	}
	userDir := filepath.Join(uploadRoot, fmt.Sprintf("%d", userID))
	if err := os.MkdirAll(userDir, 0755); err != nil {
		return "", err // 如果创建用户目录失败，返回错误
	}
	filename := RandomFileName() + ".md"
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
// @Param			filename	path	string	true	"文件名"
// @Success		200		{file}		file
// @Failure		404		{object}	types.ErrorResponse
// @router			/file/getFile [get]
func GetFileHandler(ctx *gin.Context) {
	filename := ctx.Param("filename")
	filepath := filepath.Join(config.Get().Upload.Dir, filename)
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		ctx.JSON(http.StatusNotFound, types.ErrorResponse{Message: "File not found"})
		return
	}
	ctx.File(filepath) // 返回文件内容
}
