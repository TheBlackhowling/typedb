package seed

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
}

func (u *User) TableName() string {
	return "users"
}

func (u *User) QueryByID() string {
	return "SELECT id, name, email, created_at FROM users WHERE id = $1"
}

func (u *User) QueryByEmail() string {
	return "SELECT id, name, email, created_at FROM users WHERE email = $1"
}

func init() {
	typedb.RegisterModel[*User]()
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
	return "SELECT id, user_id, bio, avatar_url, location, website, created_at FROM profiles WHERE id = $1"
}

func (p *Profile) QueryByUserID() string {
	return "SELECT id, user_id, bio, avatar_url, location, website, created_at FROM profiles WHERE user_id = $1"
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
	return "SELECT id, user_id, title, content, published, created_at FROM posts WHERE id = $1"
}

func init() {
	typedb.RegisterModel[*Post]()
}
