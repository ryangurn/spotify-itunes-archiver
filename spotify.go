package main

import (
	"bufio"
	"fmt"
	"github.com/zmb3/spotify"
	"log"
	"net/http"
	"os"
)

func completeAuth(w http.ResponseWriter, r *http.Request) {
	tok, err := auth.Token(state, r)
	if err != nil {
		http.Error(w, "Couldn't get token", http.StatusForbidden)
		log.Fatal(err)
	}
	if st := r.FormValue("state"); st != state {
		http.NotFound(w, r)
		log.Fatalf("State mismatch: %s != %s\n", st, state)
	}
	// use the token to get an authenticated client
	client := auth.NewClient(tok)
	fmt.Fprintf(w, "Login Completed!")
	ch <- &client
}

func PlaylistExport(client *spotify.Client) {
	// create the csv
	file, err := os.OpenFile("spotify-playlists.csv", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}
	defer file.Close()

	datawriter := bufio.NewWriter(file)
	defer datawriter.Flush()

	// write csv header
	datawriter.WriteString("\"Playlist Name\",\"Artists\",\"Album\",\"Track Name\"\n")

	// use the client to make calls that require authorization
	user, err := client.CurrentUser()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("You are logged in as:", user.DisplayName)

	// get playlist for current user
	pages, err := client.GetPlaylistsForUser(user.ID)
	if err != nil {
		log.Fatal(err)
	}

	count := 1
	for playlistPage := 1; ; playlistPage++ {
		// loop through playlists
		for _, playlist := range pages.Playlists {
			fmt.Println("Playlist Counter:", count, "Name:", playlist.Name, "| No of tracks:", playlist.Tracks.Total)
			count++
			playlistData, err := client.GetPlaylistTracks(playlist.ID)
			if err != nil {
				log.Fatal(err)
			}

			for trackPage := 1; ; trackPage++ {
				tracks := playlistData.Tracks

				for i, track := range tracks {
					t := track.Track
					fmt.Print("Track Counter: ",i+1, " ")
					var strArtist string
					for i, artists := range t.Artists {
						strArtist += artists.Name
						fmt.Print(artists.Name, "")
						if i != len(t.Artists)-1 {
							strArtist += " / "
							fmt.Print(",")
						}
					}
					fmt.Println(" >", t.Album.Name ,">", t.Name)
					datawriter.WriteString("\"" + playlist.Name + "\",\"" +  strArtist + "\",\"" + t.Album.Name + "\",\"" + t.Name + "\"\n")
				}

				err = client.NextPage(playlistData)
				if err == spotify.ErrNoMorePages {
					break
				}
				if err != nil {
					log.Fatal(err)
				}
			}
			fmt.Println("----")

		}

		err = client.NextPage(pages)
		if err == spotify.ErrNoMorePages {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
	}
}

func SongExport(client *spotify.Client) {
	// create the csv
	file, err := os.OpenFile("spotify-songs.csv", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}
	defer file.Close()

	datawriter := bufio.NewWriter(file)
	defer datawriter.Flush()

	// write csv header
	datawriter.WriteString("\"Artists\",\"Album\",\"Track Name\"\n")

	trackPage, err := client.CurrentUsersTracks()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Total Tracks: ", trackPage.Total)
	TrackCount := 1
	for page := 1; ; page++ {
		log.Printf("  Page %d has %d tracks", page, len(trackPage.Tracks))

		for _, track := range trackPage.Tracks {
			fmt.Print("Track Count: ", TrackCount, " | ")
			TrackCount++

			var strArtist string
			for i, artists := range track.Artists {
				strArtist += artists.Name
				fmt.Print(artists.Name, "")
				if i != len(track.Artists)-1 {
					strArtist += " / "
					fmt.Print(",")
				}
			}
			datawriter.WriteString("\"" +  strArtist + "\",\"" + track.Album.Name + "\",\"" + track.Name + "\"\n")
			fmt.Println(" >", track.Album.Name ,">", track.Name)
		}


		err = client.NextPage(trackPage)
		if err == spotify.ErrNoMorePages {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
	}
}

func Spotify() *spotify.Client {
	// first start an HTTP server
	http.HandleFunc("/callback", completeAuth)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Got request for:", r.URL.String())
	})
	go http.ListenAndServe(":8080", nil)

	url := auth.AuthURL(state)
	Open(url)

	// wait for auth to complete
	client := <-ch
	return client
}
