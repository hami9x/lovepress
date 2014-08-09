package main

import (
	wd "github.com/phaikawl/wade"
	"github.com/phaikawl/wade/elements/menu"
	"github.com/phaikawl/wade/utils"

	"github.com/phaikawl/lovepress/model"
)

type LoginView struct {
	utils.Validated
	Data model.UsernamePassword
}

func (r *LoginView) Submit() {
	go func() {
		utils.ProcessForm("/api/user/login", r.Data, r, model.UsernamePasswordValidator())
	}()
}

type RegisterView struct {
	utils.Validated
	Data model.UsernamePassword
}

func (r *RegisterView) Submit() {
	go func() {
		utils.ProcessForm("/api/user/register", r.Data, r, model.UsernamePasswordValidator())
	}()
}

type ErrorListModel struct {
	Errors map[string]string
}

func main() {
	wade := wd.WadeUp("pg-home", "/web", func(wade *wd.Wade) {
		wade.Pager().RegisterDisplayScopes(map[string]wd.DisplayScope{
			"pg-home":          wd.MakePage("/home", "Home"),
			"pg-user-login":    wd.MakePage("/user/login", "Login"),
			"pg-user-register": wd.MakePage("/user/register", "Register"),
		})

		wade.RegisterCustomTags("/public/elements.html", map[string]interface{}{
			"errorlist": ErrorListModel{},
		})

		wade.RegisterCustomTags("/public/menu.html", menu.Spec())

		wade.Pager().RegisterController("pg-user-login", func(p *wd.PageCtrl) interface{} {
			loginView := new(LoginView)
			loginView.Validated.Init(loginView.Data)
			return loginView
		})

		wade.Pager().RegisterController("pg-user-register", func(p *wd.PageCtrl) interface{} {
			regView := new(RegisterView)
			regView.Validated.Init(regView.Data)
			return regView
		})
	})

	wade.Start()
}
