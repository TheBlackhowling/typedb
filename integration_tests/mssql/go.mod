module github.com/TheBlackHowling/typedb/examples/mssql

go 1.24.0

toolchain go1.24.11

replace github.com/TheBlackHowling/typedb => ../..

require (
	github.com/TheBlackHowling/typedb v0.0.0-00010101000000-000000000000
	github.com/microsoft/go-mssqldb v1.9.6
)

require (
	github.com/golang-sql/civil v0.0.0-20220223132316-b832511892a9 // indirect
	github.com/golang-sql/sqlexp v0.1.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/shopspring/decimal v1.4.0 // indirect
	golang.org/x/crypto v0.45.0 // indirect
	golang.org/x/text v0.31.0 // indirect
)
