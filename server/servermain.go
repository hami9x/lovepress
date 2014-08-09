package main

import (
	"io/ioutil"
	"net/http"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
	"github.com/pilu/fresh/runner/runnerutils"

	"github.com/phaikawl/lovepress/server/dblayer"
	"github.com/phaikawl/lovepress/server/services/db"
)

const (
	MySigningKey   = "n0t9r34t6cz9r34tn0t1na9r34tw4y"
	ConfigFilePath = "data/config.toml"
)

var (
	g *Environment = environment()
)

type Environment struct {
	devMode bool
}

func environment() *Environment {
	return &Environment{
		devMode: true,
	}
}

func (g *Environment) IsDevMode() bool {
	return g.devMode
}

func reportError(err error, c *gin.Context, httpStatus int) {
	if g.IsDevMode() {
		c.Fail(httpStatus, err)
	} else {
		glog.Error(err.Error())
		c.Fail(httpStatus, nil)
	}
}

func checkError(err error, c *gin.Context, httpStatus int) {
	if err != nil {
		reportError(err, c, httpStatus)
	}
}

func runnerMiddleware(c *gin.Context) {
	if runnerutils.HasErrors() {
		runnerutils.RenderError(c.Writer)
		c.Abort(500)
	}
}

type Config struct {
	Db db.Config `toml:"database"`
}

func main() {
	r := gin.Default()
	conf := &Config{}
	_, err := toml.DecodeFile(ConfigFilePath, conf)
	if err != nil {
		panic("Cannot load config file. " + err.Error())
	}
	db.Init(conf.Db)
	dblayer.Init()

	// IMPORTANT PART STARTS HERE(=
	//
	if g.IsDevMode() {
		r.Use(runnerMiddleware)
		gopath := os.Getenv("GOPATH")
		if gopath != "" {
			r.ServeFiles("/gopath/*filepath", http.Dir(gopath))
		}
	}

	// This serves static files in the "public" directory
	r.ServeFiles("/public/*filepath", http.Dir("../public"))

	// Subpaths of /api/ provides the server API
	api := r.Group("/api/", func(c *gin.Context) {
	})

	userApi(api)

	// Subpaths of /web/ are client urls, should NOT be protected
	// Just serve the index.html for every subpaths actually, nothing else
	web := r.Group("/web/", func(c *gin.Context) {
		f, err := os.Open("../public/index.html")
		checkError(err, c, http.StatusInternalServerError)
		conts, err := ioutil.ReadAll(f)
		checkError(err, c, http.StatusInternalServerError)
		c.Data(200, "text/html", conts)
	})
	web.GET("*path", func(c *gin.Context) {})

	// Redirect the home page to /web/
	r.GET("/", func(c *gin.Context) {
		http.Redirect(c.Writer, c.Request, "/web/", http.StatusFound)
	})

	//
	// =)IMPORTANT PART ENDS HERE

	r.Run(":3000")
}
