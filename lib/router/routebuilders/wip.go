package routebuilders

import (
	"net/http"

	"github.com/hellhium/wstest/lib/httphelpers"
	"github.com/julienschmidt/httprouter"
)

func TODO(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
	httphelpers.GenericNotImplemented.Quick(resp)
	return
}
