package main

import (
	"context"
	"testing"

	"github.com/TheBlackHowling/typedb"
	_ "github.com/go-sql-driver/mysql" // MySQL driver
)

func TestMySQL_Load(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("mysql", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer closeDB(t, db)

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

	user := &User{ID: 1}
	if err := typedb.Load(ctx, db, user); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if user.Name == "" {
		t.Error("User name should be loaded")
	}
	if user.Email == "" {
		t.Error("User email should be loaded")
	}
}

func TestMySQL_LoadByField(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("mysql", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer closeDB(t, db)

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

	user := &User{Email: "alice@example.com"}
	if err := typedb.LoadByField(ctx, db, user, "Email"); err != nil {
		t.Fatalf("LoadByField failed: %v", err)
	}

	if user.ID == 0 {
		t.Error("User ID should be loaded")
	}
	if user.Name == "" {
		t.Error("User name should be loaded")
	}
}

func TestMySQL_LoadByComposite(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("mysql", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer closeDB(t, db)

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

	userPost := &UserPost{UserID: 1, PostID: 1}
	if err := typedb.LoadByComposite(ctx, db, userPost, "userpost"); err != nil {
		t.Fatalf("LoadByComposite failed: %v", err)
	}

	if userPost.FavoritedAt == "" {
		t.Error("FavoritedAt should be loaded")
	}
}

func TestMySQL_Transaction(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("mysql", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer closeDB(t, db)

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

	err = db.WithTx(ctx, func(tx *typedb.Tx) error {
		users, err := typedb.QueryAll[*User](ctx, tx, "SELECT id, name, email, created_at FROM users WHERE id = ?", 1)
		if err != nil {
			return err
		}
		if len(users) == 0 {
			t.Error("Expected at least one user in transaction")
		}
		return nil
	}, nil)

	if err != nil {
		t.Fatalf("Transaction failed: %v", err)
	}
}
