package typedb

import (
	"context"
	"database/sql"
	"errors"
	"testing"
)

// QueryTestUser is a test model for query tests
type QueryTestUser struct {
	Model
	ID    int    `db:"id"`
	Name  string `db:"name"`
	Email string `db:"email"`
}

func (u *QueryTestUser) Deserialize(row map[string]any) error {
	return Deserialize(row, u)
}

func init() {
	RegisterModel[*QueryTestUser]()
}

func TestQueryAll(t *testing.T) {
	ctx := context.Background()

	t.Run("success with multiple rows", func(t *testing.T) {
		mock := &MockExecutor{
			QueryAllFunc: func(ctx context.Context, query string, args ...any) ([]map[string]any, error) {
				return []map[string]any{
					{"id": int64(1), "name": "Alice", "email": "alice@example.com"},
					{"id": int64(2), "name": "Bob", "email": "bob@example.com"},
				}, nil
			},
			QueryRowMapFunc: func(ctx context.Context, query string, args ...any) (map[string]any, error) {
				return nil, ErrNotFound
			},
			ExecFunc: func(ctx context.Context, query string, args ...any) (sql.Result, error) {
				return nil, nil
			},
			GetIntoFunc: func(ctx context.Context, query string, args []any, dest ...any) error {
				return nil
			},
			QueryDoFunc: func(ctx context.Context, query string, args []any, scan func(rows *sql.Rows) error) error {
				return nil
			},
		}

		users, err := QueryAll[*QueryTestUser](ctx, mock, "SELECT * FROM users")
		if err != nil {
			t.Fatalf("QueryAll failed: %v", err)
		}

		if len(users) != 2 {
			t.Fatalf("Expected 2 users, got %d", len(users))
		}

		if users[0].ID != 1 || users[0].Name != "Alice" || users[0].Email != "alice@example.com" {
			t.Errorf("First user incorrect: %+v", users[0])
		}

		if users[1].ID != 2 || users[1].Name != "Bob" || users[1].Email != "bob@example.com" {
			t.Errorf("Second user incorrect: %+v", users[1])
		}
	})

	t.Run("success with empty result", func(t *testing.T) {
		mock := &MockExecutor{
			QueryAllFunc: func(ctx context.Context, query string, args ...any) ([]map[string]any, error) {
				return []map[string]any{}, nil
			},
			QueryRowMapFunc: func(ctx context.Context, query string, args ...any) (map[string]any, error) {
				return nil, ErrNotFound
			},
			ExecFunc: func(ctx context.Context, query string, args ...any) (sql.Result, error) {
				return nil, nil
			},
			GetIntoFunc: func(ctx context.Context, query string, args []any, dest ...any) error {
				return nil
			},
			QueryDoFunc: func(ctx context.Context, query string, args []any, scan func(rows *sql.Rows) error) error {
				return nil
			},
		}

		users, err := QueryAll[*TestUser](ctx, mock, "SELECT * FROM users WHERE id = 999")
		if err != nil {
			t.Fatalf("QueryAll failed: %v", err)
		}

		if users == nil {
			t.Fatal("Expected empty slice, got nil")
		}

		if len(users) != 0 {
			t.Fatalf("Expected 0 users, got %d", len(users))
		}
	})

	t.Run("error from executor", func(t *testing.T) {
		expectedErr := errors.New("database error")
		mock := &MockExecutor{
			QueryAllFunc: func(ctx context.Context, query string, args ...any) ([]map[string]any, error) {
				return nil, expectedErr
			},
			QueryRowMapFunc: func(ctx context.Context, query string, args ...any) (map[string]any, error) {
				return nil, ErrNotFound
			},
			ExecFunc: func(ctx context.Context, query string, args ...any) (sql.Result, error) {
				return nil, nil
			},
			GetIntoFunc: func(ctx context.Context, query string, args []any, dest ...any) error {
				return nil
			},
			QueryDoFunc: func(ctx context.Context, query string, args []any, scan func(rows *sql.Rows) error) error {
				return nil
			},
		}

		users, err := QueryAll[*QueryTestUser](ctx, mock, "SELECT * FROM users")
		if err != expectedErr {
			t.Fatalf("Expected error %v, got %v", expectedErr, err)
		}

		if users != nil {
			t.Fatal("Expected nil users on error")
		}
	})

	t.Run("deserialization error", func(t *testing.T) {
		mock := &MockExecutor{
			QueryAllFunc: func(ctx context.Context, query string, args ...any) ([]map[string]any, error) {
				// Return invalid data that can't be deserialized
				return []map[string]any{
					{"id": "not-an-int", "name": "Alice", "email": "alice@example.com"},
				}, nil
			},
			QueryRowMapFunc: func(ctx context.Context, query string, args ...any) (map[string]any, error) {
				return nil, ErrNotFound
			},
			ExecFunc: func(ctx context.Context, query string, args ...any) (sql.Result, error) {
				return nil, nil
			},
			GetIntoFunc: func(ctx context.Context, query string, args []any, dest ...any) error {
				return nil
			},
			QueryDoFunc: func(ctx context.Context, query string, args []any, scan func(rows *sql.Rows) error) error {
				return nil
			},
		}

		users, err := QueryAll[*QueryTestUser](ctx, mock, "SELECT * FROM users")
		if err == nil {
			t.Fatal("Expected deserialization error")
		}

		if users != nil {
			t.Fatal("Expected nil users on deserialization error")
		}
	})
}

func TestQueryFirst(t *testing.T) {
	ctx := context.Background()

	t.Run("success with one row", func(t *testing.T) {
		mock := &MockExecutor{
			QueryRowMapFunc: func(ctx context.Context, query string, args ...any) (map[string]any, error) {
				return map[string]any{"id": int64(1), "name": "Alice", "email": "alice@example.com"}, nil
			},
		}

		user, err := QueryFirst[*QueryTestUser](ctx, mock, "SELECT * FROM users WHERE id = $1", 1)
		if err != nil {
			t.Fatalf("QueryFirst failed: %v", err)
		}

		if user == nil {
			t.Fatal("Expected user, got nil")
		}

		if user.ID != 1 || user.Name != "Alice" || user.Email != "alice@example.com" {
			t.Errorf("User incorrect: %+v", user)
		}
	})

	t.Run("no rows found - returns nil", func(t *testing.T) {
		mock := &MockExecutor{
			QueryAllFunc: func(ctx context.Context, query string, args ...any) ([]map[string]any, error) {
				return nil, nil
			},
			QueryRowMapFunc: func(ctx context.Context, query string, args ...any) (map[string]any, error) {
				return nil, ErrNotFound
			},
			ExecFunc: func(ctx context.Context, query string, args ...any) (sql.Result, error) {
				return nil, nil
			},
			GetIntoFunc: func(ctx context.Context, query string, args []any, dest ...any) error {
				return nil
			},
			QueryDoFunc: func(ctx context.Context, query string, args []any, scan func(rows *sql.Rows) error) error {
				return nil
			},
		}

		user, err := QueryFirst[*TestUser](ctx, mock, "SELECT * FROM users WHERE id = $1", 999)
		if err != nil {
			t.Fatalf("QueryFirst should not return error for no rows: %v", err)
		}

		if user != nil {
			t.Fatalf("Expected nil user, got %+v", user)
		}
	})

	t.Run("error from executor", func(t *testing.T) {
		expectedErr := errors.New("database error")
		mock := &MockExecutor{
			QueryRowMapFunc: func(ctx context.Context, query string, args ...any) (map[string]any, error) {
				return nil, expectedErr
			},
		}

		user, err := QueryFirst[*TestUser](ctx, mock, "SELECT * FROM users")
		if err != expectedErr {
			t.Fatalf("Expected error %v, got %v", expectedErr, err)
		}

		if user != nil {
			t.Fatal("Expected nil user on error")
		}
	})

	t.Run("deserialization error", func(t *testing.T) {
		mock := &MockExecutor{
			QueryRowMapFunc: func(ctx context.Context, query string, args ...any) (map[string]any, error) {
				// Return invalid data
				return map[string]any{"id": "not-an-int", "name": "Alice", "email": "alice@example.com"}, nil
			},
		}

		user, err := QueryFirst[*QueryTestUser](ctx, mock, "SELECT * FROM users WHERE id = $1", 1)
		if err == nil {
			t.Fatal("Expected deserialization error")
		}

		if user != nil {
			t.Fatal("Expected nil user on deserialization error")
		}
	})
}

func TestQueryOne(t *testing.T) {
	ctx := context.Background()

	t.Run("success with one row", func(t *testing.T) {
		mock := &MockExecutor{
			QueryAllFunc: func(ctx context.Context, query string, args ...any) ([]map[string]any, error) {
				return nil, nil
			},
			QueryRowMapFunc: func(ctx context.Context, query string, args ...any) (map[string]any, error) {
				return map[string]any{"id": int64(1), "name": "Alice", "email": "alice@example.com"}, nil
			},
			ExecFunc: func(ctx context.Context, query string, args ...any) (sql.Result, error) {
				return nil, nil
			},
			GetIntoFunc: func(ctx context.Context, query string, args []any, dest ...any) error {
				return nil
			},
			QueryDoFunc: func(ctx context.Context, query string, args []any, scan func(rows *sql.Rows) error) error {
				return nil
			},
		}

		user, err := QueryOne[*QueryTestUser](ctx, mock, "SELECT * FROM users WHERE id = $1", 1)
		if err != nil {
			t.Fatalf("QueryOne failed: %v", err)
		}

		if user == nil {
			t.Fatal("Expected user, got nil")
		}

		if user.ID != 1 || user.Name != "Alice" || user.Email != "alice@example.com" {
			t.Errorf("User incorrect: %+v", user)
		}
	})

	t.Run("no rows found - returns ErrNotFound", func(t *testing.T) {
		mock := &MockExecutor{
			QueryAllFunc: func(ctx context.Context, query string, args ...any) ([]map[string]any, error) {
				return nil, nil
			},
			QueryRowMapFunc: func(ctx context.Context, query string, args ...any) (map[string]any, error) {
				return nil, ErrNotFound
			},
			ExecFunc: func(ctx context.Context, query string, args ...any) (sql.Result, error) {
				return nil, nil
			},
			GetIntoFunc: func(ctx context.Context, query string, args []any, dest ...any) error {
				return nil
			},
			QueryDoFunc: func(ctx context.Context, query string, args []any, scan func(rows *sql.Rows) error) error {
				return nil
			},
		}

		user, err := QueryOne[*TestUser](ctx, mock, "SELECT * FROM users WHERE id = $1", 999)
		if err != ErrNotFound {
			t.Fatalf("Expected ErrNotFound, got %v", err)
		}

		if user != nil {
			t.Fatalf("Expected nil user, got %+v", user)
		}
	})

	t.Run("error from executor", func(t *testing.T) {
		expectedErr := errors.New("database error")
		mock := &MockExecutor{
			QueryAllFunc: func(ctx context.Context, query string, args ...any) ([]map[string]any, error) {
				return nil, nil
			},
			QueryRowMapFunc: func(ctx context.Context, query string, args ...any) (map[string]any, error) {
				return nil, expectedErr
			},
			ExecFunc: func(ctx context.Context, query string, args ...any) (sql.Result, error) {
				return nil, nil
			},
			GetIntoFunc: func(ctx context.Context, query string, args []any, dest ...any) error {
				return nil
			},
			QueryDoFunc: func(ctx context.Context, query string, args []any, scan func(rows *sql.Rows) error) error {
				return nil
			},
		}

		user, err := QueryOne[*TestUser](ctx, mock, "SELECT * FROM users")
		if err != expectedErr {
			t.Fatalf("Expected error %v, got %v", expectedErr, err)
		}

		if user != nil {
			t.Fatal("Expected nil user on error")
		}
	})

	t.Run("deserialization error", func(t *testing.T) {
		mock := &MockExecutor{
			QueryAllFunc: func(ctx context.Context, query string, args ...any) ([]map[string]any, error) {
				return nil, nil
			},
			QueryRowMapFunc: func(ctx context.Context, query string, args ...any) (map[string]any, error) {
				// Return invalid data
				return map[string]any{"id": "not-an-int", "name": "Alice", "email": "alice@example.com"}, nil
			},
			ExecFunc: func(ctx context.Context, query string, args ...any) (sql.Result, error) {
				return nil, nil
			},
			GetIntoFunc: func(ctx context.Context, query string, args []any, dest ...any) error {
				return nil
			},
			QueryDoFunc: func(ctx context.Context, query string, args []any, scan func(rows *sql.Rows) error) error {
				return nil
			},
		}

		user, err := QueryOne[*QueryTestUser](ctx, mock, "SELECT * FROM users WHERE id = $1", 1)
		if err == nil {
			t.Fatal("Expected deserialization error")
		}

		if user != nil {
			t.Fatal("Expected nil user on deserialization error")
		}
	})
}
