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
	router.Router.DELETE("/api/:type", routebuilders.Authed(delType))
}

func delType(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
	typeName := params.ByName("type")

	datastore.DS.OpsM.Lock()
	defer datastore.DS.OpsM.Unlock()

	if _, ok := datastore.DS.Data[typeName]; !ok {
		httphelpers.GenericNotFound.D("type not found").Quick(resp)
		return
	} else {
		delete(datastore.DS.Data, typeName)
	}
	datastore.DS.Save()
	out := httphelpers.NewResp()
	out.Success = true
	out.R(resp)
}
