package database

import (
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/glebarez/sqlite"
	mysqlDriver "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	mysqlMaxIdleConnections = 2 // Default as of Go 1.20
	mysqlMaxOpenConnections = 10
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
func New(driverName, connString, encryptionSecret string) (c Connector, err error) {
	var (
		dbTuner func(*sql.DB, error) error
		innerDB gorm.Dialector
	)

	switch driverName {
	case "mysql":
		if err = mysqlDriver.SetLogger(NewLogrusLogWriterWithLevel(logrus.StandardLogger(), logrus.ErrorLevel, driverName)); err != nil {
			return nil, fmt.Errorf("setting logger on mysql driver: %w", err)
		}
		innerDB = mysql.Open(connString)
		dbTuner = tuneMySQLDatabase

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
		DisableForeignKeyConstraintWhenMigrating: true,
		Logger: logger.New(NewLogrusLogWriterWithLevel(logrus.StandardLogger(), logrus.TraceLevel, driverName), logger.Config{
			SlowThreshold:             time.Second,
			Colorful:                  false,
			IgnoreRecordNotFoundError: false,
			ParameterizedQueries:      false,
			LogLevel:                  logger.Info,
		}),
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

func (connector) Close() error {
	return nil
}

func (connector) CopyDatabase(src, target *gorm.DB) error {
	return CopyObjects(src, target, &coreKV{})
}

func (c connector) DB() *gorm.DB {
	return c.db
}

func (c connector) applyCoreSchema() error {
	return errors.Wrap(c.db.AutoMigrate(&coreKV{}), "applying coreKV schema")
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

func tuneMySQLDatabase(db *sql.DB, err error) error {
	if err != nil {
		return errors.Wrap(err, "getting database")
	}

	// By default the package allows unlimited connections and the
	// default value of a MySQL / MariaDB server is to allow 151
	// connections at most. Therefore we tune the connection pool to
	// sane values in order not to flood the database with connections
	// in case a lot of events occur at the same time.

	db.SetConnMaxIdleTime(time.Hour)
	db.SetConnMaxLifetime(time.Hour)
	db.SetMaxIdleConns(mysqlMaxIdleConnections)
	db.SetMaxOpenConns(mysqlMaxOpenConnections)

	return nil
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
