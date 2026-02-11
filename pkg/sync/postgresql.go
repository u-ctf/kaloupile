package sync

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"

	"github.com/lib/pq"
	"github.com/yyewolf/kaloupile/pkg/config"
)

func SyncPostgreSQL(cfg *config.Config) error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}

	host := envOr("PGHOST", cfg.Postgres.LocalHost)
	if host == "" {
		host = "localhost"
	}
	port := envOr("PGPORT", "")
	if port == "" {
		if cfg.Postgres.Port > 0 {
			port = strconv.Itoa(cfg.Postgres.LocalPort)
		} else {
			port = "5432"
		}
	}
	sslmode := envOr("PGSSLMODE", "disable")

	admin := cfg.Postgres.Admin
	adminDB, err := openDB(host, port, admin.User, admin.Password, admin.Database, sslmode)
	if err != nil {
		return err
	}
	defer adminDB.Close()

	if err := adminDB.Ping(); err != nil {
		return fmt.Errorf("connect admin database: %w", err)
	}

	perDB := make(map[string]*sql.DB)
	closeAll := func() {
		for _, db := range perDB {
			_ = db.Close()
		}
	}
	defer closeAll()

	for _, user := range cfg.Postgres.Users {
		exists, err := userExists(adminDB, user.Name)
		if err != nil {
			return fmt.Errorf("check user %s: %w", user.Name, err)
		}

		if exists {
			if err := setUserPassword(adminDB, user.Name, user.Password); err != nil {
				return fmt.Errorf("update user %s: %w", user.Name, err)
			}
		} else {
			if err := createUser(adminDB, user.Name, user.Password); err != nil {
				return fmt.Errorf("create user %s: %w", user.Name, err)
			}
		}

		for _, dbName := range user.Databases {
			exists, err := databaseExists(adminDB, dbName)
			if err != nil {
				return fmt.Errorf("check database %s: %w", dbName, err)
			}

			if !exists {
				if err := createDatabase(adminDB, dbName); err != nil {
					return fmt.Errorf("create database %s: %w", dbName, err)
				}
			}

			if err := grantDatabasePrivileges(adminDB, dbName, user.Name); err != nil {
				return fmt.Errorf("grant database %s: %w", dbName, err)
			}

			targetDB, ok := perDB[dbName]
			if !ok {
				targetDB, err = openDB(host, port, admin.User, admin.Password, dbName, sslmode)
				if err != nil {
					return fmt.Errorf("open database %s: %w", dbName, err)
				}
				perDB[dbName] = targetDB
				if err := targetDB.Ping(); err != nil {
					return fmt.Errorf("connect database %s: %w", dbName, err)
				}
			}

			if err := grantSchemaPrivileges(targetDB, user.Name); err != nil {
				return fmt.Errorf("grant schema on %s: %w", dbName, err)
			}
		}
	}

	return nil
}

func envOr(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func openDB(host, port, user, password, dbName, sslmode string) (*sql.DB, error) {
	if dbName == "" {
		return nil, fmt.Errorf("database name is empty")
	}

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host,
		port,
		user,
		password,
		dbName,
		sslmode,
	)
	return sql.Open("postgres", dsn)
}

func userExists(db *sql.DB, username string) (bool, error) {
	var exists int
	if err := db.QueryRow("SELECT 1 FROM pg_roles WHERE rolname=$1", username).Scan(&exists); err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func databaseExists(db *sql.DB, dbName string) (bool, error) {
	var exists int
	if err := db.QueryRow("SELECT 1 FROM pg_database WHERE datname=$1", dbName).Scan(&exists); err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func createUser(db *sql.DB, username, password string) error {
	query := fmt.Sprintf("CREATE USER %s WITH PASSWORD %s", pq.QuoteIdentifier(username), pq.QuoteLiteral(password))
	_, err := db.Exec(query)
	return err
}

func setUserPassword(db *sql.DB, username, password string) error {
	query := fmt.Sprintf("ALTER USER %s WITH PASSWORD %s", pq.QuoteIdentifier(username), pq.QuoteLiteral(password))
	_, err := db.Exec(query)
	return err
}

func createDatabase(db *sql.DB, dbName string) error {
	query := fmt.Sprintf("CREATE DATABASE %s", pq.QuoteIdentifier(dbName))
	_, err := db.Exec(query)
	return err
}

func grantDatabasePrivileges(db *sql.DB, dbName, username string) error {
	query := fmt.Sprintf(
		"GRANT ALL PRIVILEGES ON DATABASE %s TO %s",
		pq.QuoteIdentifier(dbName),
		pq.QuoteIdentifier(username),
	)
	_, err := db.Exec(query)
	return err
}

func grantSchemaPrivileges(db *sql.DB, username string) error {
	userIdent := pq.QuoteIdentifier(username)
	statements := []string{
		fmt.Sprintf("GRANT ALL ON SCHEMA public TO %s", userIdent),
		fmt.Sprintf("ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO %s", userIdent),
		fmt.Sprintf("ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO %s", userIdent),
	}

	for _, stmt := range statements {
		if _, err := db.Exec(stmt); err != nil {
			return err
		}
	}

	return nil
}
