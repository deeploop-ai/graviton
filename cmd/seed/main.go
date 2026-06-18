package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"github.com/deeploop-ai/fleet/internal/infra/bun/model"
	"github.com/deeploop-ai/fleet/internal/pkg/database"
	"github.com/deeploop-ai/fleet/pkg/idgen"
	"github.com/deeploop-ai/fleet/pkg/password"
	"github.com/joho/godotenv"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

func main() {
	_ = godotenv.Load()

	dsn := database.SourceFromEnv()
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	db := bun.NewDB(sqldb, pgdialect.New())
	defer db.Close()

	ctx := context.Background()

	// Default project
	project := &model.Project{
		ID:          "default",
		Name:        "Default Project",
		Description: "Auto-created seed project",
		Status:      "active",
		Settings:    map[string]any{},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if _, err := db.NewInsert().Model(project).On("CONFLICT (id) DO NOTHING").Exec(ctx); err != nil {
		fmt.Println("insert project:", err)
		os.Exit(1)
	}

	// Console admin
	hash, err := password.Hash("Admin@123")
	if err != nil {
		fmt.Println("hash password:", err)
		os.Exit(1)
	}
	admin := &model.ConsoleAdmin{
		ID:           idgen.UUID().String(),
		Email:        "admin@fleet.local",
		PasswordHash: hash,
		Role:         "owner",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if _, err := db.NewInsert().Model(admin).On("CONFLICT (email) DO NOTHING").Exec(ctx); err != nil {
		fmt.Println("insert admin:", err)
		os.Exit(1)
	}

	// Default API key for the default project.
	apiSecret := "fleet-default-api-key-" + idgen.UUID().String()
	apiHash := sha256.Sum256([]byte(apiSecret))
	apiKey := &model.APIKey{
		ID:         idgen.UUID().String(),
		ProjectID:  "default",
		Name:       "Default API Key",
		SecretHash: hex.EncodeToString(apiHash[:]),
		Scopes:     []string{"projects.read", "users.read", "users.write", "storage.read", "storage.write", "databases.read", "databases.write"},
		Enabled:    true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	if _, err := db.NewInsert().Model(apiKey).On("CONFLICT (id) DO NOTHING").Exec(ctx); err != nil {
		fmt.Println("insert api key:", err)
		os.Exit(1)
	}

	fmt.Println("seeded project=default admin=admin@fleet.local api_key=" + apiSecret)
}
