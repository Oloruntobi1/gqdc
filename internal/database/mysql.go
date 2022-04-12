package database

import (
	"fmt"
	"os"
	"time"

	"github.com/Oloruntobi1/qgdc/internal/models"

	"github.com/Oloruntobi1/qgdc/util"

	"github.com/google/uuid"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type MySQL struct {
	DB *gorm.DB
}

var _ Repository = (*MySQL)(nil)

const (
	SEEDNUMBER = 10
)

// get dsn from env

func (*MySQL) GetDSN() string {
	if os.Getenv("PLATFORM") == "docker" {
		return fmt.Sprintf(
			"%s:%s@tcp(%v:%v)/%v?charset=utf8mb4&parseTime=True&loc=%s",
			os.Getenv("MYSQL_USER"),
			os.Getenv("MYSQL_PASSWORD"),
			os.Getenv("MYSQL_CONTAINER_NAME"),
			os.Getenv("MYSQL_PORT"),
			os.Getenv("MYSQL_DATABASE"),
			os.Getenv("MYSQL_LOCAL"),
		)
	}
	return fmt.Sprintf(
		"%s:%s@tcp(%v:%v)/%v?charset=utf8mb4&parseTime=True&loc=%s",
		os.Getenv("MYSQL_USER"),
		os.Getenv("MYSQL_PASSWORD"),
		os.Getenv("MYSQL_HOST"),
		os.Getenv("MYSQL_PORT"),
		os.Getenv("MYSQL_DATABASE"),
		os.Getenv("MYSQL_LOCAL"),
	)
}

func (m *MySQL) Open() error {
	db, err := gorm.Open(mysql.Open(m.GetDSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return util.NewConnectionError(err)
	}
	m.DB = db
	return nil
}

func (m *MySQL) Close() error {
	db, err := m.DB.DB()
	if err != nil {
		return err
	}
	return db.Close()
}

func (m *MySQL) GetDB() *gorm.DB {
	return m.DB
}

// implement Reader interface
func (m *MySQL) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	err := m.DB.Where("email = ?", email).First(&user).Error
	if err == gorm.ErrRecordNotFound {
		return nil, util.ErrUserNotFound
	}
	return &user, err
}

func (m *MySQL) CreateUser(user *models.User) error {
	result := m.DB.Create(user)
	if result != nil {
		return result.Error
	}
	return nil
}

func (m *MySQL) CreateWallet(wallet *models.Wallet) (int64, error) {
	result := m.DB.Create(wallet)
	if result != nil {
		return 0, result.Error
	}
	return wallet.ID, nil
}

func (m *MySQL) GetWallet(id int64) (*models.Wallet, error) {
	var wallet models.Wallet
	err := m.DB.First(&wallet, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, util.ErrWalletNotFound
	}
	return &wallet, err
}

func (m *MySQL) GetWalletByUserID(userID int64) (*models.Wallet, error) {
	var wallet models.Wallet
	err := m.DB.Where("user_id = ?", userID).First(&wallet).Error
	return &wallet, err
}

func (m *MySQL) GetAllWallets() ([]*models.Wallet, error) {
	var wallets []*models.Wallet
	err := m.DB.Find(&wallets).Error
	return wallets, err
}

// create foreign key constraints
func (m *MySQL) CreateFK() error {
	err := m.DB.Exec(`
		ALTER TABLE wallets
		ADD CONSTRAINT fk_wallets_users
		FOREIGN KEY (user_id)
		REFERENCES users(id)
		ON DELETE CASCADE
	`).Error
	if err != nil {
		return util.NewCreateFKError(err)
	}
	return nil
}

func (m *MySQL) GetAllUsers() ([]*UserWallet, error) {
	// join user and wallet tables
	var users []*UserWallet
	err := m.DB.Raw(`
		SELECT u.id, u.uuid, u.full_name,  u.email, u.password, u.hashed_password, u.password_changed_at,
		u.created_at, u.updated_at, w.id as wallet_id, w.balance as wallet_balance
		FROM users u
		INNER JOIN wallets w ON u.id = w.user_id
	`).Scan(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (m *MySQL) UpdateWallet(wallet *models.Wallet) (*models.Wallet, error) {
	err := m.DB.Save(wallet).Error
	if err != nil {
		return nil, err
	}
	return wallet, nil
}

func (m *MySQL) DeleteWallet(id int64) error {
	return m.DB.Delete(&models.Wallet{}, id).Error
}

func NewMySQL() *MySQL {
	return &MySQL{}
}

func (m *MySQL) CreateTables() error {
	err := m.DB.AutoMigrate(&models.User{}, &models.Wallet{})
	if err != nil {
		return util.NewCreateSchemaError(err)
	}
	return nil
}

func (m *MySQL) Seed() {
	// seed random users
	seedUsers(m.GetDB())

	// seed random wallet data
	seedWallets(m.GetDB())

	// create foreign key constraints
	err := m.CreateFK()
	if err != nil {
		fmt.Println(err)
	}
}

func seedWallets(db *gorm.DB) {
	result := db.First(&models.Wallet{})
	if result.RowsAffected > 0 { // table already seeded
		return
	}
	var wallets []models.Wallet
	users := db.Find(&models.User{})
	if users.RowsAffected > 0 {
		var i int64
		for i = 1; i <= SEEDNUMBER; i++ {
			wallet := models.Wallet{
				ID:        i,
				UUID:      uuid.New(),
				UserID:    i,
				Balance:   util.RandomDecimal(),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			wallets = append(wallets, wallet)
		}
	}
	db.CreateInBatches(wallets, 10)
}

func seedUsers(db *gorm.DB) {
	result := db.First(&models.User{})
	if result.RowsAffected > 0 { // table already seeded
		return
	}
	var users []models.User
	var i int64
	for i = 1; i <= SEEDNUMBER; i++ {
		password := util.RandomString(6)
		hashedPassword, _ := util.HashPassword(password)
		user := models.User{
			ID:             i,
			UUID:           uuid.New(),
			Password:       password,
			HashedPassword: hashedPassword,
			FullName:       util.RandomUserName(),
			Email:          util.RandomEmail(),
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}
		users = append(users, user)
	}
	db.CreateInBatches(users, 10)
}
