package main

import (
	"database/sql"
	"encoding/json"
	"html/template"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"

	// needed but not directly used
	_ "github.com/go-sql-driver/mysql"
)

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

type MapTemplate struct {
	MAPDATA string
	SUPCODE string
}

type SafeMapTemplate struct {
	mu   sync.Mutex
	mapT MapTemplate
}

const messageTableName = "unread_messages"
const nudgeTableName = "user_nudge"

// NOTE: map demo has fixed data
var mapDemoTemplate MapTemplate = MapTemplate{}

var mapTemplate SafeMapTemplate = SafeMapTemplate{}

// registers the routes and handlers
func setupRoutes(e *echo.Echo, db *sql.DB, mydb DataSource) {
	initTemplateCache(db)

	e.GET("/", index)
	e.GET("/map", func(c echo.Context) error {
		mapTemplate.mu.Lock()
		mapT := mapTemplate.mapT
		mapTemplate.mu.Unlock()
		return c.Render(http.StatusOK, "map.html", mapT)
	})
	e.GET("/mapDemo", func(c echo.Context) error {
		return c.Render(
			http.StatusOK,
			"map.html",
			mapDemoTemplate)
	})
	// makes the geojson postcode data available:
	e.GET("/Postcode_Polygons/LONDON/*.geojson", func(c echo.Context) error {
		urlString := c.Request().URL.String()[1:]
		return c.File(urlString)
	})

	e.POST("/add-wellbeing-record", func(c echo.Context) error {
		record := new(WellbeingRecord)
		// bind the json body into `record`:
		if err := c.Bind(record); err != nil {
			return err
		}

		err := insertWellbeingRecord(*record, db)
		if err != nil {
			log.Print(err)
			return err
		}
		return c.JSON(http.StatusOK, map[string]bool{"success": true})
	})

	// wellbeing sharing
	e.GET("/add-friend", handleAddFriend)
	e.POST("/user", handleCheckUser(mydb))
	e.POST("/user/new", handleAddUser(mydb))
	e.POST("/user/message", handleGetMessage(mydb, messageTableName))
	e.POST("/user/message/new", handleNewMessage(mydb, messageTableName, true))

	// p2p nudge:
	// the back-end logic of passing around 'messages' is essentially the same,
	// it's up to the clients to define and handle the 'message' format.
	// Only difference is we don't overwrite pending messages.
	e.POST("/user/nudge", handleGetMessage(mydb, nudgeTableName))
	e.POST("/user/nudge/new", handleNewMessage(mydb, nudgeTableName, false))
}

func initTemplateCache(mainDb *sql.DB) {
	mockDb := getDBConn("newdatabase")
	defer mockDb.Close()
	mapDemoTemplate = *getMapTemplate(mockDb, true)

	twoMinutes := time.Duration(2) * time.Minute
	go updateTemplateCache(mainDb, twoMinutes)
}

// updates safeMapTemplate every `duration`
func updateTemplateCache(db *sql.DB, duration time.Duration) {
	mapT := getMapTemplate(db, false)

	mapTemplate.mu.Lock()
	mapTemplate.mapT = *mapT
	mapTemplate.mu.Unlock()

	time.Sleep(duration)
	updateTemplateCache(db, duration)
}

func index(c echo.Context) error {
	greet := "Greetings! You may be looking for /map"
	return c.String(http.StatusOK, greet)
}

// inserts (a copy of) wellbeing record into the database
func insertWellbeingRecord(record WellbeingRecord, db *sql.DB) error {
	query := `INSERT INTO scores` +
		` (postCode, wellbeingScore, weeklySteps, errorRate, supportCode, date_sent)` +
		` VALUES (?, ?, ?, ?, ?, ?)`
	_, err := db.Exec(query,
		record.PostCode,
		record.WellbeingScore,
		record.WeeklySteps,
		record.ErrorRate,
		record.SupportCode,
		record.DateSent) // sql automatically converts to date from yyyy-MM-dd
	return err
}

func getMapTemplate(db *sql.DB, isMock bool) *MapTemplate {
	// column names are case insensitive
	var tableName string
	if isMock {
		tableName = "MOCK_DATA"
	} else {
		tableName = "scores"
	}
	postcodeGroupQuery := "SELECT postCode as name, AVG(wellBeingScore) as " +
		"avgscore, COUNT(postcode) as quantity FROM " + tableName + " GROUP BY (Postcode)"
	suppcodeGroupQuery := "SELECT Postcode as name, SupportCode as supportcode, " +
		"AVG(WellBeingScore)as score, COUNT(SupportCode) as entries FROM " +
		tableName + " GROUP BY SupportCode, PostCode;"
	rows, err := db.Query(postcodeGroupQuery)
	if err != nil {
		log.Print(err)
	}
	defer rows.Close()
	overlayDataMapDemo := make([]map[string]interface{}, 0)
	for rows.Next() {
		var name string
		var avgscore float32
		var quantity int
		if err := rows.Scan(&name, &avgscore, &quantity); err != nil {
			log.Print(err)
		}
		data := map[string]interface{}{"name": name, "avgscore": avgscore, "quantity": quantity}
		overlayDataMapDemo = append(overlayDataMapDemo, data)
	}
	if err := rows.Err(); err != nil {
		log.Print(err)
	}

	rows2, err := db.Query(suppcodeGroupQuery)
	if err != nil {
		log.Print(err)
	}
	defer rows2.Close()
	informationMap := make([]map[string]interface{}, 0)
	for rows2.Next() {
		var name string
		var supportCode string
		var score float32
		var entries int
		if err := rows2.Scan(&name, &supportCode, &score, &entries); err != nil {
			log.Print(err)
		}
		data := map[string]interface{}{"name": name,
			"supportcode": supportCode, "score": score, "entries": entries}
		informationMap = append(informationMap, data)
	}

	mapcodeData, err := json.Marshal(overlayDataMapDemo)
	supcodeData, err2 := json.Marshal(informationMap)
	if err != nil {
		log.Print(err)
		return nil
	}
	if err2 != nil {
		log.Print(err2)
		return nil
	}
	return &MapTemplate{string(mapcodeData), string(supcodeData)}
}
