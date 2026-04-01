package user

import (
	"errors"

	"gorm.io/gorm"
)

// Repository defines the interface for user data access.
type Repository interface {
	Create(user *User) error
	FindByEmail(email string) (*User, error)
	FindByID(id uint) (*User, error)
	FindAll() ([]User, error)
}

// repository is the GORM implementation of Repository.
type repository struct {
	db *gorm.DB
}

// NewRepository creates a new user repository.
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// Create inserts a new user into the database.
func (r *repository) Create(user *User) error {
	return r.db.Create(user).Error
}

// FindByEmail looks up a user by their email address.
func (r *repository) FindByEmail(email string) (*User, error) {
	var user User
	result := r.db.Where("email = ?", email).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &user, nil
}

// FindByID looks up a user by their primary key.
func (r *repository) FindByID(id uint) (*User, error) {
	var user User
	result := r.db.First(&user, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &user, nil
}

// FindAll returns all users from the database.
func (r *repository) FindAll() ([]User, error) {
	var users []User
	result := r.db.Find(&users)
	return users, result.Error
}
