package db

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

// Mysql object
type Mysql struct {
	Name     string
	Username string
	Password string
}

// MysqlServer object
type MysqlServer struct {
	Username string
	Password string
	Host     string
	Port     int32
	SslMode  string
	Mysql    Mysql
	DB       *sql.DB
}

// CreateUser creates a user
func (ms *MysqlServer) CreateUser() (string, error) {
	// Check if user exists on server
	var user string
	ms.DB.QueryRow(fmt.Sprintf("SELECT user FROM mysql.user WHERE user='%s'", ms.Mysql.Username)).Scan(&user)

	// If user doesn't exist create new
	if user == "" {
		_, err := ms.DB.Exec(fmt.Sprintf("CREATE USER '%s'@'%s' IDENTIFIED BY '%s'", ms.Mysql.Username, ms.Host, ms.Mysql.Password))
		if err != nil {
			return "unable to create role in database", err
		}
		return "User created successfully", nil

		// If user exists update password
	} else {
		_, err := ms.DB.Exec(fmt.Sprintf("ALTER USER '%s'@'%s' IDENTIFIED BY '%s'", ms.Mysql.Username, ms.Host, ms.Mysql.Password))
		if err != nil {
			return "unable to alter user in database", err
		}
		return "User altered successfully", nil

	}

}

// DeleteUser from server
func (ms *MysqlServer) DeleteUser() (string, error) {
	_, err := ms.DB.Exec(fmt.Sprintf("DROP USER IF EXISTS '%s'@'%s'", ms.Mysql.Username, ms.Host))
	if err != nil {
		return "unable to drop user in database server", err
	}
	return "User deleted successfully", nil
}

// CreateDatabase creates a database
func (ms *MysqlServer) CreateDatabase() (string, error) {
	// Try to create database
	_, err := ms.DB.Exec(fmt.Sprintf("CREATE DATABASE %s", ms.Mysql.Name))
	if err != nil {
		if !strings.Contains(err.Error(), "exists") {
			return "unable to create database in database server", err
		}
	}
	return "Database created successfully", nil
}

// DeleteDatabase from server
func (ms *MysqlServer) DeleteDatabase() (string, error) {
	_, err := ms.DB.Exec(fmt.Sprintf("DROP USER %s@'%s'", ms.Mysql.Username, ms.Host))
	if err != nil {
		return "unable to drop user in database server", err
	}
	return "Database deleted successfully", nil
}

// GrantPermissions to user
func (ms *MysqlServer) GrantPermissions() (string, error) {
	// Grant permissions to user
	_, err := ms.DB.Exec(fmt.Sprintf("GRANT ALL PRIVILEGES ON %s.* TO '%s'@'%s'", ms.Mysql.Name, ms.Mysql.Username, ms.Host))
	if err != nil {
		return "unable to grant permissions in database", err
	}
	return "Permissions successfully granted", nil
}

// Connect to postgresserver
func (ms *MysqlServer) Connect() (string, error) {
	url := fmt.Sprintf("%s:%s@tcp(%s:%d)/mysql?tls=%s", ms.Username, ms.Password, ms.Host, ms.Port, ms.SslMode)
	db, err := sql.Open("mysql", url)
	if err != nil {
		return "unable to connect to database", err
	}
	ms.DB = db
	return "Connection to database successful", nil
}

// Disconnect from postgresserver
func (ms *MysqlServer) Disconnect() {
	ms.DB.Close()
}
