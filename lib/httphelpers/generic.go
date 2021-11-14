package httphelpers

import (
	"encoding/json"
	"net/http"
	"sync"
)

type httpResponse struct {
	Success  bool        `json:"success"`
	Error    *HTTPErr    `json:"error,omitempty"`
	Data     interface{} `json:"data,omitempty"`
	DataType string      `json:"data_type,omitempty"`

	// Results
	ResSkip    int   `json:"skip,omitempty"`
	ResCount   int   `json:"count,omitempty"`
	ResTotal   int   `json:"total,omitempty"`
	ResHasNext *bool `json:"next,omitempty"`
}

var respPool = sync.Pool{
	New: func() interface{} {
		return &httpResponse{}
	},
}

func NewResp() *httpResponse {
	return respPool.Get().(*httpResponse)
}

func (h *httpResponse) HasNext(b bool) {
	h.ResHasNext = &b
}

func (h *httpResponse) R(resp http.ResponseWriter) {
	jse := json.NewEncoder(resp)
	jse.SetEscapeHTML(false)
	jse.SetIndent("", "  ")
	jse.Encode(h)
}
