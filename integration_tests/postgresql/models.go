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
	return "SELECT id, name, email, created_at, updated_at FROM users WHERE id = $1"
}

func (u *User) QueryByEmail() string {
	return "SELECT id, name, email, created_at, updated_at FROM users WHERE email = $1"
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
	Tags      string `db:"tags"`     // PostgreSQL array as string
	Metadata  string `db:"metadata"` // JSONB as string
	CreatedAt string `db:"created_at"`
	UpdatedAt string `db:"updated_at" dbUpdate:"auto-timestamp"`
}

func (p *Post) TableName() string {
	return "posts"
}

func (p *Post) QueryByID() string {
	return "SELECT id, user_id, title, content, tags, metadata, created_at, updated_at FROM posts WHERE id = $1"
}

func (p *Post) QueryByUserID() string {
	return "SELECT id, user_id, title, content, tags, metadata, created_at, updated_at FROM posts WHERE user_id = $1"
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
	return "SELECT user_id, post_id, favorited_at FROM user_posts WHERE post_id = $1 AND user_id = $2"
}

func init() {
	typedb.RegisterModel[*UserPost]()
}

// TypeExample demonstrates comprehensive PostgreSQL data types
type TypeExample struct {
	typedb.Model
	ID                   int    `db:"id" load:"primary"`
	SmallInt             int    `db:"small_int"`
	IntegerCol           int    `db:"integer_col"`
	BigInt               int64  `db:"big_int"`
	DecimalCol           string `db:"decimal_col"`
	NumericCol           string `db:"numeric_col"`
	RealCol              string `db:"real_col"`
	DoublePrecisionCol   string `db:"double_precision_col"`
	MoneyCol             string `db:"money_col"`
	VarcharCol           string `db:"varchar_col"`
	CharCol              string `db:"char_col"`
	TextCol              string `db:"text_col"`
	ByteaCol             string `db:"bytea_col"` // Binary as hex string
	DateCol              string `db:"date_col"`
	TimeCol              string `db:"time_col"`
	TimeTzCol            string `db:"time_tz_col"`
	TimestampCol         string `db:"timestamp_col"`
	TimestamptzCol       string `db:"timestamptz_col"`
	IntervalCol          string `db:"interval_col"`
	BooleanCol           bool   `db:"boolean_col"`
	JsonCol              string `db:"json_col"`
	JsonbCol             string `db:"jsonb_col"`
	SmallintArray        string `db:"smallint_array"`
	IntArray             string `db:"int_array"`
	BigintArray          string `db:"bigint_array"`
	RealArray            string `db:"real_array"`
	DoublePrecisionArray string `db:"double_precision_array"`
	NumericArray         string `db:"numeric_array"`
	VarcharArray         string `db:"varchar_array"`
	TextArray            string `db:"text_array"`
	BooleanArray         string `db:"boolean_array"`
	DateArray            string `db:"date_array"`
	TimestampArray       string `db:"timestamp_array"`
	JsonArray            string `db:"json_array"`
	JsonbArray           string `db:"jsonb_array"`
	UuidArray            string `db:"uuid_array"`
	ByteaArray           string `db:"bytea_array"`
	UuidCol              string `db:"uuid_col"`
	InetCol              string `db:"inet_col"`
	CidrCol              string `db:"cidr_col"`
	MacaddrCol           string `db:"macaddr_col"`
	Macaddr8Col          string `db:"macaddr8_col"`
	PointCol             string `db:"point_col"`
	LineCol              string `db:"line_col"`
	LsegCol              string `db:"lseg_col"`
	BoxCol               string `db:"box_col"`
	PathCol              string `db:"path_col"`
	PolygonCol           string `db:"polygon_col"`
	CircleCol            string `db:"circle_col"`
	Int4rangeCol         string `db:"int4range_col"`
	Int8rangeCol         string `db:"int8range_col"`
	NumrangeCol          string `db:"numrange_col"`
	TsrangeCol           string `db:"tsrange_col"`
	TstzrangeCol         string `db:"tstzrange_col"`
	DaterangeCol         string `db:"daterange_col"`
	BitCol               string `db:"bit_col"`
	VarbitCol            string `db:"varbit_col"`
	TsvectorCol          string `db:"tsvector_col"`
	TsqueryCol           string `db:"tsquery_col"`
	XmlCol               string `db:"xml_col"`
	CreatedAt            string `db:"created_at"`
}

func (t *TypeExample) QueryByID() string {
	return "SELECT id, small_int, integer_col, big_int, decimal_col, numeric_col, real_col, double_precision_col, money_col, varchar_col, char_col, text_col, bytea_col, date_col, time_col, time_tz_col, timestamp_col, timestamptz_col, interval_col, boolean_col, json_col, jsonb_col, smallint_array, int_array, bigint_array, real_array, double_precision_array, numeric_array, varchar_array, text_array, boolean_array, date_array, timestamp_array, json_array, jsonb_array, uuid_array, bytea_array, uuid_col, inet_col, cidr_col, macaddr_col, macaddr8_col, point_col, line_col, lseg_col, box_col, path_col, polygon_col, circle_col, int4range_col, int8range_col, numrange_col, tsrange_col, tstzrange_col, daterange_col, bit_col, varbit_col, tsvector_col, tsquery_col, xml_col, created_at FROM type_examples WHERE id = $1"
}

func init() {
	typedb.RegisterModel[*TypeExample]()
}
