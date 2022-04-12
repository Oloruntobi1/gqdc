package database

import (
	"time"

	"github.com/Oloruntobi1/qgdc/internal/models"
	"github.com/Oloruntobi1/qgdc/util"

	"github.com/google/uuid"
)

type InMemory struct {
	Users   []*models.User
	Wallets []*models.Wallet
}

var _ Repository = (*InMemory)(nil)

func NewInMemory() *InMemory {
	return &InMemory{
		Users:   []*models.User{},
		Wallets: []*models.Wallet{},
	}
}

// implement Reader interface
func (m *InMemory) GetUserByEmail(email string) (*models.User, error) {
	for _, user := range m.Users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, util.ErrUserNotFound
}

func (m *InMemory) GetWallet(id int64) (*models.Wallet, error) {
	for _, wallet := range m.Wallets {
		if wallet.ID == id {
			return wallet, nil
		}
	}
	return nil, util.ErrWalletNotFound
}

func (m *InMemory) GetAllWallets() ([]*models.Wallet, error) {
	return m.Wallets, nil
}

func (m *InMemory) GetWalletByUserID(userID int64) (*models.Wallet, error) {
	for _, wallet := range m.Wallets {
		if wallet.UserID == userID {
			return wallet, nil
		}
	}
	return nil, util.ErrWalletNotFound
}

func (m *InMemory) GetAllUsers() ([]*UserWallet, error) {
	var userWallets []*UserWallet
	for _, user := range m.Users {
		wallet, err := m.GetWalletByUserID(user.ID)
		if err != nil {
			return nil, err
		}
		userWallets = append(userWallets, &UserWallet{
			ID:                user.ID,
			UUID:              user.UUID,
			FullName:          user.FullName,
			Email:             user.Email,
			Password:          user.Password,
			HashedPassword:    user.HashedPassword,
			PasswordChangedAt: user.PasswordChangedAt,
			CreatedAt:         user.CreatedAt,
			UpdatedAt:         user.UpdatedAt,
			WalletID:          wallet.ID,
			WalletBalance:     wallet.Balance,
		})
	}

	return userWallets, nil
}

// implement Updater interface
func (m *InMemory) CreateUser(user *models.User) error {
	m.Users = append(m.Users, user)
	return nil
}

func (m *InMemory) CreateWallet(wallet *models.Wallet) (int64, error) {
	m.Wallets = append(m.Wallets, wallet)
	return wallet.ID, nil
}

func (m *InMemory) UpdateWallet(wallet *models.Wallet) (*models.Wallet, error) {
	// get wallet by id
	for i, w := range m.Wallets {
		if w.ID == wallet.ID {
			m.Wallets[i].Balance = wallet.Balance
		}
	}
	return wallet, nil
}

func (m *InMemory) DeleteWallet(id int64) error {
	// get wallet by id
	for i, w := range m.Wallets {
		if w.ID == id {
			m.Wallets = append(m.Wallets[:i], m.Wallets[i+1:]...)
		}
	}
	return nil
}

// implement Repository interface
func (m *InMemory) Open() error {
	return nil
}

func (m *InMemory) Close() error {
	return nil
}

func (m *InMemory) CreateTables() error {
	return nil
}

func (m *InMemory) Seed() {
	if len(m.Users) == 0 {
		var i int64
		for i = 1; i <= SEEDNUMBER; i++ {
			password := util.RandomString(6)
			hashedPassword, _ := util.HashPassword(password)
			user := &models.User{
				ID:             i,
				UUID:           uuid.New(),
				Password:       password,
				HashedPassword: hashedPassword,
				FullName:       util.RandomUserName(),
				Email:          util.RandomEmail(),
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			}
			m.Users = append(m.Users, user)
		}
	}

	if len(m.Wallets) == 0 {
		var i int64
		for i = 1; i <= SEEDNUMBER; i++ {
			wallet := &models.Wallet{
				ID:        i,
				UUID:      uuid.New(),
				UserID:    i,
				Balance:   util.RandomDecimal(),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			m.Wallets = append(m.Wallets, wallet)
		}
	}
}
