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
	router.Router.GET("/api/:type/:id", routebuilders.Authed(getOneWithType))
}

func getOneWithType(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
	typeName := params.ByName("type")
	idName := params.ByName("id")

	if _, ok := datastore.DS.Data[typeName]; !ok {
		httphelpers.GenericNotFound.D("type not found").Quick(resp)
		return
	} else {
		if doc, ok := datastore.DS.Data[typeName][idName]; !ok {
			httphelpers.GenericNotFound.D("id not found").Quick(resp)
			return
		} else {
			out := httphelpers.NewResp()
			out.Success = true
			out.Data = doc
			out.R(resp)
		}
	}
}
