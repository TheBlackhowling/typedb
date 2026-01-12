module github.com/TheBlackHowling/typedb/examples/mssql

go 1.18

require (
	github.com/TheBlackHowling/typedb v0.0.0
	github.com/TheBlackHowling/typedb/examples/seed v0.0.0
	github.com/microsoft/go-mssqldb v1.6.0
)

require (
	github.com/golang-sql/civil v0.0.0-20220223132316-b832511892a9 // indirect
	github.com/golang-sql/sqlexp v0.1.0 // indirect
	github.com/jaswdr/faker v1.19.1 // indirect
	golang.org/x/crypto v0.12.0 // indirect
	golang.org/x/text v0.12.0 // indirect
)

replace github.com/TheBlackHowling/typedb => ../..

replace github.com/TheBlackHowling/typedb/examples/seed => ../seed
