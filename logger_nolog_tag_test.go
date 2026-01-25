package typedb

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

// UserWithNolog is a test model with nolog tag
type UserWithNolog struct {
	Model
	ID       int    `db:"id" load:"primary"`
	Name     string `db:"name"`
	Password string `db:"password" nolog:"true"`
	Email    string `db:"email"`
}

func (u *UserWithNolog) TableName() string {
	return "users"
}

func (u *UserWithNolog) QueryByID() string {
	return "SELECT id, name, email FROM users WHERE id = $1"
}

// TestNologTagMasking verifies that nolog struct tags mask arguments in logs
func TestNologTagMasking(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	logger := &testLogger{}
	ctx := context.Background()

	RegisterModel[*UserWithNolog]()

	user := &UserWithNolog{
		Name:     "John",
		Password: "secret123",
		Email:    "john@example.com",
	}

	typedbDB := NewDBWithLoggerAndFlags(db, "postgres", 5*time.Second, logger, true, true)

	t.Run("Insert masks nolog fields", func(t *testing.T) {
		logger.debugs = nil
		logger.errors = nil

		mock.ExpectQuery(`INSERT INTO "users" \("name", "password", "email"\) VALUES \(\$1, \$2, \$3\) RETURNING "id"`).
			WithArgs("John", "secret123", "john@example.com").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		err := Insert(ctx, typedbDB, user)
		if err != nil {
			t.Fatalf("Insert failed: %v", err)
		}

		// Check that password argument is masked in logs
		foundArgs := false
		var loggedArgs []any
		for _, entry := range logger.debugs {
			for i := 0; i < len(entry.keyvals)-1; i += 2 {
				if entry.keyvals[i] == "args" {
					foundArgs = true
					if args, ok := entry.keyvals[i+1].([]any); ok {
						loggedArgs = args
					}
					break
				}
			}
			if foundArgs {
				break
			}
		}

		if !foundArgs {
			t.Fatal("Expected 'args' key in log")
		}

		// Password should be masked (at index 1: Name, Password, Email)
		if len(loggedArgs) < 3 {
			t.Fatalf("Expected at least 3 arguments, got %d", len(loggedArgs))
		}
		if loggedArgs[1] != "[REDACTED]" {
			t.Errorf("Expected password to be masked, got %v", loggedArgs[1])
		}
		// Other fields should not be masked
		if loggedArgs[0] != "John" {
			t.Errorf("Expected name to be 'John', got %v", loggedArgs[0])
		}
		if loggedArgs[2] != "john@example.com" {
			t.Errorf("Expected email to be 'john@example.com', got %v", loggedArgs[2])
		}
	})

	t.Run("Update masks nolog fields", func(t *testing.T) {
		logger.debugs = nil
		logger.errors = nil

		user.ID = 1
		mock.ExpectExec(`UPDATE "users" SET "name" = \$1, "password" = \$2, "email" = \$3 WHERE "id" = \$4`).
			WithArgs("John", "secret123", "john@example.com", 1).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := Update(ctx, typedbDB, user)
		if err != nil {
			t.Fatalf("Update failed: %v", err)
		}

		// Check that password argument is masked in logs
		foundArgs := false
		var loggedArgs []any
		for _, entry := range logger.debugs {
			for i := 0; i < len(entry.keyvals)-1; i += 2 {
				if entry.keyvals[i] == "args" {
					foundArgs = true
					if args, ok := entry.keyvals[i+1].([]any); ok {
						loggedArgs = args
					}
					break
				}
			}
			if foundArgs {
				break
			}
		}

		if !foundArgs {
			t.Fatal("Expected 'args' key in log")
		}

		// Password should be masked (at index 1: Name, Password, Email, then ID)
		if len(loggedArgs) < 3 {
			t.Fatalf("Expected at least 3 arguments, got %d", len(loggedArgs))
		}
		if loggedArgs[1] != "[REDACTED]" {
			t.Errorf("Expected password to be masked, got %v", loggedArgs[1])
		}
	})
}

// UserWithNologPK is a test model with nolog tag on primary key
type UserWithNologPK struct {
	Model
	ID       int    `db:"id" load:"primary" nolog:"true"`
	Name     string `db:"name"`
	Email    string `db:"email"`
}

func (u *UserWithNologPK) TableName() string {
	return "users"
}

func (u *UserWithNologPK) QueryByID() string {
	return "SELECT id, name, email FROM users WHERE id = $1"
}

// UserWithNologEmail is a test model with nolog tag on email field
type UserWithNologEmail struct {
	Model
	ID       int    `db:"id" load:"primary"`
	Name     string `db:"name"`
	Email    string `db:"email" nolog:"true"`
}

func (u *UserWithNologEmail) TableName() string {
	return "users"
}

func (u *UserWithNologEmail) QueryByID() string {
	return "SELECT id, name, email FROM users WHERE id = $1"
}

func (u *UserWithNologEmail) QueryByEmail() string {
	return "SELECT id, name, email FROM users WHERE email = $1"
}

// UserPostWithNolog is a test model with composite key where one field has nolog tag
type UserPostWithNolog struct {
	Model
	UserID   int    `db:"user_id" load:"composite:userpost"`
	PostID   int    `db:"post_id" load:"composite:userpost" nolog:"true"`
	Title    string `db:"title"`
}

func (u *UserPostWithNolog) TableName() string {
	return "user_posts"
}

func (u *UserPostWithNolog) QueryByPostIDUserID() string {
	return "SELECT user_id, post_id, title FROM user_posts WHERE post_id = $1 AND user_id = $2"
}

// TestNologTagMaskingInLoad verifies that nolog struct tags mask arguments in Load operations
func TestNologTagMaskingInLoad(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	logger := &testLogger{}
	ctx := context.Background()

	RegisterModel[*UserWithNologPK]()

	typedbDB := NewDBWithLoggerAndFlags(db, "postgres", 5*time.Second, logger, true, true)

	t.Run("Load masks nolog primary key field", func(t *testing.T) {
		logger.debugs = nil
		logger.errors = nil

		user := &UserWithNologPK{ID: 123}

		mock.ExpectQuery("SELECT id, name, email FROM users WHERE id = \\$1").
			WithArgs(123).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email"}).
				AddRow(123, "John", "john@example.com"))

		err := Load(ctx, typedbDB, user)
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}

		// Check that ID argument is masked in logs
		foundArgs := false
		var loggedArgs []any
		for _, entry := range logger.debugs {
			for i := 0; i < len(entry.keyvals)-1; i += 2 {
				if entry.keyvals[i] == "args" {
					foundArgs = true
					if args, ok := entry.keyvals[i+1].([]any); ok {
						loggedArgs = args
					}
					break
				}
			}
			if foundArgs {
				break
			}
		}

		if !foundArgs {
			t.Fatal("Expected 'args' key in log")
		}

		// ID should be masked (at index 0)
		if len(loggedArgs) < 1 {
			t.Fatalf("Expected at least 1 argument, got %d", len(loggedArgs))
		}
		if loggedArgs[0] != "[REDACTED]" {
			t.Errorf("Expected ID to be masked, got %v", loggedArgs[0])
		}
	})

	t.Run("LoadByField masks nolog field", func(t *testing.T) {
		logger.debugs = nil
		logger.errors = nil

		RegisterModel[*UserWithNologEmail]()

		user := &UserWithNologEmail{Email: "secret@example.com"}

		mock.ExpectQuery("SELECT id, name, email FROM users WHERE email = \\$1").
			WithArgs("secret@example.com").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email"}).
				AddRow(1, "John", "secret@example.com"))

		err := LoadByField(ctx, typedbDB, user, "Email")
		if err != nil {
			t.Fatalf("LoadByField failed: %v", err)
		}

		// Check that Email argument is masked in logs
		foundArgs := false
		var loggedArgs []any
		for _, entry := range logger.debugs {
			for i := 0; i < len(entry.keyvals)-1; i += 2 {
				if entry.keyvals[i] == "args" {
					foundArgs = true
					if args, ok := entry.keyvals[i+1].([]any); ok {
						loggedArgs = args
					}
					break
				}
			}
			if foundArgs {
				break
			}
		}

		if !foundArgs {
			t.Fatal("Expected 'args' key in log")
		}

		// Email should be masked (at index 0)
		if len(loggedArgs) < 1 {
			t.Fatalf("Expected at least 1 argument, got %d", len(loggedArgs))
		}
		if loggedArgs[0] != "[REDACTED]" {
			t.Errorf("Expected Email to be masked, got %v", loggedArgs[0])
		}
	})

	t.Run("LoadByComposite masks nolog fields", func(t *testing.T) {
		logger.debugs = nil
		logger.errors = nil

		RegisterModel[*UserPostWithNolog]()

		userPost := &UserPostWithNolog{UserID: 1, PostID: 2}

		mock.ExpectQuery("SELECT user_id, post_id, title FROM user_posts WHERE post_id = \\$1 AND user_id = \\$2").
			WithArgs(2, 1). // PostID first (alphabetically sorted), then UserID
			WillReturnRows(sqlmock.NewRows([]string{"user_id", "post_id", "title"}).
				AddRow(1, 2, "Test Post"))

		err := LoadByComposite(ctx, typedbDB, userPost, "userpost")
		if err != nil {
			t.Fatalf("LoadByComposite failed: %v", err)
		}

		// Check that PostID argument is masked in logs (at index 0, since PostID comes before UserID alphabetically)
		foundArgs := false
		var loggedArgs []any
		for _, entry := range logger.debugs {
			for i := 0; i < len(entry.keyvals)-1; i += 2 {
				if entry.keyvals[i] == "args" {
					foundArgs = true
					if args, ok := entry.keyvals[i+1].([]any); ok {
						loggedArgs = args
					}
					break
				}
			}
			if foundArgs {
				break
			}
		}

		if !foundArgs {
			t.Fatal("Expected 'args' key in log")
		}

		// PostID should be masked (at index 0), UserID should not be masked (at index 1)
		if len(loggedArgs) < 2 {
			t.Fatalf("Expected at least 2 arguments, got %d", len(loggedArgs))
		}
		if loggedArgs[0] != "[REDACTED]" {
			t.Errorf("Expected PostID to be masked, got %v", loggedArgs[0])
		}
		if loggedArgs[1] == "[REDACTED]" {
			t.Errorf("Expected UserID to NOT be masked, got %v", loggedArgs[1])
		}
		if loggedArgs[1] != 1 {
			t.Errorf("Expected UserID to be 1, got %v", loggedArgs[1])
		}
	})
}
