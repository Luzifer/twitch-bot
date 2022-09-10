package database

import (
	"testing"

	"github.com/pkg/errors"
)

const testEncryptionPass = "password123"

func TestNewConnector(t *testing.T) {
	dbc, err := New("sqlite", ":memory:", testEncryptionPass)
	if err != nil {
		t.Fatalf("creating database connector: %s", err)
	}
	defer dbc.Close()

	row := dbc.DB().QueryRow("SELECT count(1) AS tables FROM sqlite_master WHERE type='table' AND name='core_kv';")

	var count int
	if err = row.Scan(&count); err != nil {
		t.Fatalf("reading table count result")
	}

	if count != 1 {
		t.Errorf("expected to find one result, got %d in count of core_kv table", count)
	}
}

func TestCoreMetaRoundtrip(t *testing.T) {
	dbc, err := New("sqlite", ":memory:", testEncryptionPass)
	if err != nil {
		t.Fatalf("creating database connector: %s", err)
	}
	defer dbc.Close()

	var (
		arbitrary struct{ A string }
		testKey   = "arbitrary"
	)

	if err = dbc.ReadCoreMeta(testKey, &arbitrary); !errors.Is(err, ErrCoreMetaNotFound) {
		t.Error("expected core_kv not to contain key after init")
	}

	checkWriteRead := func(testString string) {
		arbitrary.A = testString
		if err = dbc.StoreCoreMeta(testKey, arbitrary); err != nil {
			t.Errorf("storing core_kv: %s", err)
		}

		arbitrary.A = "" // Clear to test unmarshal
		if err = dbc.ReadCoreMeta(testKey, &arbitrary); err != nil {
			t.Errorf("reading core_kv: %s", err)
		}

		if arbitrary.A != testString {
			t.Errorf("expected meta entry to have %q, got %q", testString, arbitrary.A)
		}
	}

	checkWriteRead("just a string")         // Turn one: Init from not existing
	checkWriteRead("another random string") // Turn two: Overwrite
}

func TestCoreMetaEncryption(t *testing.T) {
	dbc, err := New("sqlite", ":memory:", testEncryptionPass)
	if err != nil {
		t.Fatalf("creating database connector: %s", err)
	}
	defer dbc.Close()

	var (
		arbitrary  struct{ A string }
		testKey    = "arbitrary"
		testString = "foobar"
	)

	arbitrary.A = testString

	if err = dbc.StoreEncryptedCoreMeta(testKey, arbitrary); err != nil {
		t.Fatalf("storing encrypted core meta: %s", err)
	}

	if err = dbc.ReadCoreMeta(testKey, &arbitrary); err == nil {
		t.Error("reading encrypted meta without decryption succeeded")
	}

	arbitrary.A = ""

	if err = dbc.ReadEncryptedCoreMeta(testKey, &arbitrary); err != nil {
		t.Errorf("reading encrypted meta: %s", err)
	}

	if arbitrary.A != testString {
		t.Errorf("unexpected value: %q != %q", arbitrary.A, testString)
	}
}
