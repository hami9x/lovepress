package client

import (
	"github.com/gopherjs/gopherjs/js"

	"github.com/phaikawl/wade"
	"github.com/phaikawl/wade/elements/menu"
)

type ErrorListModel struct {
	wade.BaseProto
	Errors map[string]string
}

func Notify(title, text, messType string) {
	js.Global.Get("PNotify").New(map[string]interface{}{
		"title": title,
		"text":  text,
		"type":  messType,
	})
}

func InitFunc(r wade.Registration) {
	r.RegisterDisplayScopes([]wade.PageDesc{
		wade.MakePage("pg-home", "/home", "Home"),
		wade.MakePage("pg-user-login", "/user/login", "Login"),
		wade.MakePage("pg-user-register", "/user/register", "Register"),
	}, []wade.PageGroupDesc{})

	authModule := NewAuthModule(
		"pg-user-login",
		"/api/auth/check")

	userModule := NewUserModule(authModule, "pg-home")

	r.ModuleInit(authModule)

	//Controllers

	r.RegisterController(wade.GlobalDisplayScope, authModule.CheckAuthPage)
	r.RegisterController("pg-user-login", userModule.LoginPage)

	//

	r.RegisterCustomTags(wade.CustomTag{
		Name:       "errorlist",
		Attributes: []string{"Errors"},
		Prototype:  &ErrorListModel{},
		Html: `
			<ul class="error" bind-ifn="isEmpty(Errors)">
				<li bind-each="Errors -> _, error">
					<% error %>
				</li>
			</ul>
		`,
	})

	r.RegisterCustomTags(menu.CustomTags()...)
}
