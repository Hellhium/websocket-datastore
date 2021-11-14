package routebuilders

import (
	"net/http"

	"github.com/hellhium/wstest/lib/tlib"
	"github.com/julienschmidt/httprouter"
)

// RouteAuthed represents a route for an user
type RouteAuthed func(http.ResponseWriter, *http.Request, httprouter.Params)

// RouteGeneric represents an open route
type RouteGeneric func(http.ResponseWriter, *http.Request, httprouter.Params)

type routeMaybeAuthed func(http.ResponseWriter, *http.Request, httprouter.Params, bool)

func isAuthed(cb routeMaybeAuthed) httprouter.Handle {
	return handler(func(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
		username, password, ok := req.BasicAuth()
		if ok && tlib.MatchUser(username, password) {
			cb(resp, req, params, true)
			return
		}

		cb(resp, req, params, false)
	})
}

// Authed executes cb if user is authed
func Authed(cb RouteAuthed) httprouter.Handle {
	return isAuthed(func(resp http.ResponseWriter, req *http.Request, params httprouter.Params, authed bool) {
		if !authed {
			rejectionHandler(resp, req, params)
			return
		}
		cb(resp, req, params)
		return
	})
}

func rejectionHandler(resp http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	resp.Header().Set("WWW-Authenticate", `Basic realm="Authorization Required"`)
	resp.WriteHeader(401)
	return
}
