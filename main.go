package main

import (
	"fmt"
	"github.com/zmb3/spotify"
	"log"
	"os"
	"os/exec"
	"runtime"
)

const redirectURI = "http://localhost:8080/callback"

var (
	auth  = spotify.NewAuthenticator(redirectURI, spotify.ScopeUserReadPrivate, spotify.ScopeUserLibraryRead)
	ch    = make(chan *spotify.Client)
	state = "abc123"
)

func Open(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
	}

}

func main() {
	functionality := os.Args[1]

	if functionality == "PlaylistExport" {
		client := Spotify()
		PlaylistExport(client)
	} else if functionality == "SongExport" {
		client := Spotify()
		SongExport(client)
	} else {
		fmt.Println("Unknown func: ", functionality)
	}
}
