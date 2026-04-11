package controllers

import (
	"context"
	"crypto/sha1"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"supergit/inpatient/utils"

	"github.com/gin-gonic/gin"
)

// UploadImage — admin only, proxies file to Cloudinary and returns the secure URL
func UploadImage(c *gin.Context) {
	cloudName := os.Getenv("CLOUDINARY_CLOUD_NAME")
	apiKey := os.Getenv("CLOUDINARY_API_KEY")
	apiSecret := os.Getenv("CLOUDINARY_API_SECRET")

	if cloudName == "" || apiKey == "" || apiSecret == "" {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Status: http.StatusInternalServerError, Message: "Cloudinary not configured"})
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Status: http.StatusBadRequest, Message: "No file provided"})
		return
	}
	defer file.Close()

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	folder := "products"

	// Build signature: folder=products&timestamp=<ts><secret>
	params := map[string]string{
		"folder":    folder,
		"timestamp": timestamp,
	}
	signature := cloudinarySignature(params, apiSecret)

	uploadURL := fmt.Sprintf("https://api.cloudinary.com/v1_1/%s/image/upload", cloudName)
	secureURL, err := doCloudinaryUpload(uploadURL, file, header, apiKey, timestamp, folder, signature)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Status: http.StatusInternalServerError, Message: "Upload failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse{Status: http.StatusOK, Message: "Uploaded", Data: gin.H{"url": secureURL}})
}

func cloudinarySignature(params map[string]string, secret string) string {
	// Sort keys
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(params))
	for _, k := range keys {
		parts = append(parts, k+"="+params[k])
	}
	toSign := strings.Join(parts, "&") + secret

	h := sha1.New()
	h.Write([]byte(toSign))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func doCloudinaryUpload(uploadURL string, file multipart.File, header *multipart.FileHeader, apiKey, timestamp, folder, signature string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pr, pw := io.Pipe()
	mw := multipart.NewWriter(pw)

	go func() {
		defer pw.Close()
		defer mw.Close()

		mw.WriteField("api_key", apiKey)
		mw.WriteField("timestamp", timestamp)
		mw.WriteField("folder", folder)
		mw.WriteField("signature", signature)

		fw, err := mw.CreateFormFile("file", header.Filename)
		if err != nil {
			return
		}
		io.Copy(fw, file)
	}()

	req, err := http.NewRequestWithContext(ctx, "POST", uploadURL, pr)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("cloudinary error: %s", string(body))
	}

	// Parse secure_url from JSON response manually (avoid extra deps)
	bodyStr := string(body)
	secureURL := extractJSONString(bodyStr, "secure_url")
	if secureURL == "" {
		return "", fmt.Errorf("no secure_url in response")
	}
	// Cloudinary escapes slashes in JSON — unescape
	secureURL, _ = url.QueryUnescape(strings.ReplaceAll(secureURL, "\\/", "/"))
	return secureURL, nil
}

func extractJSONString(json, key string) string {
	needle := `"` + key + `":"`
	idx := strings.Index(json, needle)
	if idx == -1 {
		return ""
	}
	start := idx + len(needle)
	end := strings.Index(json[start:], `"`)
	if end == -1 {
		return ""
	}
	return json[start : start+end]
}
