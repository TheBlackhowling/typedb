package main

import "os"

func getTestDSN() string {
	dsn := os.Getenv("ORACLE_DSN")
	if dsn == "" {
		dsn = "oracle://user:password@localhost:1521/XE"
	}
	return dsn
}
