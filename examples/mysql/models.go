package main

import (
	"github.com/TheBlackHowling/typedb"
)

// User represents a user in the database
type User struct {
	typedb.Model
	ID        int     `db:"id" load:"primary"`
	Name      string  `db:"name"`
	Email     string  `db:"email" load:"unique"`
	Phone     *string `db:"phone"` // Optional phone number (nullable)
	CreatedAt string  `db:"created_at"`
}

func (u *User) TableName() string {
	return "users"
}

func (u *User) QueryByID() string {
	return "SELECT id, name, email, phone, created_at FROM users WHERE id = ?"
}

func (u *User) QueryByEmail() string {
	return "SELECT id, name, email, phone, created_at FROM users WHERE email = ?"
}

func init() {
	typedb.RegisterModel[*User]()
	// Register User with partial update enabled for examples
	typedb.RegisterModelWithOptions[*User](typedb.ModelOptions{PartialUpdate: true})
}

// Profile represents a user profile (one-to-one with User)
type Profile struct {
	typedb.Model
	ID        int    `db:"id" load:"primary"`
	UserID    int    `db:"user_id"`
	Bio       string `db:"bio"`
	AvatarURL string `db:"avatar_url"`
	Location  string `db:"location"`
	Website   string `db:"website"`
	CreatedAt string `db:"created_at"`
}

func (p *Profile) TableName() string {
	return "profiles"
}

func (p *Profile) QueryByID() string {
	return "SELECT id, user_id, bio, avatar_url, location, website, created_at FROM profiles WHERE id = ?"
}

func (p *Profile) QueryByUserID() string {
	return "SELECT id, user_id, bio, avatar_url, location, website, created_at FROM profiles WHERE user_id = ?"
}

func init() {
	typedb.RegisterModel[*Profile]()
}

// Post represents a blog post (many-to-one with User)
type Post struct {
	typedb.Model
	ID        int    `db:"id" load:"primary"`
	UserID    int    `db:"user_id"`
	Title     string `db:"title"`
	Content   string `db:"content"`
	Published bool   `db:"published"`
	CreatedAt string `db:"created_at"`
}

func (p *Post) TableName() string {
	return "posts"
}

func (p *Post) QueryByID() string {
	return "SELECT id, user_id, title, content, published, created_at FROM posts WHERE id = ?"
}

func init() {
	typedb.RegisterModel[*Post]()
}

// UserPost represents a many-to-many relationship between users and posts (favorites/bookmarks)
type UserPost struct {
	typedb.Model
	UserID      int    `db:"user_id" load:"composite:userpost"`
	PostID      int    `db:"post_id" load:"composite:userpost"`
	FavoritedAt string `db:"favorited_at"`
}

func (up *UserPost) TableName() string {
	return "user_posts"
}

// QueryByPostIDUserID returns the query for composite key lookup
// Note: Field names are sorted alphabetically: PostID comes before UserID
func (up *UserPost) QueryByPostIDUserID() string {
	return "SELECT user_id, post_id, favorited_at FROM user_posts WHERE post_id = ? AND user_id = ?"
}

func init() {
	typedb.RegisterModel[*UserPost]()
}
