package main

import (
	"bytes"
	"net/http"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
	"github.com/pilu/fresh/runner/runnerutils"

	"github.com/phaikawl/lovepress/client"
	"github.com/phaikawl/lovepress/model"
	"github.com/phaikawl/lovepress/server/dblayer"
	"github.com/phaikawl/lovepress/server/services/db"
	"github.com/phaikawl/wade"
	wadess "github.com/phaikawl/wade/rbackend/serverside"
)

const (
	MySigningKey   = "n0t9r34t6cz9r34tn0t1na9r34tw4y"
	ConfigFilePath = "data/config.toml"
	ServerError    = http.StatusInternalServerError
	BadRequest     = http.StatusBadRequest
)

var (
	g *Environment = environment()
)

type Environment struct {
	devMode  bool
	basePath string
}

func environment() *Environment {
	return &Environment{
		devMode:  true,
		basePath: "/",
	}
}

func (g *Environment) IsDevMode() bool {
	return g.devMode
}

func (g *Environment) BasePath() string {
	return g.basePath
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

type Context struct {
	*gin.Context

	user *model.User
}

func NewContext(c *gin.Context) *Context {
	return &Context{
		Context: c,
	}
}

func (ctx *Context) ReportError(err error, httpStatus int) {
	if g.IsDevMode() {
		ctx.Fail(httpStatus, err)
	} else {
		glog.Error(err.Error())
		ctx.Fail(httpStatus, nil)
	}
}

func (ctx *Context) CheckError(err error, httpStatus int) {
	if err != nil {
		ctx.ReportError(err, httpStatus)
	}
}

func (ctx *Context) Response(data interface{}) {
	ctx.JSON(http.StatusOK, data)
}

type RouterGroup struct {
	*gin.RouterGroup
}

type HandlerFunc func(*Context)

func (r RouterGroup) toGinList(handlers []HandlerFunc) []gin.HandlerFunc {
	list := make([]gin.HandlerFunc, len(handlers))
	for i, fn := range handlers {
		list[i] = func(c *gin.Context) {
			fn(NewContext(c))
		}
	}

	return list
}

func (r RouterGroup) Group(route string, handler HandlerFunc) RouterGroup {
	return RouterGroup{r.RouterGroup.Group(route, func(c *gin.Context) {
		if handler != nil {
			handler(NewContext(c))
		}
	})}
}

func (r RouterGroup) Use(handlers ...HandlerFunc) {
	r.RouterGroup.Use(r.toGinList(handlers)...)
}

func (r RouterGroup) GET(route string, handlers ...HandlerFunc) {
	r.RouterGroup.GET(route, r.toGinList(handlers)...)
}

func (r RouterGroup) POST(route string, handlers ...HandlerFunc) {
	r.RouterGroup.POST(route, r.toGinList(handlers)...)
}

type ApiSystem struct {
	Public    RouterGroup
	Protected RouterGroup
	Dev       RouterGroup
}

type ApiProvider interface {
	ProvideApi(apis ApiSystem)
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

	protected := api.Group("/protected/", func(c *gin.Context) {})

	devapi := api.Group("/dev", func(c *gin.Context) {
		if !g.IsDevMode() {
			c.Abort(http.StatusNotFound)
		}
	})

	apiProviders := []ApiProvider{
		AuthApi{
			jwtSigningKey: MySigningKey,
		},
	}

	apis := ApiSystem{
		Public:    RouterGroup{api},
		Protected: RouterGroup{protected},
		Dev:       RouterGroup{devapi},
	}

	for _, provider := range apiProviders {
		provider.ProvideApi(apis)
	}

	apis.Protected.GET("/test", func(c *Context) {
		c.Response("OK Baby")
	})

	// Subpaths of /web/ are client urls, should NOT be protected
	// it renders the page
	web := r.Group("/web/", func(c *gin.Context) {
		ctx := NewContext(c)
		f, err := os.Open("../public/index.html")
		ctx.CheckError(err, http.StatusInternalServerError)
		buf := bytes.NewBufferString("")

		ctx.CheckError(wadess.RenderApp(buf, wade.AppConfig{
			StartPage: "pg-home",
			BasePath:  "/web",
		}, client.InitFunc, f, r, c.Request), http.StatusInternalServerError)

		c.Data(200, "text/html", buf.Bytes())
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
