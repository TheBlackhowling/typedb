module github.com/TheBlackHowling/typedb/examples/mysql

go 1.23

replace github.com/TheBlackHowling/typedb => ../..

require (
	github.com/TheBlackHowling/typedb v0.0.0-00010101000000-000000000000
	github.com/go-sql-driver/mysql v1.9.3
)

require filippo.io/edwards25519 v1.1.0 // indirect
