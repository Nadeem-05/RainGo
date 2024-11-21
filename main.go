package main

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"log"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"golang.org/x/crypto/ripemd160"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var assets embed.FS

type Entry struct {
	ID       int    `json:"id"`
	Password string `json:"pwd"`
	Hash     string `json:"hash"`
	Type     string `json:"type"`
	Source   string `json:"source"`
}

func computeMD5(password string) string {
	hash := md5.Sum([]byte(password))
	return hex.EncodeToString(hash[:])
}
func computeSHA1(password string) string {
	hash := sha1.Sum([]byte(password))
	return hex.EncodeToString(hash[:])
}
func computeSHA256(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}
func computeRIPEMD(password string) string {
	hasher := ripemd160.New()
	hasher.Write([]byte(password))
	return hex.EncodeToString(hasher.Sum(nil))
}

func (a *App) connectDB() {
	// PostgreSQL connection
	dsn := "host=localhost user=driver password=riyalsecret dbname=hash port=5432 sslmode=disable TimeZone=Asia/Shanghai"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	// SQLite connection
	localdb, err := gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to SQLite: %v", err)
	}
	// Configure PostgreSQL
	sqldb, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to configure PostgreSQL: %v", err)
	}
	sqldb.SetMaxIdleConns(10)
	sqldb.SetMaxOpenConns(100)
	sqldb.SetConnMaxLifetime(time.Hour)

	// Migrate schemas
	if err := db.AutoMigrate(&Entry{}); err != nil {
		log.Fatalf("Failed to migrate PostgreSQL schema: %v", err)
	}
	if err := localdb.AutoMigrate(&Entry{}); err != nil {
		log.Fatalf("Failed to migrate SQLite schema: %v", err)
	}

	a.db = db
	a.localdb = localdb
	log.Println("Databases connected successfully!")
}
func (a *App) GetPassword(hash string) string {
	var entry Entry

	// Check local database
	result := a.localdb.Where("hash = ?", hash).First(&entry)
	if result.Error == nil {
		return entry.Password
	}

	// If not found in local, check PostgreSQL database
	result = a.db.Where("hash = ?", hash).First(&entry)
	if result.Error != nil {
		runtime.LogError(a.ctx, "Error fetching password: "+result.Error.Error())
		return "Failed"
	}
	return entry.Password
}

func (a *App) GetEntries(page int, usePostgres bool) []Entry {
	var entries []Entry
	offset := (page - 1) * 10

	var dbToQuery *gorm.DB
	if usePostgres {
		dbToQuery = a.db
	} else {
		dbToQuery = a.localdb
	}

	err := dbToQuery.Limit(10).Offset(offset).Find(&entries).Error
	if err != nil {
		runtime.LogError(a.ctx, "Error fetching entries: "+err.Error())
		return nil
	}
	return entries
}

func (a *App) StartHashing(passwords []string) {
	passwordChan := make(chan string, len(passwords))
	resultChan := make(chan Entry, len(passwords))

	workers := 6
	var wg sync.WaitGroup

	worker := func() {
		defer wg.Done()
		for password := range passwordChan {
			resultChan <- Entry{Password: password, Hash: computeMD5(password), Type: "MD5"}
			resultChan <- Entry{Password: password, Hash: computeSHA1(password), Type: "SHA1"}
			resultChan <- Entry{Password: password, Hash: computeSHA256(password), Type: "SHA256"}
			resultChan <- Entry{Password: password, Hash: computeRIPEMD(password), Type: "RIPEMD"}
		}
	}
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go worker()
	}
	go func() {
		for _, password := range passwords {
			passwordChan <- password
		}
		close(passwordChan)
	}()
	go func() {
		wg.Wait()
		close(resultChan)
	}()
	go func() {
		var batch []Entry
		batchSize := 10000
		for entry := range resultChan {
			batch = append(batch, entry)
			if len(batch) >= batchSize {
				if !a.AddEntries(batch) {
					runtime.LogError(a.ctx, "Error during batch insertion")
				}
				batch = batch[:0]
			}
		}
		if len(batch) > 0 {
			if !a.AddEntries(batch) {
				runtime.LogError(a.ctx, "Error during final batch insertion")
			}
		}
		log.Println("Hashing and database insertion completed.")
	}()
	runtime.EventsEmit(a.ctx, "hashingStarted", len(passwords))
}

func (a *App) GetTotalEntries(usePostgres bool) int64 {
	var total int64
	if usePostgres {
		err := a.db.Model(&Entry{}).Count(&total).Error
		if err != nil {
			runtime.LogError(a.ctx, "Error counting total entries: "+err.Error())
			return 0
		}
		return total
	}
	err := a.localdb.Model(&Entry{}).Count(&total).Error
	if err != nil {
		runtime.LogError(a.ctx, "Error counting total entries: "+err.Error())
		return 0
	}
	return total
}

func (a *App) AddEntries(newEntries []Entry) bool {
	batchSize := 1000 // Adjust batch size for optimal performance
	var batch []Entry

	for _, entry := range newEntries {
		var existing Entry
		err := a.localdb.Where("hash = ?", entry.Hash).First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			batch = append(batch, entry)
		} else if err != nil {
			runtime.LogError(a.ctx, "Error checking existing entry: "+err.Error())
			return false
		}
		// Insert batch when it reaches the batch size
		if len(batch) >= batchSize {
			if err := a.localdb.Create(&batch).Error; err != nil {
				runtime.LogError(a.ctx, "Error inserting batch: "+err.Error())
				return false
			}
			log.Printf("Inserted batch of %d entries", len(batch))
			batch = batch[:0]
		}
	}

	// Insert remaining entries
	if len(batch) > 0 {
		if err := a.localdb.Create(&batch).Error; err != nil {
			runtime.LogError(a.ctx, "Error inserting final batch: "+err.Error())
			return false
		}
		log.Printf("Inserted final batch of %d entries", len(batch))
	}

	return true
}

func main() {
	app := NewApp()
	err := wails.Run(&options.App{
		Title:  "Rainbow Password Cracker",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
			&Entry{},
		},
		CSSDragProperty:                  "--wails-draggable",
		CSSDragValue:                     "drag",
		EnableDefaultContextMenu:         false,
		EnableFraudulentWebsiteDetection: false,
		Logger:                           nil,
		LogLevel:                         logger.ERROR,
		LogLevelProduction:               logger.ERROR,
		Frameless:                        false,
		Mac: &mac.Options{
			TitleBar: &mac.TitleBar{
				TitlebarAppearsTransparent: true,
				HideTitle:                  false,
				HideTitleBar:               false,
				FullSizeContent:            false,
				UseToolbar:                 false,
				HideToolbarSeparator:       true,
				// OnFileOpen:                 app.onFileOpen,
				// OnUrlOpen:                  app.onUrlOpen,
			},
			Appearance:           mac.NSAppearanceNameDarkAqua,
			WebviewIsTransparent: true,
			WindowIsTranslucent:  false,
			About: &mac.AboutInfo{
				Title:   "Rainbow Password Cracker",
				Message: "idk",
				// Icon:    icon,
			},
		},
		Linux: &linux.Options{
			// Icon:                icon,
			WindowIsTranslucent: false,
			WebviewGpuPolicy:    linux.WebviewGpuPolicyAlways,
			ProgramName:         "Rainbow Password Cracker",
		},
		Debug: options.Debug{
			OpenInspectorOnStartup: false,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
