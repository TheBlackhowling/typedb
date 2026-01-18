package main

import "os"

func getTestDSN() string {
	dsn := os.Getenv("POSTGRES_DSN")
	if dsn == "" {
		dsn = "postgres://user:password@localhost/typedb_examples?sslmode=disable"
	}
	return dsn
}
