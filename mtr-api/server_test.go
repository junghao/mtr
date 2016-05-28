package main

import (
	"database/sql"
	"github.com/GeoNet/weft"
	_ "github.com/lib/pq"
	"io/ioutil"
	"log"
	"net/http/httptest"
	"os"
	"testing"
)

var testServer *httptest.Server

func setup(t *testing.T) {
	var err error
	if db, err = sql.Open("postgres",
		os.ExpandEnv("host=${DB_HOST} connect_timeout=30 user=${DB_USER} password=${DB_PASSWORD} dbname=mtr sslmode=disable")); err != nil {
		t.Fatal(err)
	}

	if err = db.Ping(); err != nil {
		t.Fatal(err)
	}

	dbR, err = sql.Open("postgres",
		os.ExpandEnv("host=${DB_HOST} connect_timeout=30 user=${DB_USER_R} password=${DB_PASSWORD_R} dbname=mtr sslmode=disable"))
	if err != nil {
		t.Fatal(err)
	}

	if err = dbR.Ping(); err != nil {
		t.Fatal(err)
	}

	// Only needed when testing map180 calls (which depends on loading the map data)
	//wm, err = map180.Init(dbR, map180.Region(`newzealand`), 256000000)
	//if err != nil {
	//	t.Fatalf("ERROR: problem with map180 config: %s", err)
	//}

	testServer = httptest.NewServer(inbound(mux))

	// Silence the logging unless running with
	// go test -v
	if !testing.Verbose() {
		log.SetOutput(ioutil.Discard)
	}

	if r := delApplication("test-app"); !r.Ok {
		t.Error(r.Msg)
	}

}

func delApplication(applicationID string) *weft.Result {
	if applicationID == "" {
		return weft.BadRequest("empty applicationID")
	}

	if _, err := db.Exec(`DELETE FROM app.application WHERE applicationID = $1`, applicationID); err != nil {
		return weft.InternalServerError(err)
	}

	return &weft.StatusOK
}

func setupBench(t *testing.B) {
	var err error
	if db, err = sql.Open("postgres",
		os.ExpandEnv("host=${DB_HOST} connect_timeout=30 user=${DB_USER} password=${DB_PASSWORD} dbname=mtr sslmode=disable")); err != nil {
		t.Fatal(err)
	}

	if err = db.Ping(); err != nil {
		t.Fatal(err)
	}

	dbR, err = sql.Open("postgres",
		os.ExpandEnv("host=${DB_HOST} connect_timeout=30 user=${DB_USER_R} password=${DB_PASSWORD_R} dbname=mtr sslmode=disable"))
	if err != nil {
		t.Fatal(err)
	}

	if err = dbR.Ping(); err != nil {
		t.Fatal(err)
	}
}

func teardown() {
	testServer.Close()
	db.Close()
	defer dbR.Close()
}
