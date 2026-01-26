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
	Tags      string `db:"tags"`     // MySQL JSON as string
	Metadata  string `db:"metadata"` // MySQL JSON as string
	CreatedAt string `db:"created_at"`
	UpdatedAt string `db:"updated_at" dbUpdate:"auto-timestamp"`
}

func (p *Post) TableName() string {
	return "posts"
}

func (p *Post) QueryByID() string {
	return "SELECT id, user_id, title, content, tags, metadata, created_at, updated_at FROM posts WHERE id = ?"
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

// TypeExample demonstrates comprehensive MySQL data types
type TypeExample struct {
	typedb.Model
	ID                    int    `db:"id" load:"primary"`
	TinyInt               int    `db:"tiny_int"`
	TinyIntUnsigned       uint   `db:"tiny_int_unsigned"`
	SmallInt              int    `db:"small_int"`
	SmallIntUnsigned      uint   `db:"small_int_unsigned"`
	MediumInt             int    `db:"medium_int"`
	MediumIntUnsigned     uint   `db:"medium_int_unsigned"`
	IntegerCol            int    `db:"integer_col"`
	IntegerColUnsigned    uint   `db:"integer_col_unsigned"`
	BigInt                int64  `db:"big_int"`
	BigIntUnsigned        uint64 `db:"big_int_unsigned"` // MySQL returns unsigned BIGINT as string, handled by DeserializeUint64
	DecimalCol            string `db:"decimal_col"`
	DecimalColUnsigned    string `db:"decimal_col_unsigned"`
	NumericCol            string `db:"numeric_col"`
	NumericColUnsigned    string `db:"numeric_col_unsigned"`
	FloatCol              string `db:"float_col"`
	FloatColPrecision     string `db:"float_col_precision"`
	DoubleCol             string `db:"double_col"`
	DoubleColPrecision    string `db:"double_col_precision"`
	BitCol                string `db:"bit_col"`
	BitCol64              string `db:"bit_col_64"`
	CharCol               string `db:"char_col"`
	VarcharCol            string `db:"varchar_col"`
	BinaryCol             string `db:"binary_col"`
	VarbinaryCol          string `db:"varbinary_col"`
	TinytextCol           string `db:"tinytext_col"`
	TextCol               string `db:"text_col"`
	MediumtextCol         string `db:"mediumtext_col"`
	LongtextCol           string `db:"longtext_col"`
	EnumCol               string `db:"enum_col"`
	SetCol                string `db:"set_col"`
	TinyblobCol           string `db:"tinyblob_col"`
	BlobCol               string `db:"blob_col"`
	MediumblobCol         string `db:"mediumblob_col"`
	LongblobCol           string `db:"longblob_col"`
	DateCol               string `db:"date_col"`
	TimeCol               string `db:"time_col"`
	DatetimeCol           string `db:"datetime_col"`
	TimestampCol          string `db:"timestamp_col"`
	YearCol               int    `db:"year_col"`
	JsonCol               string `db:"json_col"`
	GeometryCol           string `db:"geometry_col"`
	PointCol              string `db:"point_col"`
	LinestringCol         string `db:"linestring_col"`
	PolygonCol            string `db:"polygon_col"`
	MultipointCol         string `db:"multipoint_col"`
	MultilinestringCol    string `db:"multilinestring_col"`
	MultipolygonCol       string `db:"multipolygon_col"`
	GeometrycollectionCol string `db:"geometrycollection_col"`
	CreatedAt             string `db:"created_at"`
}

// Deserialize uses Model.Deserialize which handles deserialization
// No custom implementation needed since we embed Model

func (t *TypeExample) QueryByID() string {
	return "SELECT id, tiny_int, tiny_int_unsigned, small_int, small_int_unsigned, medium_int, medium_int_unsigned, integer_col, integer_col_unsigned, big_int, big_int_unsigned, decimal_col, decimal_col_unsigned, numeric_col, numeric_col_unsigned, float_col, float_col_precision, double_col, double_col_precision, bit_col, bit_col_64, char_col, varchar_col, binary_col, varbinary_col, tinytext_col, text_col, mediumtext_col, longtext_col, enum_col, set_col, tinyblob_col, blob_col, mediumblob_col, longblob_col, date_col, time_col, datetime_col, timestamp_col, year_col, json_col, geometry_col, point_col, linestring_col, polygon_col, multipoint_col, multilinestring_col, multipolygon_col, geometrycollection_col, created_at FROM type_examples WHERE id = ?"
}

func init() {
	typedb.RegisterModel[*TypeExample]()
}
