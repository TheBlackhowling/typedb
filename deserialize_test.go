package typedb

import (
	"time"
)

// Test models for deserialization
type DeserializeUser struct {
	Model
	CreatedAt time.Time `db:"created_at"` // 24 bytes
	Name      string    `db:"name"`       // 16 bytes
	Email     string    `db:"email"`      // 16 bytes
	ID        int       `db:"id"`         // 8 bytes
	Active    bool      `db:"active"`     // 1 byte
}

type DeserializePost struct {
	Model
	Title     string `db:"title"`     // 16 bytes
	Content   string `db:"content"`   // 16 bytes
	ID        int    `db:"id"`        // 8 bytes
	UserID    int    `db:"user_id"`   // 8 bytes
	Published bool   `db:"published"` // 1 byte
}

type DeserializeModelWithPointers struct {
	Model
	ID      *int    `db:"id"`
	Name    *string `db:"name"`
	Active  *bool   `db:"active"`
	Deleted *bool   `db:"deleted"`
}

type DeserializeModelWithArrays struct {
	Model
	Tags    []string `db:"tags"`    // 24 bytes (slice header)
	Numbers []int    `db:"numbers"` // 24 bytes (slice header)
	ID      int      `db:"id"`      // 8 bytes
}

type DeserializeModelWithJSON struct {
	Model
	Metadata map[string]any    `db:"metadata"` // 24 bytes (map header)
	Config   map[string]string `db:"config"`   // 24 bytes (map header)
	ID       int               `db:"id"`       // 8 bytes
}

type DeserializeModelWithDotNotation struct {
	Model
	Name string `db:"users.name"` // 16 bytes
	ID   int    `db:"users.id"`   // 8 bytes
}

type BaseModel struct {
	Model
	ID int `db:"id"`
}

type DerivedModel struct {
	BaseModel
	Name string `db:"name"`
}
