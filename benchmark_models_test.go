package typedb

import (
	"time"
)

// SimpleUser represents a simple user model for performance benchmarking.
// Used to measure deserialization overhead with basic field types.
type SimpleUser struct {
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
	Model
	Name   string  `db:"name"`
	Email  string  `db:"email" load:"unique"`
	ID     int64   `db:"id" load:"primary"`
	Score  float64 `db:"score"`
	Age    int     `db:"age"`
	Active bool    `db:"active"`
}

// TableName returns the table name for SimpleUser.
func (u *SimpleUser) TableName() string {
	return "simple_users"
}

// QueryByID returns the SQL query for loading a user by ID.
func (u *SimpleUser) QueryByID() string {
	return "SELECT id, name, email, created_at, updated_at, active, age, score FROM simple_users WHERE id = $1"
}

// QueryByEmail returns the SQL query for loading a user by email.
func (u *SimpleUser) QueryByEmail() string {
	return "SELECT id, name, email, created_at, updated_at, active, age, score FROM simple_users WHERE email = $1"
}

// BenchmarkAddress represents an address nested within ComplexUser.
type BenchmarkAddress struct {
	Street  string `json:"street"`
	City    string `json:"city"`
	State   string `json:"state"`
	ZipCode string `json:"zip_code"`
}

// BenchmarkRole represents a role nested within ComplexUser.
type BenchmarkRole struct {
	GrantedAt   time.Time `json:"granted_at"`
	Name        string    `json:"name"`
	Permissions []string  `json:"permissions"`
	ID          int64     `json:"id"`
}

// BenchmarkSettings represents settings nested within ComplexUser.
type BenchmarkSettings struct {
	Notifications map[string]bool `json:"notifications"`
	Preferences   map[string]any  `json:"preferences"`
	Theme         string          `json:"theme"`
}

// ComplexUser represents a complex user model for performance benchmarking.
// Used to measure deserialization overhead with advanced field types including
// JSONB, arrays, nested structs, and nullable fields.
type ComplexUser struct {
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
	Model
	Metadata    map[string]interface{} `db:"metadata"`
	Balance     *float64               `db:"balance"`
	Settings    *BenchmarkSettings     `db:"settings"`
	LastLogin   *time.Time             `db:"last_login"`
	Preferences map[string]string      `db:"preferences"`
	Address     BenchmarkAddress       `db:"address"`
	StatusCode  string                 `db:"status_code" load:"composite:user_status"`
	Status      string                 `db:"status"`
	Notes       string                 `db:"notes"`
	Email       string                 `db:"email" load:"unique"`
	Name        string                 `db:"name"`
	Tags        []string               `db:"tags"`
	Roles       []BenchmarkRole        `db:"roles"`
	UserID      int64                  `db:"user_id" load:"composite:user_status"`
	Score       float64                `db:"score"`
	ID          int64                  `db:"id" load:"primary"`
	Age         int                    `db:"age"`
	Active      bool                   `db:"active"`
}

// TableName returns the table name for ComplexUser.
func (u *ComplexUser) TableName() string {
	return "complex_users"
}

// QueryByID returns the SQL query for loading a user by ID.
func (u *ComplexUser) QueryByID() string {
	return "SELECT id, name, email, created_at, updated_at, active, age, score, metadata, tags, preferences, address, roles, settings, last_login, balance, status, notes, user_id, status_code FROM complex_users WHERE id = $1"
}

// QueryByEmail returns the SQL query for loading a user by email.
func (u *ComplexUser) QueryByEmail() string {
	return "SELECT id, name, email, created_at, updated_at, active, age, score, metadata, tags, preferences, address, roles, settings, last_login, balance, status, notes, user_id, status_code FROM complex_users WHERE email = $1"
}

// QueryByUserIDStatusCode returns the SQL query for loading a user by composite key.
// Fields are sorted alphabetically: StatusCode, UserID
func (u *ComplexUser) QueryByUserIDStatusCode() string {
	return "SELECT id, name, email, created_at, updated_at, active, age, score, metadata, tags, preferences, address, roles, settings, last_login, balance, status, notes, user_id, status_code FROM complex_users WHERE status_code = $1 AND user_id = $2"
}
