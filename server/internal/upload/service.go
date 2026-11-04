package upload

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"
	db "tano/internal/db/generated"

	"github.com/google/uuid"
)

type uploadService struct {
	query *db.Queries
	db    *sql.DB
	mu    *sync.Mutex
}

func NewUploadService(query *db.Queries, db *sql.DB) *uploadService {
	return &uploadService{
		query: query,
		db:    db,
		mu:    &sync.Mutex{},
	}
}

type uploadStatus string

const (
	Start     uploadStatus = "start"
	Resumed   uploadStatus = "resumed"
	Paused    uploadStatus = "paused"
	Cancelled uploadStatus = "cancelled"
	Finished  uploadStatus = "finished"
)

var (
	ErrInvalidRequest error = errors.New("Invalid Request")
)

var chunkSize int64 = int64(math.Pow(2, 20)) * 10

func (u *uploadService) initializeFile(uploadDir string, hash string, size int64) error {

	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return err
	}

	partPath := filepath.Join(uploadDir, hash+".part")

	file, err := os.OpenFile(
		partPath,
		os.O_CREATE|os.O_EXCL|os.O_WRONLY,
		0644,
	)
	if err != nil {
		return (err)
	}

	defer file.Close()

	if err := file.Truncate(size); err != nil {
		return (err)
	}
	return nil
}

func (u *uploadService) initializeUploadService(ctx context.Context, userID string, req NewUploadRequest) (string, error) {
	totalChunks := (req.Filesize + chunkSize - 1) / chunkSize
	bitmap := make([]byte, (totalChunks+7)/8)

	data, err := json.Marshal(req.Fields)
	if err != nil {
		return "", ErrInvalidRequest
	}

	text := string(data)
	valid := true
	if text == "" {
		valid = false
	}

	tx, err := u.db.Begin()
	if err != nil {
		return "", err
	}

	qtx := u.query.WithTx(tx)
	defer tx.Rollback()

	req.Filehash = strings.ToLower(req.Filehash)
	uploadID, err := qtx.InsertUpload(ctx,
		db.InsertUploadParams{
			UploadID:     uuid.New().String(),
			FileHash:     req.Filehash,
			FileName:     req.Filename,
			FileSize:     req.Filesize,
			MimeType:     req.Mimetype,
			UploadStatus: string(Start),
			Fields: sql.NullString{
				String: text,
				Valid:  valid,
			},
			UploadedChunks: bitmap,
			UserID:         userID,
		},
	)
	if err != nil {
		return "", err
	}

	uploadDir := filepath.Join("tmp", uploadID)
	cleanup := func(err error) (string, error) {
		_ = os.RemoveAll(uploadDir)
		return "", err
	}

	if err := u.initializeFile(uploadDir, req.Filehash, req.Filesize); err != nil {
		return cleanup(err)
	}

	if err := tx.Commit(); err != nil {
		return cleanup(err)
	}

	return uploadID, nil
}

func writeChunk(req ChunkRequest, hash string) error {
	partFilePath := filepath.Join("tmp", req.UploadID, hash+".part")

	partFile, err := os.OpenFile(partFilePath, os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer partFile.Close()

	n, err := partFile.WriteAt(req.Data, req.Start)
	if err != nil {
		return err
	}

	if n != len(req.Data) {
		return io.ErrShortWrite
	}

	return nil
}

var ErrInvalidStatus error = errors.New("Status must be started or restarted")

func (u *uploadService) addChunkService(ctx context.Context, userID string, req ChunkRequest) error {
	if !u.checkStatus(ctx, userID, req.UploadID) {
		return ErrInvalidStatus
	}
	u.mu.Lock()
	defer u.mu.Unlock()

	tx, err := u.db.Begin()
	if err != nil {
		return err
	}

	qtx := u.query.WithTx(tx)
	defer tx.Rollback()

	bitmap, err := qtx.GetUploadedChunksByUploadID(ctx,
		db.GetUploadedChunksByUploadIDParams{
			UploadID: req.UploadID,
			UserID:   userID,
		})
	if err != nil {
		return err
	}

	hash, err := qtx.GetHashByUploadID(ctx,
		db.GetHashByUploadIDParams{
			UploadID: req.UploadID,
			UserID:   userID,
		},
	)
	if err != nil {
		return err
	}

	err = writeChunk(req, hash)
	if err != nil {
		return err
	}

	chunkNumber := (req.Start) / (chunkSize)
	byteIndex := chunkNumber / 8
	bitIndex := chunkNumber % 8
	bitmap[byteIndex] |= 1 << bitIndex

	err = qtx.UpdateUploadedChunksByID(ctx, db.UpdateUploadedChunksByIDParams{
		UploadedChunks: bitmap,
		UploadID:       req.UploadID,
		UserID:         userID,
	})
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

var ErrCorruptedData error = errors.New("Data is corrupted")

func (u *uploadService) insertUploadParams(ctx context.Context, qtx *db.Queries, userID string, uploadID string) (string, error) {
	uploadData, err := qtx.GetUploadDetailsByUploadID(ctx, db.GetUploadDetailsByUploadIDParams{
		UploadID: uploadID,
		UserID:   userID,
	})
	if err != nil {
		return "", err
	}

	path := filepath.Join("tmp", uploadID, uploadData.FileHash+".part")
	if ok := verifyFileHash(path, uploadData.FileHash); !ok {
		return "", ErrCorruptedData
	}

	metadataID, err := qtx.InsertMetadata(ctx, db.InsertMetadataParams{
		MetadataID: uuid.NewString(),
		UserID:     userID,
		FileName:   uploadData.FileName,
		FileHash:   uploadData.FileHash,
		MimeType:   uploadData.MimeType,
	})
	if err != nil {
		return "", err
	}

	fields := new([]Field{})

	if err := json.NewDecoder(strings.NewReader(uploadData.Fields.String)).Decode(fields); err != nil {
		return "", err
	}

	for _, fi := range *fields {
		err := qtx.InsertMetadataField(ctx, db.InsertMetadataFieldParams{
			MetadataID: metadataID,
			FieldName:  fi.FieldKey,
			FieldValue: fi.FieldValue,
		})
		if err != nil {
			return "", err
		}
	}

	return uploadData.FileHash, nil
}

func verifyFileHash(path string, expectedHash string) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return false
	}

	actualHash := hex.EncodeToString(hasher.Sum(nil))
	return actualHash == strings.ToLower(expectedHash)
}

func (u *uploadService) checkStatus(ctx context.Context, userID string, uploadID string) bool {
	status, err := u.query.GetStatusByUploadID(ctx, db.GetStatusByUploadIDParams{
		UploadID: uploadID,
		UserID:   userID,
	})
	return err == nil && (uploadStatus(status) == Resumed || uploadStatus(status) == Start)
}

func (u *uploadService) moveFileToStorage(hash string, uploadID string) error {
	srcPath := filepath.Join("tmp", uploadID, hash+".part")
	destPath := filepath.Join("data", "storage", hash)
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("create storage directory: %w", err)
	}

	alreadyStored := verifyFileHash(destPath, hash)

	if alreadyStored {
		return nil
	}

	if err := os.Rename(srcPath, destPath); err != nil {
		return err
	}

	return nil

}

func (u *uploadService) finishUploadService(ctx context.Context, userID string, uploadID string) error {
	if !u.checkStatus(ctx, userID, uploadID) {
		return ErrInvalidStatus
	}

	u.mu.Lock()
	defer u.mu.Unlock()

	tx, err := u.db.Begin()
	if err != nil {
		return err
	}

	qtx := u.query.WithTx(tx)
	defer tx.Rollback()

	hash, err := u.insertUploadParams(ctx, qtx, userID, uploadID)
	if err != nil {
		return err
	}

	if err := u.moveFileToStorage(hash, uploadID); err != nil {
		return err
	}

	if err := qtx.SetStatusByUploadID(ctx, db.SetStatusByUploadIDParams{
		UserID:       userID,
		UploadID:     uploadID,
		UploadStatus: string(Finished),
	}); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (u *uploadService) getStatus(ctx context.Context, userID string, uploadID string) (UploadStatusRes, error) {
	uploadDetails, err := u.query.GetUploadDetailsByUploadID(ctx, db.GetUploadDetailsByUploadIDParams{
		UserID:   userID,
		UploadID: uploadID,
	})
	if err != nil {
		return UploadStatusRes{}, err
	}
	return UploadStatusRes{
		uploadID:       uploadDetails.UploadID,
		FileHash:       uploadDetails.FileHash,
		FileName:       uploadDetails.FileName,
		FileSize:       uploadDetails.FileSize,
		UploadedChunks: uploadDetails.UploadedChunks,
		MimeType:       uploadDetails.MimeType,
		UploadStatus:   uploadDetails.UploadStatus,
		Fields:         uploadDetails.Fields.String,
		CreatedAT:      uploadDetails.CreatedAt.String,
	}, nil
}

func (u *uploadService) isCancelledOrIsFinished(ctx context.Context, userID string, uploadID string) bool {

	status, err := u.query.GetStatusByUploadID(ctx, db.GetStatusByUploadIDParams{
		UploadID: uploadID,
		UserID:   userID,
	})

	return err != nil && (status == string(Cancelled) || status == string(Finished))
}

var ErrUploadIsCancelledOrFinished = errors.New("status is either cancelled or finsihed")

func (u *uploadService) resumeUpload(ctx context.Context, userID string, uploadID string) error {
	if ok := u.isCancelledOrIsFinished(ctx, userID, uploadID); ok {
		return ErrUploadIsCancelledOrFinished
	}
	return u.query.SetStatusByUploadID(ctx, db.SetStatusByUploadIDParams{
		UploadStatus: string(Resumed),
		UploadID:     uploadID,
		UserID:       userID,
	})
}

func (u *uploadService) pauseUpload(ctx context.Context, userID string, uploadID string) error {
	if ok := u.isCancelledOrIsFinished(ctx, userID, uploadID); ok {
		return ErrUploadIsCancelledOrFinished
	}
	return u.query.SetStatusByUploadID(ctx, db.SetStatusByUploadIDParams{
		UploadStatus: string(Paused),
		UploadID:     uploadID,
		UserID:       userID,
	})
}

func (u *uploadService) cancelUpload(ctx context.Context, userID string, uploadID string) error {
	if ok := u.isCancelledOrIsFinished(ctx, userID, uploadID); ok {
		return ErrUploadIsCancelledOrFinished
	}
	return u.query.SetStatusByUploadID(ctx, db.SetStatusByUploadIDParams{
		UploadStatus: string(Cancelled),
		UploadID:     uploadID,
		UserID:       userID,
	})
}
