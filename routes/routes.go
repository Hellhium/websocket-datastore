package routes

import (
	"encoding/json"
	"net/http"

	"github.com/hellhium/wstest/lib/router"
	"github.com/hellhium/wstest/lib/router/routebuilders"
	"github.com/hellhium/wstest/lib/tlib"
	"github.com/julienschmidt/httprouter"

	_ "github.com/hellhium/wstest/routes/api" // API routes
	_ "github.com/hellhium/wstest/routes/ws"  // Websocket routes
)

func init() {
	router.Router.GET("/", routebuilders.NonAuthed(r))
}

func r(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
	json.NewEncoder(resp).Encode(map[string]interface{}{
		"success": true,
		"version": tlib.Ver(),
	})
}
