package main

import (
	"github.com/phaikawl/lovepress/model"
	"github.com/phaikawl/wade"
	"github.com/phaikawl/wade/utils"
)

type UserModule struct {
	auth         *AuthModule
	redirectPage string
}

func NewUserModule(authModule *AuthModule, redirectPage string) *UserModule {
	return &UserModule{
		auth:         authModule,
		redirectPage: redirectPage,
	}
}

type LoginModel struct {
	redirectPage string
	p            wade.ThisPage
	utils.Validated
	Data model.UsernamePassword
}

func (r *LoginModel) Submit() {
	go func() {
		resp, err := utils.ProcessForm(r.p.Services().Http, "/api/auth/login", r.Data, r, model.UsernamePasswordValidator())
		if err == nil && resp.Bool() {
			r.p.RedirectToPage(r.redirectPage)
			Notify(`Login successful`, `Welcome back!`, "success")
		}
	}()
}

func (um *UserModule) LoginPage(p wade.ThisPage) interface{} {
	if um.auth.Authenticated {
		p.RedirectToPage(um.redirectPage)
		return nil
	}

	loginModel := &LoginModel{
		redirectPage: um.redirectPage,
		p:            p,
	}

	loginModel.Validated.Init(loginModel.Data)
	return loginModel
}
