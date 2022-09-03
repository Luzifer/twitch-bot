package database

import (
	"path"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

func (c connector) Migrate(module string, migrations MigrationStorage) error {
	m, err := collectMigrations(migrations, "/")
	if err != nil {
		return errors.Wrap(err, "collecting migrations")
	}

	migrationKey := strings.Join([]string{"migration_state", module}, "-")

	var lastMigration int
	if err = c.ReadCoreMeta(migrationKey, &lastMigration); err != nil && !errors.Is(err, ErrCoreMetaNotFound) {
		return errors.Wrap(err, "getting last migration")
	}

	nextMigration := lastMigration
	for {
		nextMigration++
		filename := m[nextMigration]
		if filename == "" {
			break
		}

		if err = c.applyMigration(migrations, filename); err != nil {
			return errors.Wrapf(err, "applying migration %d", nextMigration)
		}

		if err = c.StoreCoreMeta(migrationKey, nextMigration); err != nil {
			return errors.Wrap(err, "updating migration number")
		}
	}

	return nil
}

func (c connector) applyMigration(migrations MigrationStorage, filename string) error {
	rawMigration, err := migrations.ReadFile(filename)
	if err != nil {
		return errors.Wrap(err, "reading migration file")
	}

	_, err = c.db.Exec(string(rawMigration))
	return errors.Wrap(err, "executing migration statement(s)")
}

func collectMigrations(migrations MigrationStorage, dir string) (map[int]string, error) {
	out := map[int]string{}

	entries, err := migrations.ReadDir(dir)
	if err != nil {
		return nil, errors.Wrapf(err, "reading dir %q", dir)
	}

	for _, e := range entries {
		if e.IsDir() {
			sout, err := collectMigrations(migrations, path.Join(dir, e.Name()))
			if err != nil {
				return nil, errors.Wrapf(err, "scanning subdir %q", e.Name())
			}

			for n, p := range sout {
				if out[n] != "" {
					return nil, errors.Errorf("migration %d found more than once", n)
				}

				out[n] = p
			}

			continue
		}

		if !migrationFilename.MatchString(e.Name()) {
			continue
		}

		matches := migrationFilename.FindStringSubmatch(e.Name())
		n, err := strconv.Atoi(matches[1])
		if err != nil {
			return nil, errors.Wrap(err, "parsing migration number")
		}

		out[n] = path.Join(dir, e.Name())
	}

	return out, nil
}
