package dtos

type UploadFileRequest struct {
	FolderName string `form:"folder_name" binding:"required"`
}

type UploadFileResponse struct {
	FileName string `json:"file_name"`
	FilePath string `json:"file_path"`
	FileURL  string `json:"file_url"`
	FileSize int64  `json:"file_size"`
}

type GetPresignedURLRequest struct {
	FilePath string `json:"file_path" binding:"required"`
}

type GetPresignedURLResponse struct {
	URL       string `json:"url"`
	ExpiresIn int    `json:"expires_in"`
}

type DeleteFileRequest struct {
	FilePath string `json:"file_path" binding:"required"`
}
