package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}
var addr = flag.String("addr", "0.0.0.0:8080", "http service address")
var datastorePath = flag.String("dspath", "data/ds.json", "Datastore path")
var ds = dataStore{}

func main() {
	ds.Load()
	log.Println("Datastore loaded")
	flag.Parse()
	http.HandleFunc("/ws", echo)
	log.Printf("WS Listening on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

type dataStore struct {
	data  map[string]map[int64]map[string]interface{}
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
			ds.data = map[string]map[int64]map[string]interface{}{}
			return
		}
		log.Fatal(err)
	}
	err = json.NewDecoder(f).Decode(&ds.data)
	if err != nil {
		log.Fatal(err)
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
	User string `json:"user"`
	Pass string `json:"pass"`
}

type wsRequest struct {
	ReqID    int                    `json:"reqid"`
	ReqType  string                 `json:"reqtype"`  // GET / SET / GETALL / ADD / DEL
	DataType string                 `json:"datatype"` // User / item / ....
	ID       int64                  `json:"id"`
	Data     map[string]interface{} `json:"data"`
}

type wsResponse struct {
	ReqID        int         `json:"reqid"`
	LastInsertID int64       `json:"lastinsertid"`
	Data         interface{} `json:"data"`
	Error        string      `json:"error"`
	Success      bool        `json:"success"`
}

func echo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Upgrade: %s", err)
		return
	}
	defer c.Close()

	{
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			return
		}
		loginReq := loginMessage{}
		json.Unmarshal([]byte(message), &loginReq)
		if loginReq.Pass != "api" || loginReq.User != "api" {
			log.Println("Invalid login")
			return
		}
	}

	for {
		mt, message, err := c.ReadMessage()
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

	if dstype, ok := ds.data[req.DataType]; ok {
		if data, ok := dstype[req.ID]; ok {
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
		resp.Data = dstype
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

	if dstype, ok := ds.data[req.DataType]; ok {
		dstype[req.ID] = req.Data
	} else {
		ds.data[req.DataType] = map[int64]map[string]interface{}{
			req.ID: req.Data,
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
		nextID := int64(len(dstype))
		if _, exist := dstype[nextID]; exist {
			resp.Error = "Add non incremental"
			return
		}

		dstype[nextID] = req.Data
		resp.LastInsertID = nextID
		resp.Success = true
	} else {
		ds.data[req.DataType] = map[int64]map[string]interface{}{
			req.ID: req.Data,
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

	if dstype, ok := ds.data[req.DataType]; ok {
		if _, ok := dstype[req.ID]; ok {
			delete(dstype, req.ID)
			resp.Success = true
		} else {
			resp.Error = "ID not found"
		}
	} else {
		resp.Error = "Datatype not found"
	}

	return
}
