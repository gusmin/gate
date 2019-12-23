// Package database provides
package database

import (
	"encoding/json"
	"path"

	"github.com/boltdb/bolt"
)

const (
	// Default database file name
	databaseName = "securegate.db"
	// Name of the bucket where all users are stored
	usersBucketName = "users"
)

// SecureGateBoltRepository is a database repository interacting
// with a key/value embedded and lightweight database
// called Bolt(Github: https://github.com/boltdb/bolt).
type SecureGateBoltRepository struct {
	// Database directory
	Path string

	// contains filtered or unexported fields
	db *bolt.DB
}

// NewSecureGateBoltRepository instanciates a new SecureGateBoltRepository
// who can communicate with a Bolt database at the given path.
func NewSecureGateBoltRepository(path string) *SecureGateBoltRepository {
	return &SecureGateBoltRepository{
		Path: path,
	}
}

// OpenDatabase opens the database located in the path that repo
// is tied to or creates a new database file if none exist.
// Then it creates all the bucket in the database if they still
// do not exist.
func (repo *SecureGateBoltRepository) OpenDatabase() error {
	// Open database or create one if none exist already.
	db, err := bolt.Open(path.Join(repo.Path, databaseName), 0666, nil)
	if err != nil {
		return err
	}

	// Create the top-level buckets.
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(usersBucketName))
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	repo.db = db

	return nil
}

// CloseDatabase closes the database.
// Make sure to call this method after you finished using the database.
func (repo *SecureGateBoltRepository) CloseDatabase() error {
	return repo.db.Close()
}

// User is the user model stored in the database.
type User struct {
	ID       string    `json:"id"`
	Machines []Machine `json:"machines"`
}

// Machine is the machine model stored in the database.
type Machine struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	IP        string `json:"ip"`
	AgentPort int    `json:"agentPort"`
}

// UpsertUser updates the user in the database or insert it if it
// do not exists already.
func (repo *SecureGateBoltRepository) UpsertUser(user User) error {
	// Struct values in the database are stored as JSON.
	userBytes, err := json.Marshal(&user)
	if err != nil {
		return err
	}

	err = repo.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(usersBucketName))
		if err != nil {
			return err
		}
		if err := b.Put([]byte(user.ID), userBytes); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

// GetUser retrieves the user owning the given userID in the database.
func (repo *SecureGateBoltRepository) GetUser(userID string) (User, error) {
	var user User

	err := repo.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(usersBucketName))
		v := b.Get([]byte(userID))

		// Struct values in the database are stored as JSON.
		err := json.Unmarshal(v, &user)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return User{}, err
	}

	return user, nil
}
