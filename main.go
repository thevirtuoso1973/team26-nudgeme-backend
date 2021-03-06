package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/crypto/acme/autocert"

	// needed but not directly used
	_ "github.com/go-sql-driver/mysql"
)

const PORT = 443

// this is the address & port where the SQL database can be accessed,
// which is likely on the same machine
const ADDRESS = "178.79.172.202:3306"

var sqlPassword string = os.Getenv("SQL_PASSWORD")
var domain string = os.Getenv("DOMAIN_NAME")

func main() {
	db := getDBConn("team26")
	defer db.Close()

	var mydb DataSource = &MyDB{db}

	// setup web
	e := echo.New()

	e.AutoTLSManager.HostPolicy = autocert.HostWhitelist(domain)
	e.AutoTLSManager.Cache = autocert.DirCache("/var/www/.cache")

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.Gzip())

	setupTemplate(e)
	setupRoutes(e, db, mydb)

	// NOTE: since we are using HTTPS through Auto TLS, we have to use a
	// domain name for the server to work

	e.Logger.Fatal(e.StartAutoTLS(fmt.Sprintf(":%d", PORT)))
}

// opens and returns connection to DB
func getDBConn(dbName string) *sql.DB {
	db, err1 := sql.Open("mysql", fmt.Sprintf("root:%s@tcp(%s)/%s", sqlPassword, ADDRESS, dbName))
	if err1 != nil {
		log.Fatal(err1)
	}
	err2 := db.Ping() // validate correctly opened
	if err2 != nil {
		log.Fatal(err2)
	}
	return db
}

// registers the templates with the rendered
func setupTemplate(e *echo.Echo) {
	t := &Template{templates: template.Must(template.ParseGlob("template/*.html"))}
	e.Renderer = t
}
