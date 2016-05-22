package main

import (
	"database/sql"
	_ "github.com/lib/pq"
	"net/http/httptest"
	"os"
	"testing"
	"log"
	"io/ioutil"
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

	a := application{id: "test-app"}
	if r:= a.del();!r.Ok {
		t.Error(r.Msg)
	}

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

