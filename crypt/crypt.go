package crypt

import (
	"reflect"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/Luzifer/go-openssl/v4"
)

const encryptedValuePrefix = "enc:"

type encryptAction uint8

const (
	handleTagsDecrypt encryptAction = iota
	handleTagsEncrypt
)

var osslClient = openssl.New()

// DecryptFields iterates through the given struct and decrypts all
// fields marked with a struct tag of `encrypt:"true"`. The fields
// are directly manipulated and the value is replaced.
//
// The input object needs to be a pointer to a struct!
func DecryptFields(obj interface{}, passphrase string) error {
	return handleEncryptedTags(obj, passphrase, handleTagsDecrypt)
}

// EncryptFields iterates through the given struct and encrypts all
// fields marked with a struct tag of `encrypt:"true"`. The fields
// are directly manipulated and the value is replaced.
//
// The input object needs to be a pointer to a struct!
func EncryptFields(obj interface{}, passphrase string) error {
	return handleEncryptedTags(obj, passphrase, handleTagsEncrypt)
}

//nolint:gocognit,gocyclo // Reflect loop, cannot reduce complexity
func handleEncryptedTags(obj interface{}, passphrase string, action encryptAction) error {
	// Check we got a pointer and can manipulate the struct
	if kind := reflect.TypeOf(obj).Kind(); kind != reflect.Ptr {
		return errors.Errorf("expected pointer to struct, got %s", kind)
	}

	// Check we got a struct in the pointer
	if kind := reflect.ValueOf(obj).Elem().Kind(); kind != reflect.Struct {
		return errors.Errorf("expected pointer to struct, got pointer to %s", kind)
	}

	// Iterate over fields to find encrypted fields to manipulate
	st := reflect.ValueOf(obj).Elem()
	for i := 0; i < st.NumField(); i++ {
		v := st.Field(i)
		t := st.Type().Field(i)

		if t.PkgPath != "" && !t.Anonymous {
			// Caught us an non-exported field, ignore that one
			continue
		}

		hasEncryption := t.Tag.Get("encrypt") == "true"

		switch t.Type.Kind() {
		// Type: Pointer - Recurse if not nil and struct inside
		case reflect.Ptr:
			if !v.IsNil() && v.Elem().Kind() == reflect.Struct && t.Type != reflect.TypeOf(&time.Time{}) {
				if err := handleEncryptedTags(v.Interface(), passphrase, action); err != nil {
					return err
				}
			}

		// Type: String - Replace value if required
		case reflect.String:
			if hasEncryption {
				newValue, err := manipulateValue(v.String(), passphrase, action)
				if err != nil {
					return errors.Wrap(err, "manipulating value")
				}
				v.SetString(newValue)
			}

		// Type: Struct - Welcome to recursion
		case reflect.Struct:
			if t.Type != reflect.TypeOf(time.Time{}) {
				if err := handleEncryptedTags(v.Addr().Interface(), passphrase, action); err != nil {
					return err
				}
			}

		// We don't support anything else. Yet.
		default:
			if hasEncryption {
				return errors.Errorf("unsupported field type for encyption: %s", t.Type.Kind())
			}
		}
	}

	return nil
}

func manipulateValue(val, passphrase string, action encryptAction) (string, error) {
	if action == handleTagsDecrypt && !strings.HasPrefix(val, encryptedValuePrefix) {
		// This is not an encrypted string: Return the value itself for
		// working with legacy values in storage
		return val, nil
	}

	if action == handleTagsEncrypt && strings.HasPrefix(val, encryptedValuePrefix) {
		// This is an encrypted string: shouldn't happen but whatever
		return val, nil
	}

	switch action {
	case handleTagsDecrypt:
		d, err := osslClient.DecryptBytes(passphrase, []byte(strings.TrimPrefix(val, encryptedValuePrefix)), openssl.PBKDF2SHA256)
		return string(d), errors.Wrap(err, "decrypting value")

	case handleTagsEncrypt:
		e, err := osslClient.EncryptBytes(passphrase, []byte(val), openssl.PBKDF2SHA256)
		return encryptedValuePrefix + string(e), errors.Wrap(err, "encrypting value")

	default:
		return "", errors.New("invalid action")
	}
}
