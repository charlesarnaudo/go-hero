package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gdamore/tcell"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/exp/rand"
	"golang.org/x/oauth2"
)

const (
	ALocation = 0
	SLocation = 4
	JLocation = 8
	KLocation = 12
	LLocation = 16
)

type Note struct {
	Time  int
	Key   string
	X     int         // Replace with appropriate type if not int
	Y     int         // Replace with appropriate type if not int
	Style tcell.Style // Replace with appropriate type if not string
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
)

func main() {
	client, err := getSpotifyClient()

	if err != nil {
		log.Fatalf("Failed to create Spotify client: %v", err)
	}

	song := flag.String("query", "Metallica Enter Sandman", "Spotify query for song")
	flag.Parse()
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

	defStyle := tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorBlack)
	aNoteStyle := tcell.StyleDefault.Foreground(tcell.ColorGreen)
	sNoteStyle := tcell.StyleDefault.Foreground(tcell.ColorRed)
	jNoteStyle := tcell.StyleDefault.Foreground(tcell.ColorYellow)
	kNoteStyle := tcell.StyleDefault.Foreground(tcell.ColorBlue)
	lNoteStyle := tcell.StyleDefault.Foreground(tcell.ColorOrange)

	s, _ := tcell.NewScreen()
	if err := s.Init(); err != nil {
		log.Fatalf("%+v", err)
	}
	s.SetStyle(defStyle)
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

	notes := []*Note{
		{Time: 0, Key: "A", X: ALocation, Y: 0, Style: aNoteStyle},
		{Time: 0, Key: "S", X: SLocation, Y: 0, Style: sNoteStyle},
		{Time: 0, Key: "J", X: JLocation, Y: 0, Style: jNoteStyle},
		{Time: 0, Key: "K", X: KLocation, Y: 0, Style: kNoteStyle},
		{Time: 0, Key: "L", X: LLocation, Y: 0, Style: lNoteStyle},
	}

	var songNotes []*Note
	songTime := 0

	for i := 0; i < 1000; i++ {
		random := rand.Intn(len(notes))
		newNote := *notes[random]                  // Dereference to avoid modifying the original
		newNote.Time = songTime + rand.Intn(5) + 1 // Increment time by 6
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
				if ev.Key() == tcell.KeyEscape || ev.Key() == tcell.KeyCtrlC {
					s.Fini()
					done <- true
				} else if ev.Rune() == 'a' || ev.Rune() == 'A' {
					checkNotePress(songNotes, gameTime, "A", &score)
					//Simulate note press with * character
					drawText(s, ALocation, 8, 0, 0, aNoteStyle, "*")
					s.Show()
				} else if ev.Rune() == 's' || ev.Rune() == 'S' {
					checkNotePress(songNotes, gameTime, "S", &score)
					//Simulate note press with * character
					drawText(s, SLocation, 8, 0, 0, sNoteStyle, "*")
					s.Show()
				} else if ev.Rune() == 'j' || ev.Rune() == 'J' {
					checkNotePress(songNotes, gameTime, "J", &score)
					//Simulate note press with * character
					drawText(s, JLocation, 8, 0, 0, jNoteStyle, "*")
					s.Show()
				} else if ev.Rune() == 'k' || ev.Rune() == 'K' {
					checkNotePress(songNotes, gameTime, "K", &score)
					//Simulate note press with * character
					drawText(s, KLocation, 8, 0, 0, kNoteStyle, "*")
					s.Show()
				} else if ev.Rune() == 'l' || ev.Rune() == 'L' {
					checkNotePress(songNotes, gameTime, "L", &score)
					//Simulate note press with * character
					drawText(s, LLocation, 8, 0, 0, lNoteStyle, "*")
					s.Show()
				}
			}
		}
	}()

	t := time.NewTicker(time.Second / 8)
	// songStarted := false
	// Main game loop

	for {
		for _, note := range songNotes {
			if note.Time <= gameTime && note.Y < 9 {
				drawText(s, note.X, note.Y, 0, 0, note.Style, note.Key)
				note.Y += 1
			}
		}

		drawText(s, 20, 0, 30, 0, aNoteStyle, fmt.Sprintf("Score: %d", score))
		s.Show()

		select {
		case <-t.C:
			gameTime++
			s.Clear()
			drawNoteLocations(s, []tcell.Style{aNoteStyle, sNoteStyle, jNoteStyle, kNoteStyle, lNoteStyle})
			s.Sync()
			s.Show()
		}
	}
}

func checkNotePress(notes []*Note, gameTime int, key string, score *int) {
	for _, n := range notes {
		if n.Key == key && n.Y == 9 && gameTime == n.Time+8 {
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

func drawNoteLocations(s tcell.Screen, styles []tcell.Style) {
	drawText(s, ALocation, 8, 0, 0, styles[0], "A")
	drawText(s, SLocation, 8, 0, 0, styles[1], "S")
	drawText(s, JLocation, 8, 0, 0, styles[2], "J")
	drawText(s, KLocation, 8, 0, 0, styles[3], "K")
	drawText(s, LLocation, 8, 0, 0, styles[4], "L")
}

func drawText(s tcell.Screen, x1, y1, x2, y2 int, style tcell.Style, text string) {
	row := y1
	col := x1
	for _, r := range []rune(text) {
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
