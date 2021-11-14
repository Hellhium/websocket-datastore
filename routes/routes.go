package routes

import (
	"net/http"

	"github.com/hellhium/wstest/lib/httphelpers"
	"github.com/hellhium/wstest/lib/router"
	"github.com/hellhium/wstest/lib/router/routebuilders"
	"github.com/hellhium/wstest/lib/tlib"
	"github.com/julienschmidt/httprouter"

	_ "embed"

	_ "github.com/hellhium/wstest/routes/api" // API routes
	_ "github.com/hellhium/wstest/routes/ws"  // Websocket routes
)

//go:embed swagger.html
var swaggerHtml []byte

//go:embed swagger.yml
var swaggerYml []byte

func init() {
	router.Router.GET("/status", routebuilders.NonAuthed(r))
	router.Router.GET("/", routebuilders.NonAuthed(apidoc))
	router.Router.GET("/swagger.yml", routebuilders.NonAuthed(apidocyml))
}

func apidoc(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
	resp.Header().Add("content-type", "text/html")
	resp.Write(swaggerHtml)
}

func apidocyml(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
	resp.Header().Add("content-type", "text/yaml")
	resp.Write(swaggerYml)

}

func r(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
	res := httphelpers.NewResp()
	res.Data = map[string]interface{}{
		"version": tlib.Ver(),
	}
	res.Success = true
	res.R(resp)
}
