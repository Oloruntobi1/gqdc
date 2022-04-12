package database

import (
	"log"
	"os"
	"time"

	"github.com/Oloruntobi1/qgdc/internal/models"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// merge models.User and models.Wallet
type UserWallet struct {
	ID   int64     `json:"id"`
	UUID uuid.UUID `json:"uuid"`
	// Saving the password in clear text for YOUR testing purpose via Postman etc.
	Password          string          `json:"password"`
	HashedPassword    string          `json:"hashed_password"`
	FullName          string          `json:"full_name"`
	Email             string          `json:"email"`
	PasswordChangedAt *time.Time      `json:"password_changed_at"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
	WalletID          int64           `json:"wallet_id"`
	WalletBalance     decimal.Decimal `json:"wallet_balance"`
}

type Reader interface {
	GetWallet(id int64) (*models.Wallet, error)
	GetWalletByUserID(userID int64) (*models.Wallet, error)
	GetAllWallets() ([]*models.Wallet, error)
	GetUserByEmail(email string) (*models.User, error)
	GetAllUsers() ([]*UserWallet, error)
}

type Updater interface {
	CreateUser(user *models.User) error
	CreateWallet(wallet *models.Wallet) (int64, error)
	UpdateWallet(wallet *models.Wallet) (*models.Wallet, error)
	DeleteWallet(id int64) error
}

type Seeder interface {
	Seed()
}

type Repository interface {
	Reader
	Updater
	Seeder
	Open() error
	Close() error
	CreateTables() error
}

func NewRepository(storage string) Repository {
	db := getRepoToBeUsed(storage)
	// this opens a connection to the underlying db in use
	// if not inmemory, this will open a connection to the db
	// if inmemory, this will not open a connection
	_, ok := db.(*InMemory)
	if !ok {
		err := db.Open()
		if err != nil {
			log.Fatal("cannot connect to db:", err)
		}
		// this runs the migration to create the tables
		err = db.CreateTables()
		if err != nil {
			log.Fatal("cannot create tables:", err)
		}
	}
	// seed the database with some data
	db.Seed()

	return db
}

func getRepoToBeUsed(storage string) Repository {
	switch {
	case storage == "mysql":
		mysql := NewMySQL()
		return mysql
	case storage == "filesystem":
		fs := NewFileSystem(os.Getenv("FILE_SYSTEM_PATH"))
		return fs
	default:
		inMemory := NewInMemory()
		return inMemory
	}

}
