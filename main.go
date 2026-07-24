package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := NewApp()
	if err := wails.Run(&options.App{
		Title:  "LeonAgent",
		Width:  1440,
		Height: 900,
		MinWidth: 1100,
		MinHeight: 700,
		AssetServer: &assetserver.Options{Assets: assets},
		BackgroundColour: &options.RGBA{R: 15, G: 23, B: 42, A: 1},
		OnStartup: app.startup,
		Bind: []interface{}{app},
	}); err != nil {
		println("Error:", err.Error())
	}
}
