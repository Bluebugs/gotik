package main

import (
	"crypto/ed25519"
	"log"
	"time"

	"fyne.io/fyne/v2"
	"github.com/fynelabs/fyneselfupdate"
	"github.com/fynelabs/selfupdate"
)

// selfManage turns on automatic updates
func selfManage(a fyne.App, w fyne.Window) {
	publicKey := ed25519.PublicKey{23, 26, 109, 252, 247, 64, 250, 254, 174, 33, 170, 70, 13, 255, 14, 80, 104, 254, 10, 209, 78, 218, 62, 181, 203, 233, 142, 10, 104, 230, 83, 36}

	// The public key above matches the signature of the below file served by our CDN
	httpSource := selfupdate.NewHTTPSource(nil, "https://geoffrey-artefacts.fynelabs.com/self-update/e6/e64b6242-5f87-4476-a95f-44b9dc9f1ed1/{{.OS}}-{{.Arch}}/{{.Executable}}{{.Ext}}")

	config := fyneselfupdate.NewConfigWithTimeout(a, w, time.Minute, httpSource, selfupdate.Schedule{FetchOnStart: true, Interval: time.Hour * 12}, publicKey)

	_, err := selfupdate.Manage(config)
	if err != nil {
		log.Println("Error while setting up update manager: ", err)
		return
	}
}
