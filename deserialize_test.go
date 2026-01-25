package typedb

import (
	"time"
)

// Test models for deserialization
type DeserializeUser struct {
	Model
	ID        int       `db:"id"`
	Name      string    `db:"name"`
	Email     string    `db:"email"`
	Active    bool      `db:"active"`
	CreatedAt time.Time `db:"created_at"`
}

type DeserializePost struct {
	Model
	ID        int    `db:"id"`
	UserID    int    `db:"user_id"`
	Title     string `db:"title"`
	Content   string `db:"content"`
	Published bool   `db:"published"`
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
	ID      int      `db:"id"`
	Tags    []string `db:"tags"`
	Numbers []int    `db:"numbers"`
}

type DeserializeModelWithJSON struct {
	Model
	ID      int            `db:"id"`
	Metadata map[string]any `db:"metadata"`
	Config   map[string]string `db:"config"`
}

type DeserializeModelWithDotNotation struct {
	Model
	ID   int    `db:"users.id"`
	Name string `db:"users.name"`
}

type BaseModel struct {
	Model
	ID int `db:"id"`
}

type DerivedModel struct {
	BaseModel
	Name string `db:"name"`
}

