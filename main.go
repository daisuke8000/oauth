package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

const (
	scope     = "https://www.googleapis.com/auth/photoslibrary.readonly"
	grantType = "authorization_code"
)

var secret map[string]interface{}

var oauth struct {
	clientId          string
	clientSecret      string
	authEndPoint      string
	tokenEndPoint     string
	clientRedirectUri string
	scope             string
	state             string
	responseType      string
}

func readJson() {
	data, err := ioutil.ReadFile("/client_secret.json")

	if err != nil {
		log.Fatal(err)
	}

	json.Unmarshal(data, &secret)
	return
}

func setUp() {

	readJson()

	oauth.clientId = secret["web"].(map[string]interface{})["clientId"].(string)
	oauth.clientSecret = secret["web"].(map[string]interface{})["clientSecret"].(string)
	oauth.authEndPoint = secret["web"].(map[string]interface{})["authEndPoint"].(string)
	oauth.tokenEndPoint = secret["web"].(map[string]interface{})["tokenEndPoint"].(string)
	oauth.scope = scope
	oauth.state = "xyz"
	oauth.responseType = "code"
}

func start(w http.ResponseWriter, r *http.Request) {

	authEndPoint := oauth.authEndPoint

	values := url.Values{}
	values.Add("response_type", oauth.responseType)
	values.Add("client_id", oauth.clientId)
	values.Add("state", oauth.state)
	values.Add("scope", oauth.scope)
	values.Add("redirect_uri", oauth.clientRedirectUri)

	http.Redirect(w, r, authEndPoint+"?"+values.Encode(), http.StatusFound)
}

func callback(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	res, err := tokenRequest(query)
	if err != nil {
		log.Println(err)
	}

	body, err := apiRequest(r, res["accsess_token"].(string))
	if err != nil {
		log.Println(err)
	}
	w.Write(body)

}

func tokenRequest(q url.Values) (map[string]interface{}, error) {

	tokenEndPoint := oauth.tokenEndPoint
	values := url.Values{}
	values.Add("client_id", oauth.clientId)
	values.Add("client_secret", oauth.clientSecret)
	values.Add("grantType", grantType)

}

//OAuth2Client is a client
func main() {

	setUp()

	http.HandleFunc("/start", start)
	http.HandleFunc("/callback", callback)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Server started, localhost:8080")
}
