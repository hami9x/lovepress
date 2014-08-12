package main

import (
	"testing"
	jwt "github.com/dgrijalva/jwt-go"

	"github.com/phaikawl/lovepress/model"
)

func TestToken(t *testing.T) {
	key := "n0t9r34t6cz9r34tn0t1na9r34tw4y"
	api := AuthApi{key}
	tok, err := api.makeAccessToken("test", model.RoleUser)
	if err != nil {
		t.Fatal(err.Error())
	}

	token, er := jwt.Parse(tok, func(t *jwt.Token) ([]byte, error) {
		return []byte(key), nil
	})

	if er != nil {
		t.Fatal(er.Error())
	}

	if token.Claims["username"].(string) != "test" {
		t.Fatalf("Expected %v, got %v.", "test", token.Claims["username"].(string))
	}

	if !api.checkToken(token) {
		t.Fatalf("Check token failed.")
	}
}
