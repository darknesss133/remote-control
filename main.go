package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"html/template"
	"log"

	_ "github.com/mattn/go-sqlite3"

	"github.com/valyala/fasthttp"
)

//структура для хранения данных о пользователе
type user struct {
	Id             string
	Email          string
	Verified_email bool
	Name           string
	Given_name     string
	Family_name    string
	Picture        string
	Locale         string
	IsLogged       bool
}

//структура для хранения данных о токене пользователя
type token struct {
	Access_token string
	Token_type   string
	Expires_in   float64
	Id_token     string
}

var (
	m             map[string]interface{}
	userToken     token
	userInfo      user = user{IsLogged: false}
	gResp         []byte
	queryArgs     *fasthttp.Args
	authCode      string
	addr          = flag.String("addr", ":8080", "TCP address to listen to")
	authUrl       = "https://accounts.google.com/o/oauth2/auth"
	accessUrl     = "https://www.googleapis.com/oauth2/v1/userinfo"
	tokenUrl      = "https://accounts.google.com/o/oauth2/token"
	redirect_uri  = "http://localhost" + *addr + "/oauth"
	response_type = "code"
	clientSecret  = "rERfKmVC2J0ege8bw6YAZOnU"
	clientId      = "581640295716-h3aqeb6receolg99bq82jmtumcsq1l85.apps.googleusercontent.com"
	grantType     = "authorization_code"
	scope         = "https://www.googleapis.com/auth/userinfo.email https://www.googleapis.com/auth/userinfo.profile"
)

func main() {
	flag.Parse()
	h := requestHandler
	if err := fasthttp.ListenAndServe(*addr, h); err != nil {
		log.Fatalf("Error in ListenAndServe: %s", err)
	}
}

func requestHandler(ctx *fasthttp.RequestCtx) {
	switch string(ctx.Path()) {
	case "/":
		home_page(ctx)
	case "/auth":
		oauth(ctx)
	case "/logout":
		update_status(false, userInfo.Id, ctx)
	case "/oauth":
		oauth(ctx)
	default:
		ctx.Error("Страница не найдена", fasthttp.StatusNotFound)
	}
}

func home_page(ctx *fasthttp.RequestCtx) {
	ctx.SetContentType("text/html")
	tmpl, _ := template.ParseFiles("templates/index.html")
	tmpl.Execute(ctx, userInfo)
}

func auth() {
	// queryArgs.Add("redirect_uri", redirect_uri)
	// queryArgs.Add("response_type",response_type)
	// queryArgs.Add("client_id",clientId)
	// queryArgs.Add("scope",scope)
}

func oauth(ctx *fasthttp.RequestCtx) {
	queryArgs = ctx.QueryArgs()
	authCode = string(queryArgs.Peek(response_type))
	queryArgs.Reset()
	queryArgs.Add("client_id", clientId)
	queryArgs.Add("client_secret", clientSecret)
	queryArgs.Add("redirect_uri", redirect_uri)
	queryArgs.Add("grant_type", grantType)
	queryArgs.Add("code", authCode)
	_, gResp, _ = fasthttp.Post(gResp, tokenUrl, queryArgs)
	queryArgs.Reset()
	var f interface{}
	json.Unmarshal(gResp, &f)
	m = f.(map[string]interface{})
	userToken.Access_token = m["access_token"].(string)
	_, gResp, _ = fasthttp.Get(gResp, accessUrl+"?access_token="+userToken.Access_token)
	json.Unmarshal(gResp, &f)
	m := f.(map[string]interface{})
	userInfo.Name = m["name"].(string)
	userInfo.Id = m["id"].(string)
	db, err := sql.Open("sqlite3", "golang.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	row, err := db.Query("select Access_token from users where GoogleID = $1", userInfo.Id)
	if err != nil {
		panic(err)
	}
	i := 0
	for row.Next() {
		i++
	}
	if i != 0 {
		_, err := db.Exec("update users set Access_token = $1 where GoogleID = $2", userToken.Access_token, userInfo.Id)
		if err != nil {
			panic(err)
		}
	} else {
		_, err := db.Exec("insert into users (GoogleID, Name, Access_token) values ($1, $2, $3)", userInfo.Id, userInfo.Name, userToken.Access_token)
		if err != nil {
			panic(err)
		}
	}
	update_status(true, userInfo.Id, ctx)
}

func update_status(s bool, u string, ctx *fasthttp.RequestCtx) {
	userInfo.IsLogged = s
	db, err := sql.Open("sqlite3", "golang.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	_, err = db.Exec("update users set Is_logged = $1 where GoogleID = $2", s, u)
	if err != nil {
		panic(err)
	}
	if !s {
		userInfo.Name = ""
	}
	home_page(ctx)
}
