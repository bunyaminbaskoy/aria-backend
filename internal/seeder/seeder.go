package seeder

import (
	"fmt"
	"log"

	"gorm.io/gorm"

	"music-curation/internal/user"
	"music-curation/pkg/utils"
)

// SeedUsers adds 10 sample users to the database.
// This function is idempotent — it skips users that already exist.
func SeedUsers(db *gorm.DB) {
	users := []struct {
		Email    string
		Password string
	}{
		{"alice@example.com", "password123"},
		{"bob@example.com", "password123"},
		{"charlie@example.com", "password123"},
		{"diana@example.com", "password123"},
		{"eve@example.com", "password123"},
		{"frank@example.com", "password123"},
		{"grace@example.com", "password123"},
		{"hank@example.com", "password123"},
		{"ivy@example.com", "password123"},
		{"jack@example.com", "password123"},
	}

	seededCount := 0
	for _, u := range users {
		var existing user.User
		result := db.Where("email = ?", u.Email).First(&existing)
		if result.Error == nil {
			// User already exists, skip
			continue
		}

		hashedPassword, err := utils.HashPassword(u.Password)
		if err != nil {
			log.Printf("⚠️  Failed to hash password for %s: %v", u.Email, err)
			continue
		}

		newUser := user.User{
			Email:    u.Email,
			Password: hashedPassword,
		}

		if err := db.Create(&newUser).Error; err != nil {
			log.Printf("⚠️  Failed to seed user %s: %v", u.Email, err)
			continue
		}

		seededCount++
	}

	if seededCount > 0 {
		fmt.Printf("🌱 Seeded %d users successfully\n", seededCount)
	} else {
		fmt.Println("🌱 All seed users already exist, skipping")
	}
}
