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
	router.Router.GET("/api/:type", routebuilders.Authed(getAllByType))
}

func getAllByType(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
	typeName := params.ByName("type")
	if data, ok := datastore.DS.Data[typeName]; ok {
		out := httphelpers.NewResp()
		out.Data = data
		out.Success = true
		out.R(resp)
	} else {
		httphelpers.GenericNotFound.D("type not found").Quick(resp)
	}
}
