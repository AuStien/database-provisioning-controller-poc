package db

type SQLServer interface {
	Connect() (string, error)
	Disconnect()
	CreateUser() (string, error)
	DeleteUser() (string, error)
	CreateDatabase() (string, error)
	DeleteDatabase() (string, error)
	GrantPermissions() (string, error)
}
