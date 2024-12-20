package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gdamore/tcell"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/exp/rand"
	"golang.org/x/oauth2"
)

type Note struct {
	Time  int
	Key   string
	X     int
	Y     int
	Style tcell.Style
}

var (
	redirectURI   = "http://localhost:8080/callback"
	authenticator = spotifyauth.New(
		spotifyauth.WithRedirectURL(redirectURI), spotifyauth.WithScopes(
			spotifyauth.ScopeUserReadPrivate,
			spotifyauth.ScopeUserModifyPlaybackState,
		))
	state     = "spotify_auth_state"
	tokenFile = "spotify_token.json"
	ch        = make(chan *spotify.Client)

	notes = []*Note{
		{Time: 0, Key: "A", X: 0, Y: 0, Style: tcell.StyleDefault.Foreground(tcell.ColorGreen)},
		{Time: 0, Key: "S", X: 4, Y: 0, Style: tcell.StyleDefault.Foreground(tcell.ColorRed)},
		{Time: 0, Key: "J", X: 8, Y: 0, Style: tcell.StyleDefault.Foreground(tcell.ColorYellow)},
		{Time: 0, Key: "K", X: 12, Y: 0, Style: tcell.StyleDefault.Foreground(tcell.ColorBlue)},
		{Time: 0, Key: "L", X: 16, Y: 0, Style: tcell.StyleDefault.Foreground(tcell.ColorOrange)},
	}
)

func checkNotePress(notes []*Note, gameTime int, key string, score *int) {
	for _, n := range notes {
		if n.Key == strings.ToUpper(key) && n.Y == 9 && gameTime == n.Time+8 {
			n.Y++
			*score = *score + 1
		}
	}
}

func getSpotifyClient() (*spotify.Client, error) {
	// Check if a saved token exists
	if _, err := os.Stat(tokenFile); err == nil {
		// Load token from file
		token, err := loadToken()
		if err != nil {
			return nil, fmt.Errorf("failed to load token: %w", err)
		}

		// Check if the token is expired
		if token.Expiry.Before(time.Now()) {
			fmt.Println("Token expired. Please re-authenticate.")
			return authenticate()
		}

		// Create a Spotify client with the valid token
		return spotify.New(authenticator.Client(context.Background(), token)), nil
	}

	// Authenticate if no valid token exists
	return authenticate()
}

func authenticate() (*spotify.Client, error) {
	// Start HTTP server for authentication callback
	http.HandleFunc("/callback", completeAuth)
	go func() {
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Generate Spotify auth URL
	url := authenticator.AuthURL(state)
	fmt.Println("Please log in to Spotify by visiting the following page in your browser:", url)

	// Wait for authentication to complete (use the global `ch` channel)
	client := <-ch
	return client, nil
}

func completeAuth(w http.ResponseWriter, r *http.Request) {
	token, err := authenticator.Token(r.Context(), state, r)
	if err != nil {
		http.Error(w, "Couldn't get token", http.StatusForbidden)
		log.Fatalf("Couldn't get token: %v", err)
	}

	// Save the token for later use
	saveToken(token)

	// Create a Spotify client
	client := spotify.New(authenticator.Client(r.Context(), token))
	fmt.Fprintln(w, "Login Completed! You can now close this tab.")
	// Signal completion
	ch <- client
}

func saveToken(token *oauth2.Token) {
	file, err := os.Create(tokenFile)
	if err != nil {
		log.Fatalf("Failed to save token: %v", err)
	}
	defer file.Close()

	if err := json.NewEncoder(file).Encode(token); err != nil {
		log.Fatalf("Failed to encode token: %v", err)
	}
}

func loadToken() (*oauth2.Token, error) {
	file, err := os.Open(tokenFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var token oauth2.Token
	if err := json.NewDecoder(file).Decode(&token); err != nil {
		return nil, err
	}
	return &token, nil
}

func drawText(s tcell.Screen, x1, y1, x2, y2 int, style tcell.Style, text string) {
	row := y1
	col := x1
	for _, r := range text {
		s.SetContent(col, row, r, nil, style)
		col++
		if col >= x2 {
			row++
			col = x1
		}
		if row > y2 {
			break
		}
	}
}

func main() {
	song := flag.String("query", "Metallica Enter Sandman", "Spotify query for song")
	flag.Parse()

	client, err := getSpotifyClient()
	if err != nil {
		log.Fatalf("Failed to create Spotify client: %v", err)
	}

	results, err := client.Search(context.Background(), *song, spotify.SearchTypeTrack)
	if err != nil {
		log.Fatalf("Error searching for track: %v", err)
	}

	if len(results.Tracks.Tracks) > 0 {
		track := results.Tracks.Tracks[0]
		fmt.Printf("Playing: %s by %s\n ", track.Name, track.Artists[0].Name)

		// Play the track
		err := client.PlayOpt(context.Background(), &spotify.PlayOptions{
			URIs: []spotify.URI{track.URI},
		})
		if err != nil {
			log.Fatalf("Error playing track: %v", err)
		}
	} else {
		fmt.Println("No results found!")
	}

	s, _ := tcell.NewScreen()
	s.Init()
	s.SetStyle(tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorBlack))
	s.Clear()

	quit := func() {
		maybePanic := recover()
		s.Fini()
		if maybePanic != nil {
			panic(maybePanic)
		}
	}
	defer quit()

	gameTime := 0
	score := 0

	done := make(chan bool)
	go func() {
		<-done
		s.Fini()
		fmt.Println("Score: ", score)
		os.Exit(0)
	}()

	// Generate notes for Track
	var songNotes []*Note
	songTime := 0
	for i := 0; i < 1000; i++ {
		random := rand.Intn(len(notes))
		newNote := *notes[random]                  // Dereference to avoid modifying the original
		newNote.Time = songTime + rand.Intn(5) + 1 // Generate random space between notes
		songTime = newNote.Time                    // Update the time for the next note
		songNotes = append(songNotes, &newNote)    // Append the new note to songNotes
	}

	// Handle events
	go func() {
		for {
			ev := s.PollEvent()
			switch ev := ev.(type) {
			case *tcell.EventResize:
				s.Sync()
			case *tcell.EventKey:
				// Escape Screen
				if ev.Key() == tcell.KeyEscape || ev.Key() == tcell.KeyCtrlC {
					s.Fini()
					done <- true
				} else { // Match all other key presses
					notePress := strings.ToUpper(string(ev.Rune()))
					checkNotePress(songNotes, gameTime, notePress, &score)

					var noteInfo Note
					for _, note := range notes {
						if note.Key == notePress {
							noteInfo = *note
						}
					}

					drawText(s, noteInfo.X, 8, 0, 0, noteInfo.Style, "*")
					s.Show()
				}
			}
		}
	}()

	t := time.NewTicker(time.Second / 8)
	for {
		for _, note := range songNotes {
			if note.Time <= gameTime && note.Y < 9 {
				drawText(s, note.X, note.Y, 0, 0, note.Style, note.Key)
				note.Y += 1
			}
		}

		drawText(s, 20, 0, 30, 0, tcell.StyleDefault.Foreground(tcell.ColorGreen), fmt.Sprintf("Score: %d", score))
		s.Show()

		<-t.C // Directly receive from the channel
		gameTime++
		s.Clear()
		for _, note := range notes {
			drawText(s, note.X, 8, 0, 0, note.Style, note.Key)
		}
		s.Sync()
		s.Show()
	}
}
