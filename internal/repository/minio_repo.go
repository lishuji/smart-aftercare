package repository

import (
	"smart-aftercare/config"
	"context"
	"fmt"
	"io"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinioRepo MinIO 对象存储数据访问层
type MinioRepo struct {
	client *minio.Client
	bucket string
}

// NewMinIORepo 创建 MinIO 数据访问层
func NewMinIORepo(cfg config.MinioConfig) (*MinioRepo, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("创建MinIO客户端失败: %w", err)
	}

	// 确保 Bucket 存在
	ctx := context.Background()
	exists, err := client.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("检查Bucket失败: %w", err)
	}
	if !exists {
		if err := client.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("创建Bucket失败: %w", err)
		}
	}

	return &MinioRepo{
		client: client,
		bucket: cfg.Bucket,
	}, nil
}

// UploadFile 上传文件到 MinIO，返回对象访问 URL
func (m *MinioRepo) UploadFile(filePath, objectKey string) (string, error) {
	ctx := context.Background()

	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return "", fmt.Errorf("获取文件信息失败: %w", err)
	}

	// 上传文件
	_, err = m.client.PutObject(ctx, m.bucket, objectKey, file, fileInfo.Size(), minio.PutObjectOptions{
		ContentType: detectContentType(filePath),
	})
	if err != nil {
		return "", fmt.Errorf("上传文件到MinIO失败: %w", err)
	}

	// 返回对象访问路径
	return fmt.Sprintf("/%s/%s", m.bucket, objectKey), nil
}

// UploadReader 上传 io.Reader 到 MinIO
func (m *MinioRepo) UploadReader(reader io.Reader, objectKey string, size int64, contentType string) (string, error) {
	ctx := context.Background()

	_, err := m.client.PutObject(ctx, m.bucket, objectKey, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("上传数据到MinIO失败: %w", err)
	}

	return fmt.Sprintf("/%s/%s", m.bucket, objectKey), nil
}

// DownloadFile 从 MinIO 下载文件
func (m *MinioRepo) DownloadFile(objectKey, destPath string) error {
	ctx := context.Background()
	return m.client.FGetObject(ctx, m.bucket, objectKey, destPath, minio.GetObjectOptions{})
}

// GetObject 获取对象的 io.Reader
func (m *MinioRepo) GetObject(objectKey string) (io.ReadCloser, error) {
	ctx := context.Background()
	return m.client.GetObject(ctx, m.bucket, objectKey, minio.GetObjectOptions{})
}

// DeleteObject 删除对象
func (m *MinioRepo) DeleteObject(objectKey string) error {
	ctx := context.Background()
	return m.client.RemoveObject(ctx, m.bucket, objectKey, minio.RemoveObjectOptions{})
}

// ObjectExists 检查对象是否存在
func (m *MinioRepo) ObjectExists(objectKey string) (bool, error) {
	ctx := context.Background()
	_, err := m.client.StatObject(ctx, m.bucket, objectKey, minio.StatObjectOptions{})
	if err != nil {
		errResp := minio.ToErrorResponse(err)
		if errResp.Code == "NoSuchKey" {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// ListObjects 列出指定前缀下的所有对象
func (m *MinioRepo) ListObjects(prefix string) ([]minio.ObjectInfo, error) {
	ctx := context.Background()

	var objects []minio.ObjectInfo
	objectCh := m.client.ListObjects(ctx, m.bucket, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})

	for obj := range objectCh {
		if obj.Err != nil {
			return nil, fmt.Errorf("列出对象失败: %w", obj.Err)
		}
		objects = append(objects, obj)
	}

	return objects, nil
}

// detectContentType 根据文件扩展名检测 Content-Type
func detectContentType(filePath string) string {
	ext := ""
	for i := len(filePath) - 1; i >= 0; i-- {
		if filePath[i] == '.' {
			ext = filePath[i:]
			break
		}
	}

	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".pdf":
		return "application/pdf"
	case ".doc", ".docx":
		return "application/msword"
	case ".txt":
		return "text/plain"
	default:
		return "application/octet-stream"
	}
}
