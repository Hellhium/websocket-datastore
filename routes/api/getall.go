package api

import (
	"net/http"

	"github.com/hellhium/wstest/datastore"
	"github.com/hellhium/wstest/lib/httphelpers"
	"github.com/hellhium/wstest/lib/router"
	"github.com/hellhium/wstest/lib/router/routebuilders"
	"github.com/julienschmidt/httprouter"
)

func init() {
	router.Router.GET("/api", routebuilders.Authed(getAll))
}

func getAll(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
	out := httphelpers.NewResp()
	out.Data = datastore.DS.Data
	out.Success = true
	out.R(resp)
}
