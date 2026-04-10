package utils

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioClient struct {
	Client      *minio.Client
	BucketName  string
	BaseURL     string
	Endpoint    string
	AccessKey   string
	SecretKey   string
	UseSSL      bool
	initialized bool
}

var (
	instance *MinioClient
	once     sync.Once
)

func GetMinioClient() *MinioClient {
	once.Do(func() {
		instance = &MinioClient{}
		if err := instance.Initialize(); err != nil {
			log.Fatalf("Failed to initialize MinIO client: %v", err)
		}
	})
	return instance
}

func (m *MinioClient) Initialize() error {
	m.Endpoint = getEnv("MINIO_ENDPOINT", "")
	m.AccessKey = getEnv("MINIO_ACCESS_KEY", "")
	m.SecretKey = getEnv("MINIO_SECRET_KEY", "")
	m.BucketName = getEnv("MINIO_BUCKET_NAME", "")
	m.UseSSL = getEnvBool("MINIO_USE_SSL", false)

	if m.Endpoint == "" || m.AccessKey == "" || m.SecretKey == "" || m.BucketName == "" {
		return fmt.Errorf("MinIO configuration is incomplete. Please set MINIO_ENDPOINT, MINIO_ACCESS_KEY, MINIO_SECRET_KEY, and MINIO_BUCKET_NAME in .env file")
	}

	protocol := "http"
	if m.UseSSL {
		protocol = "https"
	}
	m.BaseURL = fmt.Sprintf("%s://%s", protocol, m.Endpoint)

	client, err := minio.New(m.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(m.AccessKey, m.SecretKey, ""),
		Secure: m.UseSSL,
	})
	if err != nil {
		return fmt.Errorf("failed to create MinIO client: %v", err)
	}

	m.Client = client
	m.initialized = true

	if err := m.EnsureBucket(); err != nil {
		return fmt.Errorf("failed to ensure bucket exists: %v", err)
	}

	log.Printf("MinIO client initialized successfully. Endpoint: %s, Bucket: %s", m.Endpoint, m.BucketName)
	return nil
}

func (m *MinioClient) EnsureBucket() error {
	ctx := context.Background()
	exists, err := m.Client.BucketExists(ctx, m.BucketName)
	if err != nil {
		return fmt.Errorf("failed to check if bucket exists: %v", err)
	}

	if !exists {
		err = m.Client.MakeBucket(ctx, m.BucketName, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("failed to create bucket: %v", err)
		}
		log.Printf("Bucket '%s' created successfully", m.BucketName)
	} else {
		log.Printf("Bucket '%s' already exists", m.BucketName)
	}

	return nil
}

func (m *MinioClient) IsInitialized() bool {
	return m.initialized
}

func (m *MinioClient) GetClient() *minio.Client {
	return m.Client
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getEnvBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value == "true" || value == "1"
}
