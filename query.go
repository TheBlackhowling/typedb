package typedb

import (
	"context"
)

// QueryAll executes a query and returns all rows as a slice of model pointers.
// Returns an empty slice if no rows are found.
// T must be a pointer type (e.g., *User).
//
// Example:
//
//	users, err := typedb.QueryAll[*User](ctx, db, "SELECT id, name, email FROM users")
func QueryAll[T ModelInterface](ctx context.Context, exec Executor, query string, args ...any) ([]T, error) {
	rows, err := exec.QueryAll(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	if len(rows) == 0 {
		return []T{}, nil
	}

	result := make([]T, 0, len(rows))
	for _, row := range rows {
		model, err := deserializeForType[T](row)
		if err != nil {
			return nil, err
		}
		result = append(result, model)
	}

	return result, nil
}

// QueryFirst executes a query and returns the first row as a model pointer.
// Returns nil if no rows are found (no error).
// T must be a pointer type (e.g., *User).
//
// Example:
//
//	user, err := typedb.QueryFirst[*User](ctx, db, "SELECT id, name, email FROM users WHERE id = $1", 123)
//	if user == nil {
//	    // No user found
//	}
func QueryFirst[T ModelInterface](ctx context.Context, exec Executor, query string, args ...any) (T, error) {
	row, err := exec.QueryRowMap(ctx, query, args...)
	if err != nil {
		if err == ErrNotFound {
			var zero T
			return zero, nil
		}
		var zero T
		return zero, err
	}

	model, err := deserializeForType[T](row)
	if err != nil {
		var zero T
		return zero, err
	}

	return model, nil
}

// QueryOne executes a query and returns exactly one row as a model pointer.
// Returns ErrNotFound if no rows are found.
// Returns an error if multiple rows are found.
// T must be a pointer type (e.g., *User).
//
// Example:
//
//	user, err := typedb.QueryOne[*User](ctx, db, "SELECT id, name, email FROM users WHERE id = $1", 123)
//	if err == typedb.ErrNotFound {
//	    // User not found
//	}
func QueryOne[T ModelInterface](ctx context.Context, exec Executor, query string, args ...any) (T, error) {
	row, err := exec.QueryRowMap(ctx, query, args...)
	if err != nil {
		var zero T
		return zero, err
	}

	model, err := deserializeForType[T](row)
	if err != nil {
		var zero T
		return zero, err
	}

	return model, nil
}
