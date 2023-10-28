package Database

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type DB struct {
	path string
	mux  *sync.RWMutex
}

type DBStructure struct {
	Chirps      map[int]Chirp      `json:"chirps"`
	Users       map[int]User       `json:"users"`
	Revocations map[int]Revocation `json:"revocations"`
}

type Chirp struct {
	Id       int
	Body     string
	AuthorId int
}

type User struct {
	Id          int
	Email       string
	Password    string
	IsChirpyRed bool
}

type Revocation struct {
	Token string
	Time  time.Time
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
func (db *DB) CreateChirp(body string, author int) (Chirp, error) {
	var chirp Chirp
	dbStructure, err := db.loadDB()
	if err != nil {
		return chirp, err
	}
	chirps, err := db.GetChirps(nil)
	if err != nil {
		return chirp, err
	}
	if len(chirps) > 0 {
		chirp.Id = chirps[len(chirps)-1].Id + 1
	} else {
		chirp.Id = 1
	}
	chirp.Body = body
	chirp.AuthorId = author
	if len(dbStructure.Chirps) == 0 {
		dbStructure.Chirps = make(map[int]Chirp)
	}
	dbStructure.Chirps[chirp.Id-1] = chirp
	db.writeDB(dbStructure)
	return chirp, nil
}

func (db *DB) DeleteChirp(id, authorId int) error {
	var chirps []Chirp
	dbStructure, err := db.loadDB()
	if err != nil {
		return err
	}
	for _, chirp := range dbStructure.Chirps {
		chirps = append(chirps, chirp)
	}
	for i, chirp := range chirps {
		if chirp.Id == id && chirp.AuthorId == authorId {
			chirps = append(chirps[:i], chirps[i+1:]...)
			dbStructure.Chirps = make(map[int]Chirp, len(chirps))
			for i := 0; i < len(chirps); i++ {
				dbStructure.Chirps[i] = chirps[i]
			}
			db.writeDB(dbStructure)
			return nil
		}
	}
	return errors.New("Chirp not possible to delete")
}

func (db *DB) GetChirp(id int) (Chirp, error) {
	var chirp Chirp
	dbStructure, err := db.loadDB()
	if err != nil {
		return chirp, err
	}

	for _, chirp := range dbStructure.Chirps {
		if chirp.Id == id {
			return chirp, nil
		}
	}
	return chirp, errors.New("Chirp not found")
}

// GetChirps returns all chirps in the database
func (db *DB) GetChirps(authorId *int) ([]Chirp, error) {
	var chirps []Chirp
	dbStructure, err := db.loadDB()
	if err != nil {
		return chirps, err
	}

	sortByAuthor := false
	if authorId != nil {
		sortByAuthor = true
	}
	for _, chirp := range dbStructure.Chirps {
		if sortByAuthor && *authorId == chirp.AuthorId {
			chirps = append(chirps, chirp)
		} else if !sortByAuthor {
			chirps = append(chirps, chirp)
		}

	}
	return chirps, nil
}

func (db *DB) CreateUser(email, password string) (User, error) {
	var user User
	dbStructure, err := db.loadDB()
	if err != nil {
		return user, err
	}
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
	user.IsChirpyRed = false
	encryptedPass, err := bcrypt.GenerateFromPassword([]byte(password), 0)
	if err != nil {
		return user, err
	}
	user.Password = string(encryptedPass)
	if len(dbStructure.Users) == 0 {
		dbStructure.Users = make(map[int]User)
	}
	dbStructure.Users[user.Id-1] = user
	db.writeDB(dbStructure)
	return user, nil
}

func (db *DB) UpdateUser(id int, email string, password *string, isChirpyRed bool) (User, error) {
	var modUser User
	dbStructure, err := db.loadDB()
	if err != nil {
		return modUser, err
	}
	users, err := db.GetUsers()
	if err != nil {
		return modUser, err
	}
	for i, user := range users {
		if user.Id == id {
			users = append(users[:i], users[i+1:]...)
			if password != nil {
				encryptedPass, err := bcrypt.GenerateFromPassword([]byte(*password), 0)
				if err != nil {
					return modUser, err
				}
				modUser.Password = string(encryptedPass)
			} else {
				modUser.Password = user.Password
			}
			break
		}
	}
	modUser.Id = id
	modUser.Email = email
	modUser.IsChirpyRed = isChirpyRed
	users = append(users, modUser)
	dbStructure.Users = make(map[int]User, len(users))
	for i := 0; i < len(users); i++ {
		dbStructure.Users[i] = users[i]
	}
	db.writeDB(dbStructure)
	return modUser, nil
}

func (db *DB) GetUser(id int) (User, error) {
	var user User
	dbStructure, err := db.loadDB()
	if err != nil {
		return user, err
	}

	for _, user := range dbStructure.Users {
		if user.Id == id {
			return user, nil
		}
	}
	return user, errors.New("Chirp not found")
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

func (db *DB) RevokeToken(token string) error {
	var revocation Revocation
	dbStructure, err := db.loadDB()
	if err != nil {
		return err
	}
	revocations, err := db.GetRevocations()
	if err != nil {
		return err
	}
	revocation.Token = token
	revocation.Time = time.Now()
	revocations = append(revocations, revocation)
	if len(dbStructure.Revocations) == 0 {
		dbStructure.Revocations = make(map[int]Revocation)
	}
	dbStructure.Revocations[len(revocations)-1] = revocation
	db.writeDB(dbStructure)
	return nil
}

func (db *DB) GetRevocations() ([]Revocation, error) {
	var revocations []Revocation
	dbStructure, err := db.loadDB()
	if err != nil {
		return revocations, err
	}
	for _, revocation := range dbStructure.Revocations {
		revocations = append(revocations, revocation)
	}
	return revocations, nil
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
