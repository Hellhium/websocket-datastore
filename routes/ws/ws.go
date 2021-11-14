package ws

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/hellhium/wstest/datastore"
	"github.com/hellhium/wstest/lib/tlib"
)

type loginMessage struct {
	User  string `json:"user"`
	Pass  string `json:"pass"`
	Debug string `json:"debug"`
}

type wsRequest struct {
	ReqID    int                    `json:"reqid"`
	ReqType  string                 `json:"reqtype"`  // GET / SET / GETALL / ADD / DEL
	DataType string                 `json:"datatype"` // User / item / ....
	ID       interface{}            `json:"id"`
	Data     map[string]interface{} `json:"data"`
}

type wsResponse struct {
	ReqID        int                    `json:"reqid"`
	LastInsertID uint64                 `json:"lastinsertid"`
	Data         map[string]interface{} `json:"data"`
	Error        string                 `json:"error"`
	Success      bool                   `json:"success"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func wsApi(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		tlib.Error("msg", err)
		return
	}
	defer c.Close()
	debug := 0

	{
		_, message, err := c.ReadMessage()
		if err != nil {
			tlib.Error("msg", err)
			return
		}
		loginReq := loginMessage{}
		json.Unmarshal([]byte(message), &loginReq)

		if tlib.MatchUser(loginReq.User, loginReq.Pass) {
			tlib.Error("msg", "login failed", "addr", r.RemoteAddr)
			return
		}

		switch loginReq.Debug {
		case "full":
			debug = 2
		case "verbose":
			debug = 1
		}
		if debug > 0 {
			tlib.Info("msg", "new session with debug", "level", debug)
		}
	}

	for {
		mt, message, err := c.ReadMessage()
		if debug > 1 {
			tlib.Info("msg", fmt.Sprintf("REQU %d %s", mt, message))
		}

		if err != nil {
			tlib.Info("msg", "closed", "err", err)
			return
		}

		wsReq := &wsRequest{}
		err = json.Unmarshal(message, wsReq)
		if err != nil {
			tlib.Error("err", err)
			return
		}

		var wsResp *wsResponse
		switch wsReq.ReqType {
		case "GET":
			wsResp = opGet(wsReq)
		case "SET":
			wsResp = opSet(wsReq)
		case "GETALL":
			wsResp = opGetAll(wsReq)
		case "ADD":
			wsResp = opAdd(wsReq)
		case "DEL":
			wsResp = opDel(wsReq)
		}

		data, _ := json.Marshal(wsResp)
		if debug > 1 {
			tlib.Info("msg", fmt.Sprintf("RESP %d %s", mt, data))
		}
		err = c.WriteMessage(mt, data)
		if err != nil {
			tlib.Error("err", err)
			return
		}
	}
}

func opGet(req *wsRequest) (resp *wsResponse) {
	resp = &wsResponse{
		ReqID: req.ReqID,
	}

	datastore.DS.OpsM.Lock()
	defer datastore.DS.OpsM.Unlock()

	var id string
	switch dta := req.ID.(type) {
	case string:
		id = dta
	case float64:
		id = fmt.Sprintf("%f", dta)
	default:
		resp.Error = "Invalid id type"
		return
	}

	if dstype, ok := datastore.DS.Data[req.DataType]; ok {
		if data, ok := dstype[id]; ok {
			resp.Data = data
			resp.Success = true
		} else {
			resp.Error = "ID not found"
		}
	} else {
		resp.Error = "Datatype not found"
	}

	return
}

func opGetAll(req *wsRequest) (resp *wsResponse) {
	resp = &wsResponse{
		ReqID: req.ReqID,
	}

	datastore.DS.OpsM.Lock()
	defer datastore.DS.OpsM.Unlock()

	if dstype, ok := datastore.DS.Data[req.DataType]; ok {
		resp.Data = map[string]interface{}{
			"list":  dstype,
			"count": len(dstype),
		}
		resp.Success = true
	} else {
		resp.Error = "Datatype not found"
	}

	return
}

func opSet(req *wsRequest) (resp *wsResponse) {
	resp = &wsResponse{
		ReqID: req.ReqID,
	}

	datastore.DS.OpsM.Lock()
	defer datastore.DS.OpsM.Unlock()
	defer func() {
		if resp.Success {
			datastore.DS.Save()
		}
	}()

	var id string
	switch dta := req.ID.(type) {
	case string:
		id = dta
	case float64:
		id = fmt.Sprintf("%f", dta)
	default:
		resp.Error = "Invalid id type"
		return
	}

	if dstype, ok := datastore.DS.Data[req.DataType]; ok {
		dstype[id] = req.Data
	} else {
		datastore.DS.Data[req.DataType] = map[string]map[string]interface{}{
			id: req.Data,
		}
	}

	resp.Success = true

	return
}

func opAdd(req *wsRequest) (resp *wsResponse) {
	resp = &wsResponse{
		ReqID: req.ReqID,
	}

	datastore.DS.OpsM.Lock()
	defer datastore.DS.OpsM.Unlock()
	defer func() {
		if resp.Success {
			datastore.DS.Save()
		}
	}()

	if dstype, ok := datastore.DS.Data[req.DataType]; ok {
		nextID := uint64(len(dstype))
		nextIDS := fmt.Sprintf("%d", nextID)
		if _, exist := dstype[nextIDS]; exist {
			resp.Error = "Add non incremental"
			return
		}

		dstype[nextIDS] = req.Data
		resp.LastInsertID = nextID
		resp.Success = true
	} else {
		datastore.DS.Data[req.DataType] = map[string]map[string]interface{}{
			"0": req.Data,
		}
		resp.Success = true
	}

	return
}

func opDel(req *wsRequest) (resp *wsResponse) {
	resp = &wsResponse{
		ReqID: req.ReqID,
	}

	datastore.DS.OpsM.Lock()
	defer datastore.DS.OpsM.Unlock()
	defer func() {
		if resp.Success {
			datastore.DS.Save()
		}
	}()

	var id string
	switch dta := req.ID.(type) {
	case string:
		id = dta
	case float64:
		id = fmt.Sprintf("%f", dta)
	default:
		resp.Error = "Invalid id type"
		return
	}

	if dstype, ok := datastore.DS.Data[req.DataType]; ok {
		if _, ok := dstype[id]; ok {
			delete(dstype, id)
			resp.Success = true
		} else {
			resp.Error = "ID not found"
		}
	} else {
		resp.Error = "Datatype not found"
	}

	return
}
