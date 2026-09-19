package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
)

// Tutto il frontend finisce dentro l'eseguibile: alla fine si distribuisce
// un file solo, senza cartelle di appoggio accanto.
//
//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := NewApp()

	err := wails.Run(&options.App{
		Title:     "Macchina di Turing",
		Width:     1100,
		Height:    800,
		MinWidth:  820,
		MinHeight: 600,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		// Colore dipinto prima che la pagina sia pronta. È un valore fisso
		// e non può seguire il tema, quindi resta un grigio neutro; su macOS
		// ci pensa Appearance qui sotto a seguire chiaro e scuro.
		BackgroundColour: &options.RGBA{R: 128, G: 128, B: 128, A: 1},
		Mac: &mac.Options{
			Appearance: mac.DefaultAppearance,
			About: &mac.AboutInfo{
				Title:   "Macchina di Turing",
				Message: "Simulatore didattico di macchina di Turing\nnastro infinito, quintuple, esecuzione passo passo",
			},
		},
		OnStartup: app.startup,
		Bind: []interface{}{
			app,
		},
	})
	if err != nil {
		println("Errore:", err.Error())
	}
}
