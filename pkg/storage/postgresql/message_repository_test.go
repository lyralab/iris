package postgresql

import (
	"testing"
	"time"
)

func TestListMessagesBySender(t *testing.T) {
	// Everything runs in a rolled-back transaction on a throwaway schema, so real tables are untouched.
	tx := openDB(t).Begin()
	t.Cleanup(func() { tx.Rollback() })

	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	for _, q := range []string{
		`CREATE SCHEMA lmbs_test`,
		`SET LOCAL search_path TO lmbs_test`,
		`CREATE TABLE users (id TEXT PRIMARY KEY, user_name TEXT NOT NULL, deleted_at TIMESTAMP)`,
		`CREATE TABLE message (id TEXT PRIMARY KEY, user_id TEXT NOT NULL, receptor TEXT NOT NULL,
			message TEXT NOT NULL, sender TEXT NOT NULL, status TEXT NOT NULL, response TEXT, created_at TIMESTAMPTZ)`,
		`INSERT INTO users VALUES ('u-alice', 'alice', NULL), ('u-bob', 'bob', now())`,
	} {
		if err := tx.Exec(q).Error; err != nil {
			t.Fatalf("setup %q: %v", q, err)
		}
	}
	rows := []struct{ user, sender, receptor string }{
		{"u-alice", "kavenegar", "0911"},
		{"u-bob", "kavenegar", "0912"}, // soft-deleted user
		{"", "kavenegar", "0913"},      // receptor without a user
		{"u-alice", "telegram", "123"}, // other provider
	}
	for i, r := range rows {
		if err := tx.Exec(`INSERT INTO message VALUES (?, ?, ?, 'body', ?, 'Sent', NULL, ?)`,
			r.receptor, r.user, r.receptor, r.sender, base.Add(time.Duration(i)*time.Minute)).Error; err != nil {
			t.Fatalf("seed: %v", err)
		}
	}

	s := newStorage(t, tx)
	got, err := s.ListMessagesBySender("kavenegar", 10)
	if err != nil {
		t.Fatal(err)
	}
	// newest first; bob resolves although soft-deleted; no user -> empty name
	want := []struct{ receptor, user string }{{"0913", ""}, {"0912", "bob"}, {"0911", "alice"}}
	if len(got) != len(want) {
		t.Fatalf("got %d rows, want %d: %+v", len(got), len(want), got)
	}
	for i, w := range want {
		if got[i].Receptor != w.receptor || got[i].UserName != w.user {
			t.Fatalf("row %d = %q/%q, want %q/%q", i, got[i].Receptor, got[i].UserName, w.receptor, w.user)
		}
	}

	got, err = s.ListMessagesBySender("kavenegar", 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].Receptor != "0913" {
		t.Fatalf("limit 2: %+v", got)
	}
}
