package httptransport

type createRequestBody struct {
	Files   []fileInput `json:"files"`
	Timeout string      `json:"timeout"`
}

type fileInput struct {
	URL string `json:"url"`
}

type createResponse struct {
	ID     int    `json:"id"`
	Status string `json:"status"`
}

type getRequestResponse struct {
	ID     int           `json:"id"`
	Status string        `json:"status"`
	Files  []fileOutcome `json:"files"`
}

type fileOutcome struct {
	URL   string     `json:"url"`
	ID    int        `json:"file_id,omitempty"`
	Error *errorInfo `json:"error,omitempty"`
}

type errorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message,omitempty"`
}

type errorResponse struct {
	Error errorInfo `json:"error"`
}
