package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hellhium/wstest/datastore"
	"github.com/hellhium/wstest/lib/httphelpers"
	"github.com/hellhium/wstest/lib/router"
	"github.com/hellhium/wstest/lib/router/routebuilders"
	"github.com/julienschmidt/httprouter"
)

func init() {
	router.Router.POST("/api/:type", routebuilders.Authed(addOneWithType))
}

func addOneWithType(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
	typeName := params.ByName("type")

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
		nextID := uint64(len(dstype))
		nextIDS := fmt.Sprintf("%d", nextID)
		if _, exist := dstype[nextIDS]; exist {
			httphelpers.GenericInvalidParam.D("Add non incremental").Quick(resp)
			return
		}

		dstype[nextIDS] = data
		out.Data = nextID
	} else {
		datastore.DS.Data[typeName] = map[string]map[string]interface{}{
			"0": data,
		}
		out.Data = 0
	}
	datastore.DS.Save()
	out.R(resp)
}
