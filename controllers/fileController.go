package controllers

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"gorm.io/gorm"

	"supergit/inpatient/dtos"
	"supergit/inpatient/middleware"
	"supergit/inpatient/models"
	"supergit/inpatient/utils"
)

func UploadFile(c *gin.Context, sqlDB *gorm.DB) {
	user := middleware.GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse{Status: http.StatusUnauthorized, Message: "User not found in context"})
		return
	}
	var req dtos.UploadFileRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Status: http.StatusBadRequest, Message: "Invalid request: " + err.Error()})
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Status: http.StatusBadRequest, Message: "File is required"})
		return
	}

	fileContent, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Status: http.StatusInternalServerError, Message: "Failed to open file"})
		return
	}
	defer fileContent.Close()
	var business models.Business
	if err := sqlDB.Where("id = ?", user.BusinessId).First(&business).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Status: http.StatusInternalServerError, Message: "Failed to retrieve business information"})
		return
	}

	businessName := sanitizeFolderName(business.NameEn)
	if businessName == "" {
		businessName = fmt.Sprintf("business_%d", user.BusinessId)
	}
	var branch models.Branch
	branchName := "default"
	if user.BranchId > 0 {
		if err := sqlDB.Where("id = ?", user.BranchId).First(&branch).Error; err == nil {
			branchName = sanitizeFolderName(branch.NameEn)
			if branchName == "" {
				branchName = fmt.Sprintf("branch_%d", user.BranchId)
			}
		}
	}

	now := time.Now()
	year := now.Format("2006")
	month := now.Format("01")
	date := now.Format("02")

	folderPath := filepath.Join(businessName, branchName, year, month, date, sanitizeFolderName(req.FolderName))
	fileName := fmt.Sprintf("%d_%s", now.Unix(), file.Filename)
	objectName := filepath.Join(folderPath, fileName)
	objectName = strings.ReplaceAll(objectName, "\\", "/")

	minioClient := utils.GetMinioClient()

	ctx := context.Background()
	_, err = minioClient.Client.PutObject(ctx, minioClient.BucketName, objectName, fileContent, file.Size,
		minio.PutObjectOptions{
			ContentType: file.Header.Get("Content-Type"),
		},
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Status: http.StatusInternalServerError, Message: "Failed to upload file: " + err.Error()})
		return
	}

	fileURL := fmt.Sprintf("%s/%s/%s", minioClient.BaseURL, minioClient.BucketName, objectName)

	c.JSON(http.StatusOK, utils.SuccessResponse{Status: http.StatusOK, Message: "File uploaded successfully",
		Data: map[string]interface{}{
			"file": dtos.UploadFileResponse{
				FileName: fileName,
				FilePath: objectName,
				FileURL:  fileURL,
				FileSize: file.Size,
			},
		},
	})
}
func GetPresignedURL(c *gin.Context) {
	var req dtos.GetPresignedURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Status: http.StatusBadRequest, Message: "Invalid request: " + err.Error()})
		return
	}

	minioClient := utils.GetMinioClient()
	ctx := context.Background()
	expiry := time.Hour * 1

	presignedURL, err := minioClient.Client.PresignedGetObject(ctx, minioClient.BucketName, req.FilePath, expiry, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Status: http.StatusInternalServerError, Message: "Failed to generate presigned URL: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, utils.SuccessResponse{
		Status:  http.StatusOK,
		Message: "Presigned URL generated successfully",
		Data: map[string]interface{}{
			"url":        presignedURL.String(),
			"expires_in": 3600,
		},
	})
}

func DeleteFile(c *gin.Context) {
	var req dtos.DeleteFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Status: http.StatusBadRequest, Message: "Invalid request: " + err.Error()})
		return
	}

	minioClient := utils.GetMinioClient()
	ctx := context.Background()
	err := minioClient.Client.RemoveObject(
		ctx,
		minioClient.BucketName,
		req.FilePath,
		minio.RemoveObjectOptions{},
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Status: http.StatusInternalServerError, Message: "Failed to delete file: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse{
		Status:  http.StatusOK,
		Message: "File deleted successfully",
		Data:    nil,
	})
}
func sanitizeFolderName(name string) string {
	name = strings.ReplaceAll(name, " ", "_")
	var result strings.Builder
	for _, char := range name {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '_' || char == '-' {
			result.WriteRune(char)
		}
	}
	return result.String()
}
