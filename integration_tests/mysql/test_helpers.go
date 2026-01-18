package main

import "os"

func getTestDSN() string {
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		dsn = "user:password@tcp(localhost:3306)/typedb_examples?parseTime=true"
	}
	return dsn
}
