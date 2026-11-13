package main

import (
	"context"
	"fmt"
	"log"

	"github.com/nyouaku/SOPQ/backend/internal/config"
	"github.com/nyouaku/SOPQ/backend/internal/infrastructure/postgres"
	"github.com/nyouaku/SOPQ/backend/internal/repository"
	"github.com/nyouaku/SOPQ/backend/pkg/utils"
)

func main() {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Connect to database
	db, err := postgres.NewPostgresDB(&cfg.Database)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Create repository
	userRepo := repository.NewUserRepository(db.DB)

	// Find user
	ctx := context.Background()
	fmt.Println("Searching for username: staff")

	// Test 1: Direct query without Preload
	var testUser struct {
		Username string
		PasswordHash string
		RoleID int
	}
	err = db.DB.WithContext(ctx).Table("users").Where("username = ?", "staff").First(&testUser).Error
	if err != nil {
		log.Fatal("Direct query failed:", err)
	}
	fmt.Printf("Direct query SUCCESS: username=%s, role_id=%d\n", testUser.Username, testUser.RoleID)

	// Test 2: Repository query with Preload
	user, err := userRepo.FindByUsername(ctx, "staff")
	if err != nil {
		log.Fatal("Repository query failed:", err)
	}
	if user == nil {
		log.Fatal("User not found from repository")
	}

	fmt.Printf("User found:\n")
	fmt.Printf("  ID: %s\n", user.ID)
	fmt.Printf("  Username: %s\n", user.Username)
	fmt.Printf("  FullName: %s\n", user.FullName)
	fmt.Printf("  RoleID: %d\n", user.RoleID)
	fmt.Printf("  IsActive: %v\n", user.IsActive)
	fmt.Printf("  PasswordHash: %s\n", user.PasswordHash)
	if user.Role != nil {
		fmt.Printf("  Role: %s\n", user.Role.Name)
	} else {
		fmt.Printf("  Role: <nil>\n")
	}

	// Test password verification
	password := "staff123"
	fmt.Printf("\nTesting password: %s\n", password)
	err = utils.VerifyPassword(user.PasswordHash, password)
	if err != nil {
		fmt.Printf("❌ Password verification FAILED: %v\n", err)
	} else {
		fmt.Printf("✅ Password verification SUCCESS\n")
	}
}
