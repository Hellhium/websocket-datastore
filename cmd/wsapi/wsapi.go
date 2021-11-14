package main

import (
	"net/http"
	"os"
	"os/signal"
	"sync"

	"github.com/hellhium/wstest/datastore"
	"github.com/hellhium/wstest/lib/router"
	"github.com/hellhium/wstest/lib/tlib"

	_ "github.com/hellhium/wstest/routes" // Init routes
)

var config = struct {
	listenAddr string
}{}

var appClose sync.WaitGroup

func init() {
	tlib.KP.Flag("listen.addr", "Listen addr for http api").Envar("LISTEN_ADDR").Default(":8080").StringVar(&config.listenAddr)
	tlib.Init()
}

func main() {

	datastore.DS.Load()

	{
		go func() {
			tlib.Info("msg", "Start listening for connections", "address", config.listenAddr)
			err := http.ListenAndServe(config.listenAddr, router.Router)
			if err != http.ErrServerClosed {
				tlib.Fatal("msg", err)
			}
		}()
	}

	{
		appClose.Add(1)
		var sigChan = make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt)

		go func() {
			<-sigChan
			tlib.Info("msg", "Got signal, shutting down")
			appClose.Done()
		}()

		appClose.Wait()

		tlib.Info("msg", "Bye")
	}
}
