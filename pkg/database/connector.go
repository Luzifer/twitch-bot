package database

import (
	"database/sql"
	"net/url"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type (
	connector struct {
		db               *gorm.DB
		encryptionSecret string
	}
)

// ErrCoreMetaNotFound is the error thrown when reading a non-existent
// core_kv key
var ErrCoreMetaNotFound = errors.New("core meta entry not found")

// New creates a new Connector with the given driver and database
func New(driverName, connString, encryptionSecret string) (Connector, error) {
	var (
		dbTuner func(*sql.DB, error) error
		innerDB gorm.Dialector
	)

	switch driverName {
	case "mysql":
		innerDB = mysql.Open(connString)

	case "postgres":
		innerDB = postgres.Open(connString)

	case "sqlite":
		var err error
		if connString, err = patchSQLiteConnString(connString); err != nil {
			return nil, errors.Wrap(err, "patching connection string")
		}
		innerDB = sqlite.Open(connString)
		dbTuner = tuneSQLiteDatabase

	default:
		return nil, errors.Errorf("unknown database driver %s", driverName)
	}

	db, err := gorm.Open(innerDB, &gorm.Config{
		Logger: gormLogger(),
	})
	if err != nil {
		return nil, errors.Wrap(err, "connecting database")
	}

	if dbTuner != nil {
		if err = dbTuner(db.DB()); err != nil {
			return nil, errors.Wrap(err, "tuning database")
		}
	}

	conn := &connector{
		db:               db,
		encryptionSecret: encryptionSecret,
	}
	return conn, errors.Wrap(conn.applyCoreSchema(), "applying core schema")
}

func (c connector) Close() error {
	// return errors.Wrap(c.db.Close(), "closing database")
	return nil
}

func (c connector) DB() *gorm.DB {
	return c.db
}

func (c connector) applyCoreSchema() error {
	return errors.Wrap(c.db.AutoMigrate(&coreKV{}), "applying coreKV schema")
}

func gormLogger() logger.Interface {
	return logger.New(
		newLogrusLogWriterWithLevel(logrus.TraceLevel),
		logger.Config{},
	)
}

func patchSQLiteConnString(connString string) (string, error) {
	u, err := url.Parse(connString)
	if err != nil {
		return connString, errors.Wrap(err, "parsing connString")
	}

	q := u.Query()

	q.Add("_pragma", "locking_mode(EXCLUSIVE)")
	q.Add("_pragma", "synchronous(FULL)")

	u.RawQuery = strings.NewReplacer(
		"%28", "(",
		"%29", ")",
	).Replace(q.Encode())

	return u.String(), nil
}

func tuneSQLiteDatabase(db *sql.DB, err error) error {
	if err != nil {
		return errors.Wrap(err, "getting database")
	}

	db.SetConnMaxIdleTime(0)
	db.SetConnMaxLifetime(0)
	db.SetMaxIdleConns(1)
	db.SetMaxOpenConns(1)

	return nil
}
