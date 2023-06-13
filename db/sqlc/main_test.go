package db

import (
	"database/sql"
	"log"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

const (
	dbDriver = "postgres"
	dbSource = "postgresql://root:secret@localhost:5432/bankdb?sslmode=disable"
)

var testQueries *Queries
var testDB *sql.DB // store result of sql.Open() command

func TestMain(m * testing.M){
	var err error
	// conn, err := sql.Open(dbDriver,dbSource)
	testDB, err = sql.Open(dbDriver,dbSource)
	if err != nil{
		log.Fatal("cannot connect to the db: ", err)
	}
	testQueries = New(testDB)
	os.Exit(m.Run())
}