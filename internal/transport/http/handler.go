package httptransport

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"async-file-storage/internal/usecase"
)

type Handler struct {
	service *usecase.Service
}

func NewHandler(service *usecase.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.Trim(r.URL.Path, "/")
	parts := strings.Split(path, "/")

	if len(parts) == 1 && parts[0] == "downloads" {
		if r.Method == http.MethodPost {
			h.handleCreate(w, r)
			return
		}
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed")
		return
	}

	if len(parts) == 2 && parts[0] == "downloads" {
		if r.Method == http.MethodGet {
			h.handleGet(w, r, parts[1])
			return
		}
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed")
		return
	}

	if len(parts) == 4 && parts[0] == "downloads" && parts[2] == "files" {
		if r.Method == http.MethodGet {
			h.handleGetFile(w, r, parts[1], parts[3])
			return
		}
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed")
		return
	}

	writeError(w, http.StatusNotFound, "NOT_FOUND", "route not found")
}

func (h *Handler) handleCreate(w http.ResponseWriter, r *http.Request) {
	var body createRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "invalid json")
		return
	}

	urls := make([]string, 0, len(body.Files))
	for _, f := range body.Files {
		urls = append(urls, f.URL)
	}

	timeout, err := time.ParseDuration(body.Timeout)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_TIMEOUT", "invalid timeout")
		return
	}

	out, err := h.service.CreateRequest(r.Context(), usecase.CreateRequestInput{
		URLs:    urls,
		Timeout: timeout,
	})
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, createResponse{ID: out.ID, Status: string(out.Status)})
}

func (h *Handler) handleGet(w http.ResponseWriter, r *http.Request, idValue string) {
	id, err := strconv.Atoi(idValue)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid id")
		return
	}

	out, err := h.service.GetRequest(r.Context(), id)
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	resp := getRequestResponse{ID: out.ID, Status: string(out.Status)}
	resp.Files = make([]fileOutcome, 0, len(out.Files))
	for _, f := range out.Files {
		item := fileOutcome{URL: f.URL}
		if f.ErrorCode != "" {
			item.Error = &errorInfo{Code: f.ErrorCode}
		} else {
			item.ID = f.FileID
		}
		resp.Files = append(resp.Files, item)
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) handleGetFile(w http.ResponseWriter, r *http.Request, requestIDValue string, fileIDValue string) {
	requestID, err := strconv.Atoi(requestIDValue)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid request id")
		return
	}
	fileID, err := strconv.Atoi(fileIDValue)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid file id")
		return
	}

	out, err := h.service.GetFile(r.Context(), requestID, fileID)
	if err != nil {
		if bErr := (usecase.BusinessError{}); errors.As(err, &bErr) {
			writeJSON(w, http.StatusOK, errorResponse{Error: errorInfo{Code: bErr.Code, Message: bErr.Msg}})
			return
		}
		writeUsecaseError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(out.Data)
}

func writeUsecaseError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, usecase.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, "INVALID_INPUT", err.Error())
	case errors.Is(err, usecase.ErrNotFound):
		writeError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal error")
	}
}

func writeError(w http.ResponseWriter, status int, code string, msg string) {
	writeJSON(w, status, errorResponse{Error: errorInfo{Code: code, Message: msg}})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	_ = enc.Encode(payload)
}
