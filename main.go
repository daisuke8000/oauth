package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

const (
	scope     = "https://www.googleapis.com/auth/photoslibrary.readonly"
	grantType = "authorization_code"
	verifier  = "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
)

var secret map[string]interface{}

var oauth struct {
	clientId            string
	clientSecret        string
	authEndPoint        string
	tokenEndPoint       string
	clientRedirectUri   string
	codeChallengeMethod string
	codeChallenge       string
	scope               string
	state               string
	responseType        string
}

func readJson() {
	data, err := ioutil.ReadFile("./client_secret.json")

	if err != nil {
		log.Fatal(err)
	}

	json.Unmarshal(data, &secret)
	return
}

func setUp() {

	readJson()

	oauth.clientId = secret["web"].(map[string]interface{})["client_id"].(string)
	oauth.clientSecret = secret["web"].(map[string]interface{})["client_secret"].(string)
	oauth.authEndPoint = secret["web"].(map[string]interface{})["auth_uri"].(string)
	oauth.tokenEndPoint = secret["web"].(map[string]interface{})["token_uri"].(string)
	oauth.clientRedirectUri = secret["web"].(map[string]interface{})["redirect_uris"].([]interface{})[0].(string)
	oauth.scope = scope
	oauth.state = "xyz"
	oauth.responseType = "token"
	oauth.codeChallengeMethod = "S256"
	oauth.codeChallenge = base64URLEncode()
}

func base64URLEncode() string {
	hash := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}

func start(w http.ResponseWriter, r *http.Request) {

	authEndPoint := oauth.authEndPoint

	values := url.Values{}
	values.Add("response_type", oauth.responseType)
	values.Add("client_id", oauth.clientId)
	values.Add("state", oauth.state)
	values.Add("scope", oauth.scope)
	values.Add("redirect_uri", oauth.clientRedirectUri)
	values.Add("code_challenge_method", oauth.codeChallengeMethod)
	values.Add("code_challenge", oauth.codeChallenge)

	http.Redirect(w, r, authEndPoint+"?"+values.Encode(), http.StatusFound)
}

func callback(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	res, err := tokenRequest(query)
	if err != nil {
		log.Println(err)
	}

	body, err := apiRequest(r, res["access_token"].(string))
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
	values.Add("code", q.Get("code"))
	values.Add("redirect_uri", oauth.clientRedirectUri)

	req, err := http.NewRequest("POST", tokenEndPoint, strings.NewReader(values.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	log.Printf("token response : %s", string(body))
	var data map[string]interface{}
	json.Unmarshal(body, &data)
	return data, nil
}

func apiRequest(r *http.Request, token string) ([]byte, error) {

	photoAPI := "https://photoslibrary.googleapis.com/v1/mediaItems"

	r, err := http.NewRequest("GET", photoAPI, nil)
	if err != nil {
		return nil, err
	}

	r.Header.Set("Authorization", fmt.Sprintf("Bearer "+token))
	client := &http.Client{}
	resp, err := client.Do(r)
	if err != nil || resp.StatusCode != 200 {
		log.Printf("http status code is %d, err: %s", resp.StatusCode, err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return body, nil
}

//OAuth2Client is a client
func main() {

	setUp()

	http.HandleFunc("/start", start)
	http.HandleFunc("/callback", callback)
	err := http.ListenAndServe("localhost:8080", nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Server started, localhost:8080")
}
