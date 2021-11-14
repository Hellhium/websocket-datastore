package httphelpers

import "net/http"

type HTTPErr struct {
	Err        string `json:"msg"`
	Code       uint   `json:"code"`
	HTTPCode   int    `json:"-"`
	Additional string `json:"additional"`
}

// 1000 = Generic
// 2000 = User

var (
	Generic               = HTTPErr{"Error", 1000, 500, ""}
	GenericNotFound       = HTTPErr{"NotFound", 1001, 404, ""}
	GenericJSONErr        = HTTPErr{"JSONErr", 1002, 400, ""}
	GenericNotImplemented = HTTPErr{"NotImplemented", 1003, 501, ""}
	GenericInvalidParam   = HTTPErr{"InvalidParam", 1004, 400, ""}
	GenericSQLError       = HTTPErr{"SQLError", 1005, 500, ""}
	GenericAntiSpam       = HTTPErr{"GenericAntiSpam", 1006, 400, ""}
	GenericEmailError     = HTTPErr{"GenericEmail", 1007, 500, ""}
)

// D returns
func (hte HTTPErr) D(add string) HTTPErr {
	hte.Additional = add
	return hte
}

func (hte HTTPErr) Quick(resp http.ResponseWriter) {
	resp.WriteHeader(hte.HTTPCode)
	r := NewResp()
	r.Error = &hte
	r.R(resp)
}
