package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

var addr = flag.String("addr", "0.0.0.0:8080", "http service address")
var datastorePath = flag.String("dspath", "data/ds.json", "Datastore path")
var ds = dataStore{}
var username = flag.String("username", "api", "Username")
var password = flag.String("password", "api", "Password")

func main() {
	ds.Load()
	log.Println("Datastore loaded")
	flag.Parse()
	http.HandleFunc("/ws", wsApi)
	http.HandleFunc("/api", getAll)
	log.Printf("WS Listening on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

type dataStore struct {
	data  map[string]map[string]map[string]interface{}
	opsM  sync.Mutex
	saveM sync.Mutex
}

func (ds *dataStore) Load() {
	ds.saveM.Lock()
	defer ds.saveM.Unlock()

	f, err := os.Open(*datastorePath)
	if err != nil {
		if os.IsNotExist(err) {
			if fp := filepath.Dir(*datastorePath); fp != "." {
				os.MkdirAll(fp, 0755)
			}
			ioutil.WriteFile(*datastorePath, []byte("{}"), 0644)
			ds.data = map[string]map[string]map[string]interface{}{}
			return
		}
		log.Fatal(err)
	}
	err = json.NewDecoder(f).Decode(&ds.data)
	f.Close()
	if err != nil {
		err := os.Rename(*datastorePath, *datastorePath+".old"+time.Now().Format("2006-01-02-15-04-05"))
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Datastore invalid, file marked as broken")
		ioutil.WriteFile(*datastorePath, []byte("{}"), 0644)
		ds.data = map[string]map[string]map[string]interface{}{}
	}
}

func (ds *dataStore) Save() {
	f, err := os.OpenFile(*datastorePath, os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	jse := json.NewEncoder(f)
	jse.SetEscapeHTML(false)
	jse.SetIndent("", "  ")
	jse.Encode(ds.data)
	f.Close()
}

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

type BasicAuthFunc func(username, password string) bool

func (f BasicAuthFunc) RequireAuth(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Basic realm="Authorization Required"`)
	w.WriteHeader(401)
}

func (f BasicAuthFunc) Authenticate(r *http.Request) bool {
	username, password, ok := r.BasicAuth()
	return ok && f(username, password)
}

func getAll(w http.ResponseWriter, r *http.Request) {
	f := BasicAuthFunc(func(user, pass string) bool {
		return *username == user && *password == pass
	})

	if !f.Authenticate(r) {
		f.RequireAuth(w)
		return
	}

	jse := json.NewEncoder(w)
	jse.SetEscapeHTML(false)
	jse.SetIndent("", "  ")
	w.Header().Add("content-type", "application/json")
	jse.Encode(ds.data)
}

func wsApi(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Upgrade: %s", err)
		return
	}
	defer c.Close()
	debug := 0

	{
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			return
		}
		loginReq := loginMessage{}
		json.Unmarshal([]byte(message), &loginReq)
		if loginReq.Pass != *password || loginReq.User != *username {
			log.Println("Invalid login")
			log.Printf("%s\n\n%+#v", message, loginReq)
			return
		}
		switch loginReq.Debug {
		case "full":
			debug = 2
		case "verbose":
			debug = 1
		}
		if debug > 0 {
			log.Print("New session with debug", debug)
		}
	}

	for {
		mt, message, err := c.ReadMessage()
		if debug > 1 {
			log.Printf("REQU %d %s", mt, message)
		}

		if err != nil {
			log.Printf("WS closed: %s", err)
			return
		}

		wsReq := &wsRequest{}
		err = json.Unmarshal(message, wsReq)
		if err != nil {
			log.Printf("Json WS: %s", err)
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
			log.Printf("RESP %d %s", mt, data)
		}
		err = c.WriteMessage(mt, data)
		if err != nil {
			log.Println("WS write closed: :", err)
			return
		}
	}
}

func opGet(req *wsRequest) (resp *wsResponse) {
	resp = &wsResponse{
		ReqID: req.ReqID,
	}

	ds.opsM.Lock()
	defer ds.opsM.Unlock()

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

	if dstype, ok := ds.data[req.DataType]; ok {
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

	ds.opsM.Lock()
	defer ds.opsM.Unlock()

	if dstype, ok := ds.data[req.DataType]; ok {
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

	ds.opsM.Lock()
	defer ds.opsM.Unlock()
	defer func() {
		if resp.Success {
			ds.Save()
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

	if dstype, ok := ds.data[req.DataType]; ok {
		dstype[id] = req.Data
	} else {
		ds.data[req.DataType] = map[string]map[string]interface{}{
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

	ds.opsM.Lock()
	defer ds.opsM.Unlock()
	defer func() {
		if resp.Success {
			ds.Save()
		}
	}()

	if dstype, ok := ds.data[req.DataType]; ok {
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
		ds.data[req.DataType] = map[string]map[string]interface{}{
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

	ds.opsM.Lock()
	defer ds.opsM.Unlock()
	defer func() {
		if resp.Success {
			ds.Save()
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

	if dstype, ok := ds.data[req.DataType]; ok {
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
