package typedb

import (
	"time"
)

// SimpleUser represents a simple user model for performance benchmarking.
// Used to measure deserialization overhead with basic field types.
type SimpleUser struct {
	Model
	ID        int64     `db:"id" load:"primary"`
	Score     float64   `db:"score"`
	Age       int       `db:"age"`
	Name      string    `db:"name"`
	Email     string    `db:"email" load:"unique"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
	Active    bool      `db:"active"`
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
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Permissions []string  `json:"permissions"`
	GrantedAt   time.Time `json:"granted_at"`
}

// BenchmarkSettings represents settings nested within ComplexUser.
type BenchmarkSettings struct {
	Theme         string          `json:"theme"`
	Notifications map[string]bool `json:"notifications"`
	Preferences   map[string]any  `json:"preferences"`
}

// ComplexUser represents a complex user model for performance benchmarking.
// Used to measure deserialization overhead with advanced field types including
// JSONB, arrays, nested structs, and nullable fields.
type ComplexUser struct {
	Model
	ID          int64                  `db:"id" load:"primary"`
	Score       float64                `db:"score"`
	UserID      int64                  `db:"user_id" load:"composite:user_status"`
	Balance     *float64               `db:"balance"`    // Nullable float
	Settings    *BenchmarkSettings     `db:"settings"`   // Pointer to struct (JSONB, nullable)
	LastLogin   *time.Time             `db:"last_login"` // Nullable time
	Age         int                    `db:"age"`
	Name        string                 `db:"name"`
	Email       string                 `db:"email" load:"unique"`
	Status      string                 `db:"status"`
	Notes       string                 `db:"notes"`
	StatusCode  string                 `db:"status_code" load:"composite:user_status"`
	CreatedAt   time.Time              `db:"created_at"`
	UpdatedAt   time.Time              `db:"updated_at"`
	Metadata    map[string]interface{} `db:"metadata"`    // JSONB field
	Tags        []string               `db:"tags"`        // Array field
	Preferences map[string]string      `db:"preferences"` // Map field
	Address     BenchmarkAddress       `db:"address"`     // Nested struct (JSONB)
	Roles       []BenchmarkRole        `db:"roles"`       // Array of structs (JSONB)
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
