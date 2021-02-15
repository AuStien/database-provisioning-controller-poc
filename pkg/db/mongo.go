package db

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"

	"go.mongodb.org/mongo-driver/mongo"
	_ "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Mongo object
type Mongo struct {
	Name     string
	Username string
	Password string
}

// MongoServer object
type MongoServer struct {
	Username string
	Password string
	Host     string
	Port     int32
	Ssl      string
	Mongo    Mongo
	Client   *mongo.Client
	DB       *mongo.Database
}

// CreateUser creates a user
func (ms *MongoServer) CreateUser() (string, error) {
	// Check if user exists on server
	if res := ms.DB.RunCommand(context.Background(), bson.D{
		{Key: "createUser", Value: ms.Mongo.Username},
		{Key: "pwd", Value: ms.Mongo.Password},
		{Key: "roles", Value: []bson.M{
			{"role": "readWrite",
				"db": ms.Mongo.Name}}}}); res.Err() != nil {
		return "unable to create user", res.Err()
	}
	return "User created successfully", nil
}

// DeleteUser from server
func (ms *MongoServer) DeleteUser() (string, error) {
	if res := ms.DB.RunCommand(context.Background(), bson.D{{Key: "dropUser", Value: ms.Mongo.Username}}); res.Err() != nil {
		return "unable to drop user", res.Err()
	}
	return "User dropped successfully", nil
}

// CreateDatabase creates a database
func (ms *MongoServer) CreateDatabase() (string, error) {
	// Try to create database
	ms.DB = ms.Client.Database(ms.Mongo.Name)

	return "Database created successfully", nil
}

// DeleteDatabase from server
func (ms *MongoServer) DeleteDatabase() (string, error) {
	if err := ms.DB.Drop(context.Background()); err != nil {
		return "unable to delete database", err
	}
	return "Database deleted successfully", nil
}

// GrantPermissions to user
func (ms *MongoServer) GrantPermissions() (string, error) {
	// Grant permissions to user
	if res := ms.DB.RunCommand(context.Background(), bson.D{
		{Key: "grantRolesToUser", Value: ms.Mongo.Username},
		{Key: "roles", Value: []bson.M{
			{"role": "readWrite",
				"db": ms.Mongo.Name}}}}); res.Err() != nil {
		return "unable to grant permissions", res.Err()
	}
	return "Permissions successfully granted", nil
}

// Connect to Mongoserver
func (ms *MongoServer) Connect() (string, error) {
	url := fmt.Sprintf("mongodb://%s:%s@%s:%d/?ssl=%s", ms.Username, ms.Password, ms.Host, ms.Port, ms.Ssl)
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(url))
	if err != nil {
		return "unable to connect to database", err
	}

	ms.Client = client
	ms.DB = ms.Client.Database(ms.Mongo.Name)

	return "Connection to database successful", nil
}

// Disconnect from mongoserver
func (ms *MongoServer) Disconnect() {
	ms.Client.Disconnect(context.Background())
}
