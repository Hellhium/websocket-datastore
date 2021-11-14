package router

import (
	"net/http"
	"runtime/debug"

	"github.com/hellhium/wstest/lib/httphelpers"
	"github.com/hellhium/wstest/lib/tlib"
	"github.com/julienschmidt/httprouter"
)

// Router is the default exposed router
var Router *httprouter.Router

func init() {
	Router = httprouter.New()
	Router.OPTIONS("/*path", options)
	//Router.PanicHandler = panicHandler
}

func panicHandler(w http.ResponseWriter, req *http.Request, _ interface{}) {
	tlib.Error("msg", "caught panic", "path", req.URL.Path, "method", req.Method)
	debug.PrintStack()
	httphelpers.Generic.Quick(w)
}

func options(w http.ResponseWriter, req *http.Request, p httprouter.Params) {
	tlib.Debug("type", "preflight", "url", req.URL)
	w.Header().Set("Allow", "GET, POST, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Methods", "OPTIONS, GET, POST, PUT, DELETE")
	origin := req.Header.Get("origin")
	if origin == "" {
		origin = "http://localhost"
	}
	w.Header().Set("Access-Control-Allow-Origin", req.Header.Get("origin"))
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Headers", req.Header.Get("Access-Control-Request-Headers"))
}
