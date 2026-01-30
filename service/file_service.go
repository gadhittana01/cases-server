package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/gadhittana01/cases-app-server/db/repository"
	"github.com/gadhittana01/cases-app-server/dto"
	"github.com/gadhittana01/cases-modules/utils"
	"github.com/google/uuid"
)

type FileService struct {
	repo          repository.Repository
	s3Client      *s3.Client
	presignClient *s3.PresignClient
	config        *utils.Config
}

func NewFileService(repo repository.Repository, s3Client *s3.Client, presignClient *s3.PresignClient, config *utils.Config) *FileService {
	return &FileService{
		repo:          repo,
		s3Client:      s3Client,
		presignClient: presignClient,
		config:        config,
	}
}

func (s *FileService) UploadCaseFile(ctx context.Context, caseID uuid.UUID, fileHeader *multipart.FileHeader) (*dto.FileResponse, error) {
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if ext != ".pdf" && ext != ".png" {
		return nil, fmt.Errorf("only PDF and PNG files are allowed")
	}

	const maxSize = 10 * 1024 * 1024
	if fileHeader.Size > maxSize {
		return nil, fmt.Errorf("file size exceeds 10MB limit")
	}

	count, err := s.repo.CountCaseFilesByCaseID(ctx, caseID)
	if err != nil {
		return nil, fmt.Errorf("failed to check file count: %w", err)
	}
	if count >= 10 {
		return nil, fmt.Errorf("maximum 10 files allowed per case")
	}

	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	secureFilename := generateSecureFilename(fileHeader.Filename)
	filePath := fmt.Sprintf("cases/%s/%s", caseID.String(), secureFilename)

	contentType := fileHeader.Header.Get("Content-Type")

	_, err = s.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.config.StorageBucket),
		Key:         aws.String(filePath),
		Body:        file,
		ContentType: aws.String(contentType),
		ACL:         types.ObjectCannedACLPrivate,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload file to storage: %w", err)
	}

	fileRecord, err := s.repo.CreateCaseFile(ctx, &repository.CreateCaseFileParams{
		CaseID:   caseID,
		FileName: fileHeader.Filename,
		FilePath: filePath,
		FileSize: fileHeader.Size,
		MimeType: contentType,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to save file record: %w", err)
	}

	return &dto.FileResponse{
		ID:        fileRecord.ID,
		FileName:  fileRecord.FileName,
		FileSize:  fileRecord.FileSize,
		MimeType:  fileRecord.MimeType,
		CreatedAt: utils.PgtypeTimeToTime(fileRecord.CreatedAt),
	}, nil
}

func (s *FileService) GenerateDownloadURL(ctx context.Context, fileID uuid.UUID, userID uuid.UUID, userRole string) (string, error) {

	fileRecord, err := s.repo.GetCaseFileByID(ctx, fileID)
	if err != nil {
		return "", fmt.Errorf("file not found: %w", err)
	}

	caseRecord, err := s.repo.GetCaseByID(ctx, fileRecord.CaseID)
	if err != nil {
		return "", fmt.Errorf("case not found: %w", err)
	}

	if userRole == "client" {
		if caseRecord.ClientID != userID {
			return "", fmt.Errorf("unauthorized: you can only access files for your own cases")
		}
	} else if userRole == "lawyer" {

		acceptedQuote, err := s.repo.GetAcceptedQuoteByCaseID(ctx, caseRecord.ID)
		if err != nil || acceptedQuote == nil {
			return "", fmt.Errorf("unauthorized: you must have an accepted quote to access files")
		}
		if acceptedQuote.LawyerID != userID {
			return "", fmt.Errorf("unauthorized: you can only access files for cases where your quote was accepted")
		}
		if caseRecord.Status != "engaged" {
			return "", fmt.Errorf("unauthorized: case must be engaged to access files")
		}
	} else {
		return "", fmt.Errorf("unauthorized")
	}

	request, err := s.presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.config.StorageBucket),
		Key:    aws.String(fileRecord.FilePath),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = time.Duration(1 * time.Hour)
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate signed URL: %w", err)
	}

	return request.URL, nil
}

func generateSecureFilename(originalFilename string) string {

	b := make([]byte, 16)
	rand.Read(b)
	randomHex := hex.EncodeToString(b)

	ext := filepath.Ext(originalFilename)

	return fmt.Sprintf("%s_%d%s", randomHex, time.Now().Unix(), ext)
}
