package tlib

import (
	"os"
	"path/filepath"

	"gopkg.in/alecthomas/kingpin.v2"
)

// KP is the exported kingpin app, use it before calling init
var KP = kingpin.New(filepath.Base(os.Args[0]), "Websocket Datastore")

var config = struct {
	logLevel string
	env      string
	user     string
	password string
}{}

var BaseURL = ""

// Config is the exported config struct
var Config struct {
	Prod bool
}

func init() {
	KP.Author("Jemy SCHNEPP")
	KP.Version(Ver())
	KP.HelpFlag.Short('h')

	KP.Flag("log.level", "Log level (all, debug, info, warn, error)").Envar("LOG_LEVEL").Default("info").StringVar(&config.logLevel)

	KP.Flag("app.baseurl", "App base url").Envar("BASE_URL").Default("").StringVar(&BaseURL)

	KP.Flag("app.user", "App user").Envar("APP_USER").Default("api").StringVar(&config.user)
	KP.Flag("app.password", "App password").Envar("APP_PASSWORD").Default("api").StringVar(&config.password)

	KP.Flag("env", "Environment").Envar("env").Default("prod").StringVar(&config.env)
}

func MatchUser(user, password string) bool {
	return user == config.user && password == config.password
}
