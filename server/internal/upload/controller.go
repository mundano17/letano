package upload

import (
	"database/sql"
	"encoding/hex"
	"errors"
	"log"
	"net/http"
	"regexp"
	"strings"
	"tano/http/middleware"
	db "tano/internal/db/generated"
	"tano/internal/utils"
)

type UploadController struct {
	service *uploadService
}

type Field struct {
	FieldKey   string `json:"fieldkey"`
	FieldValue string `json:"fieldvalue"`
}

type NewUploadRequest struct {
	Filehash string  `json:"filehash"`
	Filename string  `json:"filename"`
	Filesize int64   `json:"filesize"`
	Mimetype string  `json:"mimetype"`
	Fields   []Field `json:"fields"`
}

var validFilename = regexp.MustCompile(
	`^[a-zA-Z0-9][a-zA-Z0-9._ ()[\]+,-]{0,254}$`,
)

func (f *NewUploadRequest) validate() error {
	if !validFilename.MatchString(f.Filename) ||
		f.Filename == "." ||
		f.Filename == ".." {
		return errors.New("invalid filename")
	}

	if f.Filesize <= 0 {
		return errors.New("invalid filesize")
	}

	if len(f.Filehash) != 64 {
		return errors.New("invalid hash")
	}

	if _, err := hex.DecodeString(f.Filehash); err != nil {
		return errors.New("invalid hash")
	}

	for _, field := range f.Fields {

		field.FieldKey = strings.TrimSpace(field.FieldKey)
		field.FieldValue = strings.TrimSpace(field.FieldValue)
		if field.FieldKey == "" ||
			field.FieldValue == "" {
			return errors.New("invalid fields")
		}
	}

	return nil
}

type NewUploadRes struct {
	UploadID string `json:"uploadID"`
}

func NewUploadController(query *db.Queries, db *sql.DB) *UploadController {
	return &UploadController{service: NewUploadService(query, db)}
}

func (u *UploadController) InitializeUploadController(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	req := utils.GetJson[NewUploadRequest](w, r)
	if req == nil {
		return
	}

	if err := req.validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	uploadID, err := u.service.initializeUploadService(r.Context(), userID, *req)
	if err != nil {
		if errors.Is(err, ErrInvalidRequest) {
			http.Error(w, "Invalid Request", http.StatusBadRequest)
			return
		}
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, NewUploadRes{UploadID: uploadID})
}

type ChunkRequest struct {
	UploadID string `json:"uploadID"`
	Start    int64  `json:"start"`
	Length   int64  `json:"length"`
	Data     []byte `json:"data"`
	FileSize int64  `json:"filesize"`
}

func (c ChunkRequest) Validate() bool {
	if c.FileSize < 0 ||
		c.Start < 0 ||
		c.Length <= 0 ||
		c.Start >= c.FileSize ||
		c.Length > c.FileSize-c.Start ||
		c.Length != int64(len(c.Data)) ||
		c.Start%chunkSize != 0 {
		return false
	}

	expectedLength := min(chunkSize, c.FileSize-c.Start)
	return c.Length == expectedLength
}

func (u *UploadController) AddChunk(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	req := utils.GetJson[ChunkRequest](w, r)
	if req == nil {
		return
	}

	if ok := req.Validate(); !ok {
		http.Error(w, "Invalid Chunk Details", http.StatusBadRequest)
		return
	}

	err = u.service.addChunkService(r.Context(), userID, *req)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Invalid Upload ID", http.StatusBadRequest)
			return
		}
		if errors.Is(err, ErrInvalidStatus) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (u *UploadController) FinishUpload(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	uploadID := r.PathValue("uploadID")
	if uploadID == "" {
		http.Error(w, "Upload ID not found", http.StatusBadRequest)
		return
	}

	err = u.service.finishUploadService(r.Context(), userID, uploadID)
	if err != nil {
		if errors.Is(err, ErrInvalidStatus) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if errors.Is(err, ErrCorruptedData) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Invalid Upload ID", http.StatusBadRequest)
			return
		}
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)

}

type UploadStatusRes struct {
	uploadID       string
	FileHash       string
	FileName       string
	FileSize       int64
	UploadedChunks []byte
	MimeType       string
	UploadStatus   string
	Fields         string
	CreatedAT      string
}

func (u *UploadController) GetStatusController(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	uploadID := r.PathValue("uploadID")
	if uploadID == "" {
		http.Error(w, "Upload ID not found", http.StatusBadRequest)
		return
	}

	status, err := u.service.getStatus(r.Context(), userID, uploadID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Invalid Upload ID", http.StatusBadRequest)
			return
		}
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	utils.WriteJSON(w, http.StatusOK, status)
}

func (u *UploadController) ResumeUploadController(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	uploadID := r.PathValue("uploadID")
	if uploadID == "" {
		http.Error(w, "Upload ID not found", http.StatusBadRequest)
		return
	}

	if err := u.service.resumeUpload(r.Context(), userID, uploadID); err != nil {
		if errors.Is(err, ErrUploadIsCancelledOrFinished) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func (u *UploadController) PauseUploadController(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	uploadID := r.PathValue("uploadID")
	if uploadID == "" {
		http.Error(w, "Upload ID not found", http.StatusBadRequest)
		return
	}

	if err := u.service.pauseUpload(r.Context(), userID, uploadID); err != nil {
		if errors.Is(err, ErrUploadIsCancelledOrFinished) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func (u *UploadController) CancelUploadController(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	uploadID := r.PathValue("uploadID")
	if uploadID == "" {
		http.Error(w, "Upload ID not found", http.StatusBadRequest)
		return
	}

	if err := u.service.cancelUpload(r.Context(), userID, uploadID); err != nil {
		if errors.Is(err, ErrUploadIsCancelledOrFinished) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
