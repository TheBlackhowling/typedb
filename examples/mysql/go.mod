module github.com/TheBlackHowling/typedb/examples/mysql

go 1.18

require (
	github.com/TheBlackHowling/typedb v0.0.0
	github.com/TheBlackHowling/typedb/examples/seed v0.0.0
	github.com/go-sql-driver/mysql v1.7.1
)

require github.com/jaswdr/faker v1.19.1 // indirect

replace github.com/TheBlackHowling/typedb => ../..

replace github.com/TheBlackHowling/typedb/examples/seed => ../seed
