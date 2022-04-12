package database

import (
	"os"

	"github.com/Oloruntobi1/qgdc/internal/models"
)

type FileSystem struct {
	Path string
	File *os.File
}

var _ Repository = (*FileSystem)(nil)

func NewFileSystem(path string) *FileSystem {
	return &FileSystem{
		Path: path,
	}
}

// implement Repository interface
func (fs *FileSystem) Open() error {
	src, err := os.OpenFile(fs.Path, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	fs.File = src
	return nil
}

func (fs *FileSystem) Close() error {
	return fs.File.Close()

}

// implement Reader interface
func (fs *FileSystem) GetWallet(id int64) (*models.Wallet, error) {
	return nil, nil
}

func (fs *FileSystem) GetWalletByUserID(id int64) (*models.Wallet, error) {
	return nil, nil
}

func (fs *FileSystem) GetAllWallets() ([]*models.Wallet, error) {
	return nil, nil
}

func (fs *FileSystem) GetUserByEmail(email string) (*models.User, error) {
	return nil, nil
}

func (fs *FileSystem) GetAllUsers() ([]*UserWallet, error) {
	return nil, nil
}

// implement Updater interface
func (fs *FileSystem) CreateUser(user *models.User) error {
	return nil
}

// create wallet
func (fs *FileSystem) CreateWallet(wallet *models.Wallet) (int64, error) {
	return 0, nil
}

// update wallet
func (fs *FileSystem) UpdateWallet(wallet *models.Wallet) (*models.Wallet, error) {
	return nil, nil
}

// delete wallet
func (fs *FileSystem) DeleteWallet(id int64) error {
	return nil
}

// implement Seeder interface
func (fs *FileSystem) Seed() {

}

// implement Repository interface
func (fs *FileSystem) CreateTables() error {
	return nil
}
