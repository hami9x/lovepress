package main

import (
	"github.com/gopherjs/gopherjs/js"

	wd "github.com/phaikawl/wade"
	"github.com/phaikawl/wade/elements/menu"
	"github.com/phaikawl/wade/services/http"
	"github.com/phaikawl/wade/services/pdata"
	"github.com/phaikawl/wade/utils"

	"github.com/phaikawl/lovepress/model"
)

var (
	gWade *wd.Wade
)

type LoginView struct {
	utils.Validated
	Data model.UsernamePassword
}

func (r *LoginView) Submit() {
	go func() {
		resp := utils.ProcessForm("/api/auth/login", r.Data, r, model.UsernamePasswordValidator())
		if resp.Bool() {
			gWade.Pager().Redirect("pg-home")
			pnotify(`Login successful`, `Welcome back!`, "success")
		}
	}()
}

type RegisterView struct {
	utils.Validated
	Data model.UsernamePassword
}

func (r *RegisterView) Submit() {
	go func() {
		utils.ProcessForm("/api/auth/register", r.Data, r, model.UsernamePasswordValidator())
	}()
}

type ErrorListModel struct {
	Errors map[string]string
}

type TestStruct struct {
	Message string
}

type Auth struct {
	Authenticated bool
}

func pnotify(title, text, messType string) {
	js.Global.Get("PNotify").New(map[string]interface{}{
		"title": title,
		"text":  text,
		"type":  messType,
	})
}

func main() {
	wade := wd.WadeUp("pg-home", "/web", func(wade *wd.Wade) {
		gWade = wade

		authenticated := false

		wade.Pager().RegisterDisplayScopes(map[string]wd.DisplayScope{
			"pg-home":          wd.MakePage("/home", "Home"),
			"pg-user-login":    wd.MakePage("/user/login", "Login"),
			"pg-user-register": wd.MakePage("/user/register", "Register"),
		})

		wade.RegisterCustomTags("/public/elements.html", map[string]interface{}{
			"errorlist": ErrorListModel{},
		})

		wade.RegisterCustomTags("/public/menu.html", menu.Spec())

		wade.Pager().RegisterController(wd.GlobalDisplayScope, func(p *wd.PageCtrl) interface{} {
			auth := new(Auth)
			ss := pdata.Service(pdata.SessionStorage)
			if authed, ok := ss.GetBool("authed"); ok {
				auth.Authenticated = authed
			} else {
				resp := http.NewRequest("GET", "/api/auth/check").Do()
				resp.DecodeDataTo(&auth.Authenticated)
				ss.Set("authed", true)
			}

			authenticated = auth.Authenticated
			return auth
		})

		wade.Pager().RegisterController("pg-user-login", func(p *wd.PageCtrl) interface{} {
			if authenticated {
				wade.Pager().Redirect("pg-home")
				return nil
			}
			loginView := new(LoginView)
			loginView.Validated.Init(loginView.Data)
			return loginView
		})

		wade.Pager().RegisterController("pg-user-register", func(p *wd.PageCtrl) interface{} {
			regView := new(RegisterView)
			regView.Validated.Init(regView.Data)
			return regView
		})

		pendingRequests := make(chan *http.Request)
		http.Service().AddRequestInterceptor(func(r *http.Request) {
			go func() {
				pendingRequests <- r
			}()
		})

		http.Service().AddResponseInterceptor(func(finishChannel chan bool, r *http.Response) {
			if r.Status() == 401 {
				go func() {
					rt := http.Service().NewRequest(http.MethodGet, "/api/auth/check").Do()

					if !rt.Bool() {
						wade.Pager().Redirect("pg-user-login")
						return
					}

					for {
						select {
						case request := <-pendingRequests:
							request.Do()
						default:
							break
						}
					}

					finishChannel <- true
				}()
			} else {
				go func() {
					finishChannel <- true
				}()
			}
		})
	})

	wade.Start()
}
