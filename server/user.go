package main

import (
	"net/http"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"

	"github.com/phaikawl/lovepress/model"
	"github.com/phaikawl/lovepress/server/dblayer"
)

func makeUserAccessToken(username string, role model.UserRole) (tokenString string, err error) {
	token := jwt.New(jwt.GetSigningMethod("HS256"))
	token.Claims["username"] = username
	token.Claims["role"] = role
	tokenString, err = token.SignedString([]byte(MySigningKey))
	return
}

func userApi(api *gin.RouterGroup) {
	// The api to register a user
	// Used in the pg-user-register page
	uapi := api.Group("/user", func(c *gin.Context) {})

	uapi.POST("/login", func(c *gin.Context) {
		logindata := model.UsernamePassword{}
		if err := c.ParseBody(&logindata); err != nil {
			reportError(err, c, http.StatusBadRequest)
			return
		}

		user := dblayer.User.ByUsername(logindata.Username)
		if user == nil || user.Password != logindata.Password {
			c.JSON(http.StatusOK, false)
			return
		}

		http.SetCookie(c.Writer, &http.Cookie{
			Name:  "accessToken",
			Value: user.Token,
		})

		c.JSON(http.StatusOK, true)
	})

	uapi.POST("/register", func(c *gin.Context) {
		udata := model.UsernamePassword{}
		if err := c.ParseBody(&udata); err != nil {
			reportError(err, c, http.StatusBadRequest)
			return
		}

		failed := model.UsernamePasswordValidator().Validate(udata).HasErrors()
		if !failed {
			token, err := makeUserAccessToken(udata.Username, model.RoleUser)
			checkError(err, c, http.StatusInternalServerError)
			checkError(dblayer.User.Create(&model.User{
				UsernamePassword: udata,
				Role:             model.RoleUser,
				Token:            token,
			}), c, http.StatusInternalServerError)
		}
		c.JSON(http.StatusOK, !failed)
	})
}
