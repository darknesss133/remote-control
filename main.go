package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"

	"github.com/valyala/fasthttp"
)

//структура для хранения данных о пользователе
type user struct {
	id             string
	email          string
	verified_email bool
	name           string
	given_name     string
	family_name    string
	picture        string
	locale         string
}

//структура для хранения данных о токене пользователя
type token struct {
	access_token string
	token_type   string
	expires_in   float64
	id_token     string
}

var (
	m            map[string]interface{}
	userToken    token
	userInfo     user
	gResp        []byte
	queryArgs    *fasthttp.Args
	authCode     string
	addr         = flag.String("addr", ":8080", "TCP address to listen to")
	authUrl      = "https://accounts.google.com/o/oauth2/auth"
	accessUrl    = "https://www.googleapis.com/oauth2/v1/userinfo"
	tokenUrl     = "https://accounts.google.com/o/oauth2/token"
	clientSecret = "rERfKmVC2J0ege8bw6YAZOnU"
	clientId     = "581640295716-h3aqeb6receolg99bq82jmtumcsq1l85.apps.googleusercontent.com"
	grantType    = "authorization_code"
	scope        = "https://www.googleapis.com/auth/userinfo.email https://www.googleapis.com/auth/userinfo.profile"
)

func main() {
	flag.Parse()
	h := requestHandler
	if err := fasthttp.ListenAndServe(*addr, h); err != nil {
		log.Fatalf("Error in ListenAndServe: %s", err)
	}
}

func requestHandler(ctx *fasthttp.RequestCtx) {
	queryArgs = ctx.QueryArgs()
	authCode = string(queryArgs.Peek("code"))
	if authCode != "" {
		queryArgs.Reset()
		queryArgs.Add("client_id", clientId)
		queryArgs.Add("client_secret", clientSecret)
		queryArgs.Add("redirect_uri", "http://localhost"+*addr)
		queryArgs.Add("grant_type", grantType)
		queryArgs.Add("code", authCode)
		_, gResp, _ = fasthttp.Post(gResp, tokenUrl, queryArgs)
		queryArgs.Reset()
		var f interface{}
		json.Unmarshal(gResp, &f)
		m = f.(map[string]interface{})
		userToken.access_token = m["access_token"].(string)
		_, gResp, _ = fasthttp.Get(gResp, accessUrl+"?access_token="+userToken.access_token)
		json.Unmarshal(gResp, &f)
		m := f.(map[string]interface{})
		userInfo.name = m["name"].(string)
		fmt.Fprintf(ctx, "Hello %s!", userInfo.name)
	}
}
