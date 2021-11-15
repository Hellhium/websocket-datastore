package api

import (
	"encoding/json"
	"net/http"

	"github.com/hellhium/wstest/datastore"
	"github.com/hellhium/wstest/lib/httphelpers"
	"github.com/hellhium/wstest/lib/router"
	"github.com/hellhium/wstest/lib/router/routebuilders"
	"github.com/julienschmidt/httprouter"
)

func init() {
	router.Router.PUT("/api/:type/:id", routebuilders.Authed(setOneWithType))
}

func setOneWithType(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
	typeName := params.ByName("type")
	idName := params.ByName("id")

	data := map[string]interface{}{}
	err := json.NewDecoder(req.Body).Decode(&data)
	if err != nil {
		httphelpers.GenericJSONErr.Quick(resp)
		return
	}

	datastore.DS.OpsM.Lock()
	defer datastore.DS.OpsM.Unlock()

	out := httphelpers.NewResp()
	out.Success = true
	if dstype, ok := datastore.DS.Data[typeName]; ok {
		dstype[idName] = data
	} else {
		datastore.DS.Data[typeName] = map[string]map[string]interface{}{
			idName: data,
		}
	}
	datastore.DS.Save()
	out.R(resp)
}
