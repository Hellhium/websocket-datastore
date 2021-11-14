package routebuilders

import (
	"net/http"
	"strings"

	"github.com/hellhium/wstest/lib/tlib"
	"github.com/julienschmidt/httprouter"
)

// NonAuthed grants access to route inconditionally
var NonAuthed = handler

func handler(cb func(http.ResponseWriter, *http.Request, httprouter.Params)) func(http.ResponseWriter, *http.Request, httprouter.Params) {
	return func(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
		if !strings.HasPrefix(req.URL.String(), "/prerend") {
			tlib.Debug("path", req.URL, "method", req.Method)
		}
		resp.Header().Set("Allow", "OPTIONS, GET, POST, PUT, DELETE")
		origin := req.Header.Get("origin")
		if origin == "" {
			origin = "http://localhost"
		}
		resp.Header().Set("Access-Control-Allow-Origin", origin)
		resp.Header().Set("Access-Control-Allow-Credentials", "true")
		resp.Header().Set("Access-Control-Allow-Headers", req.Header.Get("Access-Control-Request-Headers"))
		cb(resp, req, params)
	}
}

// BasicHandler is a wrapper for basic go routes provided by other packages
func BasicHandler(cb func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request, httprouter.Params) {
	return handler(func(resp http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		cb(resp, req)
	})
}
