# API

## Requirements
* Go version 1.12.7 or later

## Setup
* Fill in chat-overlay-api.json with callback url, client id and client secret.

## Available endpoints
* GET - /login (Redirects user to spotify oauth login page and returns to callback page. Optional "?returnurl=" Which redirects the user to given url. If returnurl is empty the user will be presented with the token on the screen.)
* GET - /session/{token} (Parameter is the token paramert generated on logon. Returns the Spotify token.)
* GET - /callback (Only used by the spotify as a way to return data to the api.)
