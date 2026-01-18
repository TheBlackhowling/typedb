package main

import "os"

func getTestDSN() string {
	dsn := os.Getenv("MSSQL_DSN")
	if dsn == "" {
		dsn = "server=localhost;user id=sa;password=YourPassword123;database=typedb_examples"
	}
	return dsn
}
