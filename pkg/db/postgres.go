package db

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx"
)

// Postgres object
type Postgres struct {
	Name     string
	Username string
	Password string
}

// PostgresServer object
type PostgresServer struct {
	Host     string
	Port     int32
	Postgres Postgres
	DB       *sql.DB
	NewDB    *sql.DB
}

// CreateUser creates a user
func (ps *PostgresServer) CreateUser() (string, error) {
	// Check if user exists on server
	commandTag, err := ps.DB.Exec(fmt.Sprintf("SELECT usename FROM pg_user WHERE usename='%s'", ps.Postgres.Username))
	rows, _ := commandTag.RowsAffected()
	// If user doesn't exist create new
	if err != nil || rows == 0 {
		_, err = ps.DB.Exec(fmt.Sprintf("CREATE USER \"%s\" WITH PASSWORD '%s'", ps.Postgres.Username, ps.Postgres.Password))
		if err != nil {
			return "unable to create role in database", err
		}
		return "User created successfully", nil
		// If user exists update password
	} else {
		_, err = ps.DB.Exec(fmt.Sprintf("ALTER USER \"%s\" WITH PASSWORD '%s'", ps.Postgres.Username, ps.Postgres.Password))
		if err != nil {
			return "unable to alter user in database", err
		}
		return "User altered successfully", nil
	}
}

// DeleteUser from server
func (ps *PostgresServer) DeleteUser() (string, error) {
	_, err := ps.DB.Exec(fmt.Sprintf("DROP USER \"%s\"", ps.Postgres.Username))
	if err != nil {
		return "unable to drop user in database server", err
	}
	return "User deleted successfully", nil
}

// CreateDatabase creates a database
func (ps *PostgresServer) CreateDatabase() (string, error) {
	// Try to create database
	_, err := ps.DB.Exec(fmt.Sprintf("CREATE DATABASE \"%s\" TEMPLATE \"template0\"", ps.Postgres.Name))
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			return "Database already exisis", nil
		} else {
			return "unable to create database in database server", err
		}
	}
	return "Database created successfully", nil
}

// DeleteDatabase from server
func (ps *PostgresServer) DeleteDatabase() (string, error) {
	_, err := ps.DB.Exec(fmt.Sprintf("DROP DATABASE \"%s\"", ps.Postgres.Name))
	if err != nil {
		return "unable to drop database in database server", err
	}
	return "Database deleted successfully", nil
}

// GrantPermissions to user
func (ps *PostgresServer) GrantPermissions() (string, error) {
	// Grant permissions to user
	_, err := ps.DB.Exec(fmt.Sprintf("GRANT ALL ON DATABASE \"%s\" TO \"%s\"", ps.Postgres.Name, ps.Postgres.Username))
	if err != nil {
		return "unable to grant permissions in database", err
	}
	return "Permissions successfully granted", nil
}

// TestNewConnection with newly created role and database
func (ps *PostgresServer) TestNewConnection() (string, error) {
	// Test if connection to new database works
	newURL := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s", ps.Postgres.Username, ps.Postgres.Password, ps.Host, ps.Port, ps.Postgres.Name)

	newdb, err := sql.Open("postgres", newURL)
	if err != nil {
		return "Unable to connect to new database", err
	}
	ps.NewDB = newdb
	return "Connection to new database successfull", nil
}

// Migrate database to newest version
func (ps *PostgresServer) Migrate(url string) (string, error) {
	driver, err := postgres.WithInstance(ps.NewDB, &postgres.Config{})
	if err != nil {
		return "error getting driver", err
	}
	m, err := migrate.NewWithDatabaseInstance(
		url,
		"postgres", driver)
	if err != nil {
		return "migration failed", err
	}
	m.Up()
	return "Migration successfull", nil
}

// Connect to postgresserver
func (ps *PostgresServer) Connect(username, password string) (string, error) {
	url := fmt.Sprintf("postgresql://%s:%s@%s:%d/postgres", username, password, ps.Host, ps.Port)
	db, err := sql.Open("postgres", url)
	if err != nil {
		return "unable to connect to database", err
	}
	ps.DB = db
	return "Connection to database successful", nil
}

// Disconnect from postgresserver
func (ps *PostgresServer) Disconnect() {
	ps.DB.Close()
	if ps.NewDB != nil {
		ps.NewDB.Close()
	}

}
