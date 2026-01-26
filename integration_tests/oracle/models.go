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
	return "SELECT id, name, email, created_at, updated_at FROM users WHERE id = :1"
}

func (u *User) QueryByEmail() string {
	return "SELECT id, name, email, created_at, updated_at FROM users WHERE email = :1"
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
	Tags      string `db:"tags"`     // JSON stored as CLOB
	Metadata  string `db:"metadata"` // JSON stored as CLOB
	CreatedAt string `db:"created_at"`
	UpdatedAt string `db:"updated_at" dbUpdate:"auto-timestamp"`
}

func (p *Post) TableName() string {
	return "posts"
}

func (p *Post) QueryByID() string {
	return "SELECT id, user_id, title, content, tags, metadata, created_at, updated_at FROM posts WHERE id = :1"
}

func (p *Post) QueryByUserID() string {
	return "SELECT id, user_id, title, content, tags, metadata, created_at, updated_at FROM posts WHERE user_id = :1"
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
	return "SELECT user_id, post_id, favorited_at FROM user_posts WHERE post_id = :1 AND user_id = :2"
}

func init() {
	typedb.RegisterModel[*UserPost]()
}

// TypeExample demonstrates comprehensive Oracle data types
type TypeExample struct {
	typedb.Model
	ID                 int    `db:"id" load:"primary"`
	NumberCol          string `db:"number_col"`
	NumberPrecisionCol string `db:"number_precision_col"`
	FloatCol           string `db:"float_col"`
	BinaryFloatCol     string `db:"binary_float_col"`
	BinaryDoubleCol    string `db:"binary_double_col"`
	CharCol            string `db:"char_col"`
	Varchar2Col        string `db:"varchar2_col"`
	VarcharCol         string `db:"varchar_col"`
	NcharCol           string `db:"nchar_col"`
	Nvarchar2Col       string `db:"nvarchar2_col"`
	ClobCol            string `db:"clob_col"`
	NclobCol           string `db:"nclob_col"`
	LongCol            string `db:"long_col"`
	RawCol             string `db:"raw_col"` // Binary as hex string
	BlobCol            string `db:"blob_col"`
	// Note: LONG RAW is tested in a separate table (long_raw_examples) due to Oracle's limitation
	// that only one LONG column is allowed per table
	BfileCol              string `db:"bfile_col"`
	NumberScaleCol        string `db:"number_scale_col"`
	FloatPrecisionCol     string `db:"float_precision_col"`
	TimestampPrecisionCol string `db:"timestamp_precision_col"`
	RowidCol              string `db:"rowid_col"`
	DateCol               string `db:"date_col"`
	TimestampCol          string `db:"timestamp_col"`
	TimestampTzCol        string `db:"timestamp_tz_col"`
	TimestampLtzCol       string `db:"timestamp_ltz_col"`
	IntervalYearCol       string `db:"interval_year_col"`
	IntervalDayCol        string `db:"interval_day_col"`
	XmltypeCol            string `db:"xmltype_col"`
	CreatedAt             string `db:"created_at"`
}

func (t *TypeExample) QueryByID() string {
	return "SELECT id, number_col, number_precision_col, number_scale_col, float_col, float_precision_col, binary_float_col, binary_double_col, char_col, varchar2_col, varchar_col, nchar_col, nvarchar2_col, clob_col, nclob_col, long_col, raw_col, blob_col, bfile_col, date_col, timestamp_col, timestamp_precision_col, timestamp_tz_col, timestamp_ltz_col, interval_year_col, interval_day_col, rowid_col, urowid_col, xmltype_col, created_at FROM type_examples WHERE id = :1"
}

func init() {
	typedb.RegisterModel[*TypeExample]()
}

// LongRawExample demonstrates LONG RAW type in a separate table
// (Oracle allows only one LONG column per table)
type LongRawExample struct {
	typedb.Model
	ID         int    `db:"id" load:"primary"`
	LongRawCol string `db:"long_raw_col"`
	CreatedAt  string `db:"created_at"`
	UpdatedAt  string `db:"updated_at" dbUpdate:"auto-timestamp"`
}

func (l *LongRawExample) TableName() string {
	return "long_raw_examples"
}

func (l *LongRawExample) QueryByID() string {
	return "SELECT id, long_raw_col, created_at FROM long_raw_examples WHERE id = :1"
}

func init() {
	typedb.RegisterModel[*LongRawExample]()
}
