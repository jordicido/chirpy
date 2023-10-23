package Database

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"sync"

	"golang.org/x/crypto/bcrypt"
)

type DB struct {
	path string
	mux  *sync.RWMutex
}

type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
	Users  map[int]User  `json:"users"`
}

type Chirp struct {
	Id   int
	Body string
}

type User struct {
	Id       int
	Email    string
	Password string
}

const filename = "database.json"

// NewDB creates a new database connection
// and creates the database file if it doesn't exist
func NewDB(path string) (*DB, error) {
	var db = &DB{path: path, mux: &sync.RWMutex{}}
	_, err := os.Stat(path + filename)
	if os.IsNotExist(err) {
		db.ensureDB()
	}
	return db, nil
}

// CreateChirp creates a new chirp and saves it to disk
func (db *DB) CreateChirp(body string) (Chirp, error) {
	var chirp Chirp
	var dbStructure DBStructure
	chirps, err := db.GetChirps()
	if err != nil {
		return chirp, err
	}
	if len(chirps) > 0 {
		chirp.Id = chirps[len(chirps)-1].Id + 1
	} else {
		chirp.Id = 1
	}
	chirp.Body = body
	chirps = append(chirps, chirp)
	dbStructure.Chirps = make(map[int]Chirp, len(chirps))
	for i := 0; i < len(chirps); i++ {
		dbStructure.Chirps[i] = chirps[i]
	}
	db.writeDB(dbStructure)
	return chirp, nil
}

// GetChirps returns all chirps in the database
func (db *DB) GetChirps() ([]Chirp, error) {
	var chirps []Chirp
	dbStructure, err := db.loadDB()
	if err != nil {
		return chirps, err
	}
	for _, chirp := range dbStructure.Chirps {
		chirps = append(chirps, chirp)
	}
	return chirps, nil
}

func (db *DB) CreateUser(email, password string) (User, error) {
	var user User
	var dbStructure DBStructure
	users, err := db.GetUsers()
	if err != nil {
		return user, err
	}
	if len(users) > 0 {
		user.Id = users[len(users)-1].Id + 1
	} else {
		user.Id = 1
	}
	for _, user := range users {
		if user.Email == email {
			return user, errors.New("User already existing")
		}
	}
	user.Email = email
	encryptedPass, err := bcrypt.GenerateFromPassword([]byte(password), 0)
	if err != nil {
		return user, err
	}
	user.Password = string(encryptedPass)
	users = append(users, user)
	dbStructure.Users = make(map[int]User, len(users))
	for i := 0; i < len(users); i++ {
		dbStructure.Users[i] = users[i]
	}
	db.writeDB(dbStructure)
	return user, nil
}

func (db *DB) UpdateUser(id int, email, password string) (User, error) {
	var modUser User
	var dbStructure DBStructure
	users, err := db.GetUsers()
	if err != nil {
		return modUser, err
	}
	for i, user := range users {
		if user.Id == id {
			users = append(users[:i], users[i+1:]...)
			break
		}
	}
	modUser.Id = id
	modUser.Email = email
	encryptedPass, err := bcrypt.GenerateFromPassword([]byte(password), 0)
	if err != nil {
		return modUser, err
	}
	modUser.Password = string(encryptedPass)
	users = append(users, modUser)
	dbStructure.Users = make(map[int]User, len(users))
	for i := 0; i < len(users); i++ {
		dbStructure.Users[i] = users[i]
	}
	db.writeDB(dbStructure)
	return modUser, nil
}

func (db *DB) GetUsers() ([]User, error) {
	var users []User
	dbStructure, err := db.loadDB()
	if err != nil {
		return users, err
	}
	for _, user := range dbStructure.Users {
		users = append(users, user)
	}
	return users, nil
}

// ensureDB creates a new database file if it doesn't exist
func (db *DB) ensureDB() error {
	file, err := os.Create(db.path + filename)
	if err != nil {
		return err
	}
	defer file.Close()
	return nil
}

// loadDB reads the database file into memory
func (db *DB) loadDB() (DBStructure, error) {
	var dbStructure DBStructure
	db.mux.RLock()
	defer db.mux.RUnlock()
	data, err := os.ReadFile(db.path + filename)
	if err != nil {
		return dbStructure, err
	}
	json.Unmarshal(data, &dbStructure)
	return dbStructure, nil
}

// writeDB writes the database file to disk
func (db *DB) writeDB(dbStructure DBStructure) error {
	db.mux.Lock()
	defer db.mux.Unlock()
	data, err := json.Marshal(dbStructure)
	if err != nil {
		return err
	}
	os.WriteFile(db.path+filename, data, fs.FileMode(os.O_TRUNC))
	return nil
}
