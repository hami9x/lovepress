package main

import (
	"github.com/gopherjs/gopherjs/js"

	"github.com/phaikawl/wade"
	"github.com/phaikawl/wade/elements/menu"
)

type ErrorListModel struct {
	Errors map[string]string
}

func Notify(title, text, messType string) {
	js.Global.Get("PNotify").New(map[string]interface{}{
		"title": title,
		"text":  text,
		"type":  messType,
	})
}

func initFunc(r wade.Registration, app wade.AppEnv) {
	r.RegisterDisplayScopes(map[string]wade.DisplayScope{
		"pg-home":          wade.MakePage("/home", "Home"),
		"pg-user-login":    wade.MakePage("/user/login", "Login"),
		"pg-user-register": wade.MakePage("/user/register", "Register"),
	})

	authModule := NewAuthModule(
		"pg-user-login",
		"/api/auth/check")

	userModule := NewUserModule(authModule, "pg-home")

	r.ModuleInit(authModule)

	//Controllers

	r.RegisterController(wade.GlobalDisplayScope, authModule.CheckAuthPage)
	r.RegisterController("pg-user-login", userModule.LoginPage)

	//

	r.RegisterCustomTags("/public/elements.html", map[string]interface{}{
		"errorlist": ErrorListModel{},
	})

	r.RegisterCustomTags("/public/menu.html", menu.Spec())
}

func main() {
	err := wade.StartApp("pg-home", "/web", initFunc)
	if err != nil {
		panic("Failed to load.")
	}
}
