package datastore

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/hellhium/wstest/lib/tlib"
)

// DS is the datastore instance
var DS = dataStore{}

var datastorePath string

func init() {
	tlib.KP.Flag("datastore.path", "Datastore path").Envar("DATA_PATH").Default("data/ds.json").StringVar(&datastorePath)
}

type dataStore struct {
	Data  map[string]map[string]map[string]interface{}
	OpsM  sync.Mutex
	saveM sync.Mutex
}

func (ds *dataStore) Load() {
	ds.saveM.Lock()
	defer ds.saveM.Unlock()

	f, err := os.Open(datastorePath)
	if err != nil {
		if os.IsNotExist(err) {
			if fp := filepath.Dir(datastorePath); fp != "." {
				os.MkdirAll(fp, 0755)
			}
			ioutil.WriteFile(datastorePath, []byte("{}"), 0644)
			ds.Data = map[string]map[string]map[string]interface{}{}
			tlib.Info("msg", "datastore created")
			return
		}
		tlib.FatalIf(err)
	}
	err = json.NewDecoder(f).Decode(&ds.Data)
	f.Close()
	if err != nil {
		errr := os.Rename(datastorePath, datastorePath+".old"+time.Now().Format("2006-01-02-15-04-05"))
		tlib.FatalIf(errr)
		tlib.Error("err", err, "msg", "Datastore invalid, file marked as broken")
		ioutil.WriteFile(datastorePath, []byte("{}"), 0644)
		ds.Data = map[string]map[string]map[string]interface{}{}
		tlib.Info("msg", "new datastore created")
		return
	}
	tlib.Info("msg", "datastore loaded")
}

func (ds *dataStore) Save() {
	f, err := os.OpenFile(datastorePath, os.O_WRONLY, 0644)
	tlib.FatalIf(err)
	jse := json.NewEncoder(f)
	jse.SetEscapeHTML(false)
	jse.SetIndent("", "  ")
	jse.Encode(ds.Data)
	f.Close()
}
