package main

import (
	"embed"
	"log"
	"os"
	"path/filepath"

	"ficha-tracker/internal/repository"
	"ficha-tracker/internal/service"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed db/schema.sql
var schemaSQL string

func main() {
	dbPath, err := dataFilePath("fichas.db")
	if err != nil {
		log.Fatalf("resolving database path: %v", err)
	}

	db, err := repository.OpenDB(dbPath, schemaSQL)
	if err != nil {
		log.Fatalf("opening database: %v", err)
	}
	defer db.Close()

	if err := repository.Migrate(db); err != nil {
		log.Fatalf("migrating database: %v", err)
	}

	fichaRepo := repository.NewSQLiteFichaRepository(db)
	acsRepo := repository.NewSQLiteACSRepository(db)
	userRepo := repository.NewSQLiteUserRepository(db)

	fichaService := service.NewFichaService(fichaRepo, acsRepo)
	acsService := service.NewACSService(acsRepo, fichaRepo)
	authService := service.NewAuthService(userRepo)

	app := NewApp(fichaService, acsService, authService)

	err = wails.Run(&options.App{
		Title:  "ficha-tracker",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 244, G: 248, B: 253, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		log.Fatalf("running app: %v", err)
	}
}

// dataFilePath returns the path to name inside the user's config directory,
// under a ficha-tracker subfolder, creating it if needed.
func dataFilePath(name string) (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	appDir := filepath.Join(configDir, "ficha-tracker")
	if err := os.MkdirAll(appDir, 0o755); err != nil {
		return "", err
	}

	return filepath.Join(appDir, name), nil
}
