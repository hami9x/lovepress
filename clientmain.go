package main

import (
	"github.com/phaikawl/lovepress/client"
	"github.com/phaikawl/wade"
	"github.com/phaikawl/wade/rbackend/clientside"
)

func main() {
	err := wade.StartApp(wade.AppConfig{
		StartPage: "pg-home",
		BasePath:  "/web",
	}, client.InitFunc, clientside.RenderBackend())
	if err != nil {
		panic("Failed to load.")
	}
}
