module github.com/TheBlackHowling/typedb/integration_tests/sqlite

go 1.24.0

toolchain go1.24.12

replace github.com/TheBlackHowling/typedb => ../..

require (
	github.com/TheBlackHowling/typedb v0.0.0-00010101000000-000000000000
	github.com/golang-migrate/migrate/v4 v4.19.1
	github.com/mattn/go-sqlite3 v1.14.33
)

require (
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	go.uber.org/atomic v1.7.0 // indirect
)
