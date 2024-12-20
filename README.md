# Go Hero 🎸

**Go Hero** is a terminal-based Guitar Hero-like game written in Go. Sync your gameplay with live Spotify tracks and hit falling notes to the rhythm of your favorite songs—all from the comfort of your terminal!

---

## Features

- **Spotify Integration**: Play along with live Spotify tracks.
- **Terminal-Based Experience**: Simple and clean interface powered by Tcell.
- **Oauth Token Reuse**: Dynamically issue Oauth tokens based on client status.

---

## Getting Started

### Prerequisites

- **Go**: Install [Go](https://golang.org/doc/install).
- **Spotify Developer Account**: Create an app in the [Spotify Developer Dashboard](https://developer.spotify.com/dashboard/) to get your `client_id` and `client_secret`.
- **Spotify Premium Account**: Required for playback control via Spotify's API.

### Installation

1. Clone the repository:
   ```bash
   git clone <repository_url>
   cd go-hero
    ```
2. Install go deps
   ```bash
   go mod tidy
   ```

3. Set Spotify secrets
   ```bash
   export SPOTIFY_ID="your_client_id"
   export SPOTIFY_SECRET="your_client_secret"
   ```

4. Run the application
   ```bash
   go run main.go --query "Title Fight Shed"
   ```
