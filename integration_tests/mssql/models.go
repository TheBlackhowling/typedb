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
	return "SELECT id, name, email, created_at, updated_at FROM users WHERE id = @p1"
}

func (u *User) QueryByEmail() string {
	return "SELECT id, name, email, created_at, updated_at FROM users WHERE email = @p1"
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
	Tags      string `db:"tags"`     // JSON stored as NVARCHAR(MAX)
	Metadata  string `db:"metadata"` // JSON stored as NVARCHAR(MAX)
	CreatedAt string `db:"created_at"`
	UpdatedAt string `db:"updated_at" dbUpdate:"auto-timestamp"`
}

func (p *Post) TableName() string {
	return "posts"
}

func (p *Post) QueryByID() string {
	return "SELECT id, user_id, title, content, tags, metadata, created_at FROM posts WHERE id = @p1"
}

func (p *Post) QueryByUserID() string {
	return "SELECT id, user_id, title, content, tags, metadata, created_at, updated_at FROM posts WHERE user_id = @p1"
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
	return "SELECT user_id, post_id, favorited_at FROM user_posts WHERE post_id = @p1 AND user_id = @p2"
}

func init() {
	typedb.RegisterModel[*UserPost]()
}

// TypeExample demonstrates comprehensive SQL Server data types
type TypeExample struct {
	typedb.Model
	ID                  int    `db:"id" load:"primary"`
	TinyInt             int    `db:"tiny_int"`
	SmallInt            int    `db:"small_int"`
	IntegerCol          int    `db:"integer_col"`
	BigInt              int64  `db:"big_int"`
	DecimalCol          string `db:"decimal_col"`
	NumericCol          string `db:"numeric_col"`
	MoneyCol            string `db:"money_col"`
	SmallmoneyCol       string `db:"smallmoney_col"`
	BitCol              bool   `db:"bit_col"`
	FloatCol            string `db:"float_col"`
	RealCol             string `db:"real_col"`
	CharCol             string `db:"char_col"`
	VarcharCol          string `db:"varchar_col"`
	VarcharMaxCol       string `db:"varchar_max_col"`
	NcharCol            string `db:"nchar_col"`
	NvarcharCol         string `db:"nvarchar_col"`
	NvarcharMaxCol      string `db:"nvarchar_max_col"`
	TextCol             string `db:"text_col"`
	NtextCol            string `db:"ntext_col"`
	BinaryCol           string `db:"binary_col"`
	VarbinaryCol        string `db:"varbinary_col"`
	VarbinaryMaxCol     string `db:"varbinary_max_col"`
	ImageCol            string `db:"image_col"`
	DateCol             string `db:"date_col"`
	TimeCol             string `db:"time_col"`
	DatetimeCol         string `db:"datetime_col"`
	Datetime2Col        string `db:"datetime2_col"`
	DatetimeoffsetCol   string `db:"datetimeoffset_col"`
	SmalldatetimeCol    string `db:"smalldatetime_col"`
	TimestampCol        string `db:"timestamp_col"`
	UniqueidentifierCol string `db:"uniqueidentifier_col"`
	XmlCol              string `db:"xml_col"`
	HierarchyidCol      string `db:"hierarchyid_col"`
	GeographyCol        string `db:"geography_col"`
	GeometryCol         string `db:"geometry_col"`
	SqlVariantCol       string `db:"sql_variant_col"`
	CreatedAt           string `db:"created_at"`
}

func (t *TypeExample) QueryByID() string {
	return "SELECT id, tiny_int, small_int, integer_col, big_int, decimal_col, numeric_col, money_col, smallmoney_col, bit_col, float_col, real_col, char_col, varchar_col, varchar_max_col, nchar_col, nvarchar_col, nvarchar_max_col, text_col, ntext_col, binary_col, varbinary_col, varbinary_max_col, image_col, date_col, time_col, datetime_col, datetime2_col, datetimeoffset_col, smalldatetime_col, timestamp_col, uniqueidentifier_col, xml_col, hierarchyid_col, geography_col, geometry_col, sql_variant_col, created_at FROM type_examples WHERE id = @p1"
}

func init() {
	typedb.RegisterModel[*TypeExample]()
}
