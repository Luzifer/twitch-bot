package authcache

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

//nolint:funlen // just a list of testcases
func TestValidateTokenForWithAuthBackend(t *testing.T) {
	t.Parallel()

	t.Run("twitch bypasses internal", func(t *testing.T) {
		t.Parallel()

		var calls [2]int
		svc := New(
			WithAuthBackend("internal", func(_ string) ([]string, time.Time, error) {
				calls[0]++
				return nil, time.Time{}, ErrUnauthorized
			}),
			WithAuthBackend("twitch", func(_ string) ([]string, time.Time, error) {
				calls[1]++
				return []string{"*"}, time.Now().Add(time.Minute), nil
			}),
		)

		err := svc.ValidateTokenForWithTokenType("token", "twitch", "module")
		require.NoError(t, err)
		require.Equal(t, [2]int{0, 1}, calls)
	})

	t.Run("internal bypasses twitch", func(t *testing.T) {
		t.Parallel()

		var calls [2]int
		svc := New(
			WithAuthBackend("internal", func(_ string) ([]string, time.Time, error) {
				calls[0]++
				return []string{"module"}, time.Now().Add(time.Minute), nil
			}),
			WithAuthBackend("twitch", func(_ string) ([]string, time.Time, error) {
				calls[1]++
				return []string{"*"}, time.Now().Add(time.Minute), nil
			}),
		)

		err := svc.ValidateTokenForWithTokenType("token", "internal", "module")
		require.NoError(t, err)
		require.Equal(t, [2]int{1, 0}, calls)
	})

	t.Run("legacy token uses both backends", func(t *testing.T) {
		t.Parallel()

		var calls [2]int
		svc := New(
			WithAuthBackend("internal", func(_ string) ([]string, time.Time, error) {
				calls[0]++
				return nil, time.Time{}, ErrUnauthorized
			}),
			WithAuthBackend("twitch", func(_ string) ([]string, time.Time, error) {
				calls[1]++
				return []string{"module"}, time.Now().Add(time.Minute), nil
			}),
		)

		err := svc.ValidateTokenForWithTokenType("token", AuthBackendAny, "module")
		require.NoError(t, err)
		require.Equal(t, [2]int{1, 1}, calls)
	})

	t.Run("cache key includes auth backend", func(t *testing.T) {
		t.Parallel()

		var calls [2]int
		svc := New(
			WithAuthBackend("internal", func(_ string) ([]string, time.Time, error) {
				calls[0]++
				return nil, time.Time{}, ErrUnauthorized
			}),
			WithAuthBackend("twitch", func(_ string) ([]string, time.Time, error) {
				calls[1]++
				return []string{"module"}, time.Now().Add(time.Minute), nil
			}),
		)

		require.NoError(t, svc.ValidateTokenForWithTokenType("token", AuthBackendAny, "module"))
		require.NoError(t, svc.ValidateTokenForWithTokenType("token", "twitch", "module"))
		require.Equal(t, [2]int{1, 2}, calls)
	})

	t.Run("non unauthorized errors are returned", func(t *testing.T) {
		t.Parallel()

		svc := New(
			WithAuthBackend("internal", func(_ string) ([]string, time.Time, error) {
				return nil, time.Time{}, errors.New("boom")
			}),
		)

		err := svc.ValidateTokenForWithTokenType("token", AuthBackendAny, "module")
		require.Error(t, err)
		require.Contains(t, err.Error(), "querying authorization in backend")
	})

	t.Run("unknown backend is unauthorized", func(t *testing.T) {
		t.Parallel()

		var calls [2]int
		svc := New(
			WithAuthBackend("internal", func(_ string) ([]string, time.Time, error) {
				calls[0]++
				return nil, time.Time{}, ErrUnauthorized
			}),
			WithAuthBackend("twitch", func(_ string) ([]string, time.Time, error) {
				calls[1]++
				return []string{"module"}, time.Now().Add(time.Minute), nil
			}),
		)

		err := svc.ValidateTokenForWithTokenType("token", "nope", "module")
		require.ErrorIs(t, err, ErrUnauthorized)
		require.Equal(t, [2]int{0, 0}, calls)
	})
}
