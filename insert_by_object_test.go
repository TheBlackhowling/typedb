package typedb

import (
	"context"
	"database/sql"
)

// InsertModel is a model for testing Insert function
type InsertModel struct {
	Model
	Name  string `db:"name"`
	Email string `db:"email"`
	ID    int64  `db:"id" load:"primary"`
}

// oracleTestWrapper wraps a DB to intercept Exec calls and populate sql.Out for testing
type oracleTestWrapper struct {
	*DB
	testID int64
}

func (w *oracleTestWrapper) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	// Find and populate sql.Out parameter before calling Exec
	for _, arg := range args {
		if out, ok := arg.(sql.Out); ok {
			if dest, ok := out.Dest.(*int64); ok {
				*dest = w.testID
			}
		}
	}
	return w.DB.Exec(ctx, query, args...)
}

func (w *oracleTestWrapper) GetDriverName() string {
	return w.DB.driverName
}

func (m *InsertModel) TableName() string {
	return "users"
}

func (m *InsertModel) QueryByID() string {
	return "SELECT id, name, email FROM users WHERE id = $1"
}

func init() {
	RegisterModel[*InsertModel]()
}

// JoinedModel is a model with dot notation (should fail Insert)
type JoinedModel struct {
	Model
	Bio    string `db:"profiles.bio"`
	UserID int    `db:"users.id" load:"primary"`
}

func (m *JoinedModel) TableName() string {
	return "users"
}

// NoTableNameModel is a model without TableName() method
type NoTableNameModel struct {
	Model
	ID int `db:"id" load:"primary"`
}

func (m *NoTableNameModel) QueryByID() string {
	return "SELECT id FROM notablenamemodel WHERE id = $1"
}

func init() {
	RegisterModel[*NoTableNameModel]()
}

// NoPrimaryKeyModel is a model without load:"primary" tag
type NoPrimaryKeyModel struct {
	Model
	ID int `db:"id"`
}

func (m *NoPrimaryKeyModel) TableName() string {
	return "users"
}

func init() {
	RegisterModel[*NoPrimaryKeyModel]()
}
