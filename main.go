package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
)

type Token struct {
	UUID    uuid.UUID
	Token   string
	Expiry  time.Time
	Refresh string
}

type SpotifyTokenRefresh struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
	Scope       string `json:"scope"`
}

type TokenJson struct {
	Token string `json:"token"`
}

type TokenJsonSpotify struct {
	Token string `json:"SpotifyToken"`
}

var (
	auth      spotify.Authenticator
	ch        = make(chan *spotify.Client)
	Tokens    []*Token
	url_login string
)

func main() {

	ConfigInit()
	readTokensFromFile()
	r := mux.NewRouter()

	auth = spotify.NewAuthenticator(config.Callbackuri, spotify.ScopeUserReadCurrentlyPlaying, spotify.ScopeUserReadPlaybackState)
	auth.SetAuthInfo(config.Clientid, config.Clientsecret)

	// first start an HTTP server
	r.HandleFunc("/callback", completeAuth)
	r.HandleFunc("/login", login)
	r.HandleFunc("/session/{token}", tokenToSpotify)

	go checkForUpdates()

	/*addr, err := determineListenAddress() //Get listening address
	if err != nil {
		log.Fatal(err)
	}*/

	//log.Fatal(http.ListenAndServe(addr, r))
	log.Fatal(http.ListenAndServe(":80", r))
}

func login(w http.ResponseWriter, r *http.Request) {
	keys, ok := r.URL.Query()["returnurl"]
	if !ok || len(keys[0]) < 1 {
		url_login = auth.AuthURL("abc123")
	} else {
		url_login = auth.AuthURL(url.QueryEscape(keys[0]))
	}
	http.Redirect(w, r, url_login, http.StatusSeeOther)
}

func completeAuth(w http.ResponseWriter, r *http.Request) {

	var tok *oauth2.Token
	var err error

	keys, ok := r.URL.Query()["state"]
	if (!ok || len(keys[0]) < 1) && keys[0] == "abc123" {
		tok, err = auth.Token("abc123", r)

		if err != nil {
			http.Error(w, "Couldn't get token", http.StatusForbidden)
			log.Fatal(err)
		}
		if st := r.FormValue("state"); st != "abc123" {
			http.NotFound(w, r)
			log.Fatalf("State mismatch: %s != %s\n", st, "abc123")
		}
	} else {
		tok, err = auth.Token(keys[0], r)

		if err != nil {
			http.Error(w, "Couldn't get token", http.StatusForbidden)
			log.Fatal(err)
		}
		if st := r.FormValue("state"); st != keys[0] {
			http.NotFound(w, r)
			log.Fatalf("State mismatch: %s != %s\n", st, keys[0])
		}
	}

	var newToken Token
	newToken.UUID = uuid.New()
	newToken.Token = tok.AccessToken
	newToken.Expiry = tok.Expiry
	newToken.Refresh = tok.RefreshToken

	Tokens = append(Tokens, &newToken)

	w.Header().Set("Content-Type", "application/json")
	var tokenjson TokenJson
	tokenjson.Token = newToken.UUID.String()

	tokenjsonMarshal, err := json.Marshal(tokenjson)
	if err != nil {
		log.Fatal(err)
	}

	writeTokensToFile()

	if len(keys) > 0 && keys[0] != "abc123" {
		decodedurl, err := url.QueryUnescape(keys[0])
		if err != nil {
			log.Fatal(err)
		}
		http.Redirect(w, r, decodedurl+"/"+tokenjson.Token, http.StatusSeeOther)
	} else {
		w.Write(tokenjsonMarshal)
	}

}

func tokenToSpotify(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var found bool
	w.Header().Set("Content-Type", "application/json")

	for _, token := range Tokens {
		if token.UUID.String() == vars["token"] {
			found = true
			var tokenJson TokenJsonSpotify
			tokenJson.Token = token.Token

			tokenjsonMarshal, err := json.Marshal(tokenJson)
			if err != nil {
				log.Fatal(err)
			}

			w.Write(tokenjsonMarshal)
		}
	}

	if !found {
		io.WriteString(w, "{\"error\":\"no token found\"}")
	}

}

func checkForUpdates() {
	for range time.Tick(time.Second * 10) {
		for _, token := range Tokens {
			time := strings.Split(time.Until(token.Expiry).String(), "m")
			timeInt, err := strconv.Atoi(time[0])
			if err == nil {
				if timeInt < 10 {
					updateSpotifyToken(token)
					writeTokensToFile()
				}
			} else {
				fmt.Println(err)
			}
		}
	}
}

func updateSpotifyToken(token *Token) {
	apiUrl := "https://accounts.spotify.com"
	resource := "/api/token"
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", token.Refresh)

	u, _ := url.ParseRequestURI(apiUrl)
	u.Path = resource
	urlStr := u.String()

	client := &http.Client{}
	r, _ := http.NewRequest("POST", urlStr, strings.NewReader(data.Encode())) // URL-encoded payload
	r.SetBasicAuth(config.Clientid, config.Clientsecret)
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	resp, _ := client.Do(r)

	var spot SpotifyTokenRefresh
	decoder := json.NewDecoder(resp.Body)
	err := decoder.Decode(&spot)
	if err != nil {
		panic(err)
	}

	token.Token = spot.AccessToken
	token.Expiry = time.Now().Local().Add(time.Hour)
}

func determineListenAddress() (string, error) { //Inorder to get the port heroku assigns us
	port := os.Getenv("PORT")
	if port == "" {
		return "", fmt.Errorf("$PORT not set")
	}
	return ":" + port, nil
}
