package main

import (
	"github.com/TheBlackHowling/typedb"
)

// User represents a user in the database
type User struct {
	typedb.Model
	ID        int    `db:"id" load:"primary"`
	Name      string `db:"name"`
	Email     string `db:"email" load:"unique"`
	CreatedAt string `db:"created_at"`
	UpdatedAt string `db:"updated_at" dbUpdate:"auto-timestamp"`
}

func (u *User) TableName() string {
	return "users"
}

func (u *User) QueryByID() string {
	return "SELECT id, name, email, created_at, updated_at FROM users WHERE id = ?"
}

func (u *User) QueryByEmail() string {
	return "SELECT id, name, email, created_at, updated_at FROM users WHERE email = ?"
}

func init() {
	typedb.RegisterModel[*User]()
	// Register User with partial update enabled for testing
	typedb.RegisterModelWithOptions[*User](typedb.ModelOptions{PartialUpdate: true})
}

// Post represents a blog post
type Post struct {
	typedb.Model
	ID        int    `db:"id" load:"primary"`
	UserID    int    `db:"user_id"`
	Title     string `db:"title"`
	Content   string `db:"content"`
	Tags      string `db:"tags"`     // JSON stored as TEXT
	Metadata  string `db:"metadata"` // JSON stored as TEXT
	CreatedAt string `db:"created_at"`
	UpdatedAt string `db:"updated_at" dbUpdate:"auto-timestamp"`
}

func (p *Post) TableName() string {
	return "posts"
}

func (p *Post) QueryByID() string {
	return "SELECT id, user_id, title, content, tags, metadata, created_at FROM posts WHERE id = ?"
}

func (p *Post) QueryByUserID() string {
	return "SELECT id, user_id, title, content, tags, metadata, created_at, updated_at FROM posts WHERE user_id = ?"
}

func init() {
	typedb.RegisterModel[*Post]()
}

// UserPost represents the many-to-many relationship between users and posts
type UserPost struct {
	typedb.Model
	UserID      int    `db:"user_id" load:"composite:userpost"`
	PostID      int    `db:"post_id" load:"composite:userpost"`
	FavoritedAt string `db:"favorited_at"`
}

func (up *UserPost) TableName() string {
	return "user_posts"
}

// QueryByPostIDUserID - fields sorted alphabetically: PostID, UserID
func (up *UserPost) QueryByPostIDUserID() string {
	return "SELECT user_id, post_id, favorited_at FROM user_posts WHERE post_id = ? AND user_id = ?"
}

func init() {
	typedb.RegisterModel[*UserPost]()
}

// TypeExample demonstrates comprehensive SQLite data types
type TypeExample struct {
	typedb.Model
	ID           int    `db:"id" load:"primary"`
	IntegerCol   int    `db:"integer_col"`
	RealCol      string `db:"real_col"`
	NumericCol   string `db:"numeric_col"`
	TextCol      string `db:"text_col"`
	VarcharCol   string `db:"varchar_col"`
	CharCol      string `db:"char_col"`
	ClobCol      string `db:"clob_col"`
	BlobCol      string `db:"blob_col"` // Binary as hex string
	DateCol      string `db:"date_col"`
	DatetimeCol  string `db:"datetime_col"`
	TimestampCol string `db:"timestamp_col"`
	TimeCol      string `db:"time_col"`
	BooleanCol   bool   `db:"boolean_col"`
	JsonCol      string `db:"json_col"`
	CreatedAt    string `db:"created_at"`
}

func (t *TypeExample) QueryByID() string {
	return "SELECT id, integer_col, real_col, numeric_col, text_col, varchar_col, char_col, clob_col, blob_col, date_col, datetime_col, timestamp_col, time_col, boolean_col, json_col, created_at FROM type_examples WHERE id = ?"
}

func init() {
	typedb.RegisterModel[*TypeExample]()
}
