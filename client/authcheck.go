package client

import (
	"github.com/phaikawl/wade"
	"github.com/phaikawl/wade/libs/http"
)

type AuthModule struct {
	Authenticated bool
	loginPage     string
	checkUrl      string
}

func (s *AuthModule) Init(services wade.GlobalServices) {
	s.interceptors(services)
}

func NewAuthModule(loginPage, checkurl string) *AuthModule {
	return &AuthModule{
		Authenticated: false,
		loginPage:     loginPage,
		checkUrl:      checkurl,
	}
}

func (s *AuthModule) interceptors(services wade.GlobalServices) {
	httpService := services.Http

	checkResult := -1
	processFn := func() int {
		//would retry 3 times if network fails
		var rt *http.Response
		var err error
		for i := 0; i < 4; i++ {
			rt, err = httpService.DoPure(httpService.NewRequest("GET", s.checkUrl))
			if err == nil {
				break
			}
		}

		if err != nil || !rt.Bool() {
			return 0
		} else {
			return 1
		}
	}

	httpService.AddResponseInterceptor(func(finishChannel chan bool, r *http.Request) {
		if r.Response.StatusCode == 401 && checkResult != 0 {
			var fn func()
			fn = func() {
				if checkResult == -1 {
					checkResult = processFn()
				}

				if checkResult == 0 {
					services.PageManager.RedirectToPage(s.loginPage)
					return
				}

				resp, _ := httpService.DoPure(r)
				if resp.StatusCode == 401 {
					checkResult = -1
					fn()
				}
			}
		}

		finishChannel <- true
	})
}

func (s *AuthModule) CheckAuthPage(pc wade.ThisPage) interface{} {
	session := pc.Services().SessionStorage
	if authed, ok := session.GetBool("authed"); ok {
		s.Authenticated = authed
	} else {
		resp, err := pc.Services().Http.GET("/api/auth/check")
		if err != nil || resp.Failed() {
			s.Authenticated = false
			return s
		}

		resp.DecodeTo(&s.Authenticated)
		session.Set("authed", s.Authenticated)
	}

	return s
}
