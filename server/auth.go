package main

import (
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"

	"github.com/phaikawl/lovepress/model"
	"github.com/phaikawl/lovepress/server/dblayer"
)

const (
	AccessTokenCookie   = "accessToken"
	AccessTokenLifetime = time.Second * 60
)

type AuthApi struct {
	jwtSigningKey string
}

func (api AuthApi) ProvideApi(apis ApiSystem) {
	apis.Protected.Use(func(c *Context) {
		var ok bool
		c.user, ok = api.checkAuth(c.Request)
		if !ok {
			c.Abort(http.StatusUnauthorized)
			return
		}
	})

	apis.Dev.GET("/auth/resettoken", api.DevResetToken)

	authApi := apis.Public.Group("/auth", nil)
	authApi.POST("/login", api.Login)
	authApi.POST("/register", api.Register)
	authApi.GET("/check", api.ResetToken)
}

func (api AuthApi) checkToken(token *jwt.Token, checkExpire bool) bool {
	if !token.Valid {
		return false
	}

	if t, ok := token.Claims["time"].(float64); ok {
		if !checkExpire {
			return true
		}

		ttime := time.Unix(int64(t), -1)
		if time.Now().Sub(ttime) <= AccessTokenLifetime {
			return true
		}
	}

	return false
}

func (api AuthApi) checkAuth(request *http.Request) (*model.User, bool) {
	token, ok := api.getAccessToken(request)
	if ok && api.checkToken(token, true) {
		return &model.User{
			Username: token.Claims["username"].(string),
			Role:     token.Claims["role"].(model.UserRole),
		}, true
	}

	return nil, false
}

func (api AuthApi) getAccessToken(request *http.Request) (token *jwt.Token, ok bool) {
	if cookie, err := request.Cookie(AccessTokenCookie); err == nil {
		token, err := jwt.Parse(cookie.Value, func(t *jwt.Token) ([]byte, error) {
			return []byte(api.jwtSigningKey), nil
		})

		if err == nil {
			return token, true
		}
	}

	return nil, false
}

func (api AuthApi) setTokenCookie(writer http.ResponseWriter, token string) {
	http.SetCookie(writer, &http.Cookie{
		Name:   AccessTokenCookie,
		Value:  token,
		Path:   g.BasePath(),
		MaxAge: 3600 * 24 * 60,
	})
}

func (api AuthApi) makeAccessToken(username string, role model.UserRole) (tokenString string, err error) {
	token := jwt.New(jwt.GetSigningMethod("HS256"))
	token.Claims["username"] = username
	token.Claims["role"] = role
	token.Claims["time"] = time.Now().Unix()
	tokenString, err = token.SignedString([]byte(api.jwtSigningKey))
	return
}

func (api AuthApi) tokenMatching(token *jwt.Token, utoken string) bool {
	savedToken, err := jwt.Parse(utoken, func(t *jwt.Token) ([]byte, error) {
		return []byte(api.jwtSigningKey), nil
	})

	if err != nil || !api.checkToken(savedToken, false) {
		return false
	}

	if ttime, ok := token.Claims["time"].(float64); ok {
		if int64(ttime) == int64(savedToken.Claims["time"].(float64)) {
			return true
		}
	}

	return false
}

func (api AuthApi) resetToken(c *Context, validate bool) bool {
	var err error
	token, hasToken := api.getAccessToken(c.Request)
	if hasToken && (!validate || api.checkToken(token, false)) {
		user, ok := dblayer.User.ByUsername(token.Claims["username"].(string))
		if ok {
			if !validate || api.tokenMatching(token, user.Token) {
				user.Token, err = api.makeAccessToken(user.Username, user.Role)
				if err != nil {
					return false
				}

				dblayer.User.Update(user)
				api.setTokenCookie(c.Writer, user.Token)
				return true
			}
		}
	}

	return false
}

func (api AuthApi) ResetToken(c *Context) {
	c.Response(api.resetToken(c, true))
}

func (api AuthApi) DevResetToken(c *Context) {
	c.Response(api.resetToken(c, false))
}

func (api AuthApi) Login(c *Context) {
	logindata := model.UsernamePassword{}
	if err := c.ParseBody(&logindata); err != nil {
		c.ReportError(err, http.StatusBadRequest)
		return
	}

	user, ok := dblayer.User.ByUsername(logindata.Username)
	if !ok || user.Password != logindata.Password {
		c.Response(false)
		return
	}

	api.setTokenCookie(c.Writer, user.Token)

	c.Response(true)
}

func (api AuthApi) Register(c *Context) {
	udata := model.UsernamePassword{}
	if err := c.ParseBody(&udata); err != nil {
		c.ReportError(err, http.StatusBadRequest)
		return
	}

	failed := model.UsernamePasswordValidator().Validate(udata).HasErrors()
	if !failed {
		token, err := api.makeAccessToken(udata.Username, model.RoleUser)
		c.CheckError(err, ServerError)
		c.CheckError(dblayer.User.Create(&model.User{
			Username: udata.Username,
			Password: udata.Password,
			Role:     model.RoleUser,
			Token:    token,
		}), ServerError)
	}

	c.Response(!failed)
}
