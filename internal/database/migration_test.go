package database

import (
	"embed"
	"testing"
)

var (
	//go:embed testdata/migration1/**
	testMigration1 embed.FS
	//go:embed testdata/migration2/**
	testMigration2 embed.FS
)

func TestMigration(t *testing.T) {
	dbc, err := New("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("creating database connector: %s", err)
	}
	defer dbc.Close()

	var (
		tm1 = NewEmbedFSMigrator(testMigration1, "testdata")
		tm2 = NewEmbedFSMigrator(testMigration2, "testdata")
	)

	if err = dbc.Migrate("test", tm1); err != nil {
		t.Errorf("migration 1 take 1: %s", err)
	}

	if err = dbc.Migrate("test", tm1); err != nil {
		t.Errorf("migration 1 take 2: %s", err)
	}

	if err = dbc.Migrate("test", tm2); err != nil {
		t.Errorf("migration 2 take 1: %s", err)
	}

	if err = dbc.Migrate("test", tm2); err != nil {
		t.Errorf("migration 2 take 2: %s", err)
	}
}
