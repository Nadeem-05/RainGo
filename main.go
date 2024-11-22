package main

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/joho/godotenv"
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

//go:embed all:frontend/build
var assets embed.FS

//go:embed .env
var envFile embed.FS

type Entry struct {
	ID       int    `json:"id"`
	Password string `json:"pwd"`
	Hash     string `json:"hash"`
	Type     string `json:"type"`
	Source   string `json:"source"`
}

type Meta struct {
	Point string `json:"point"`
	Value string `json:"value"`
}

type HashStats struct {
	ID          uint      `gorm:"primaryKey"`
	SuccessRate int       `gorm:"column:success_rate"`
	FailureRate int       `gorm:"column:failure_rate"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
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
func loadEnv() error {
	file, err := envFile.Open(".env")
	if err != nil {
		return err
	}
	_, err = godotenv.Parse(file)
	return err
}

func (a *App) connectDB() {
	loadEnv()
	dsn := os.Getenv("CONNECTION_STRING")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	localdb, err := gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to SQLite: %v", err)
	}
	sqldb, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to configure PostgreSQL: %v", err)
	}
	sqldb.SetMaxIdleConns(10)
	sqldb.SetMaxOpenConns(100)
	sqldb.SetConnMaxLifetime(time.Hour)

	if err := db.AutoMigrate(&Entry{}); err != nil {
		log.Fatalf("Failed to migrate PostgreSQL schema: %v", err)
	}
	if err := localdb.AutoMigrate(&Entry{}); err != nil {
		log.Fatalf("Failed to migrate SQLite schema: %v", err)
	}
	if err := db.AutoMigrate(&HashStats{}); err != nil {
		log.Fatalf("Failed to migrate PostgreSQL schema for HashStats: %v", err)
	}
	a.db = db
	a.localdb = localdb
	log.Println("Databases connected successfully!")
}

func (a *App) UpdateStats(isSuccess bool) {
	fmt.Print("Updating stats")
	var stats HashStats
	if err := a.db.First(&stats).Error; err == gorm.ErrRecordNotFound {
		stats = HashStats{SuccessRate: 0, FailureRate: 0}
		if err := a.db.Create(&stats).Error; err != nil {
			runtime.LogError(a.ctx, "Error creating HashStats record: "+err.Error())
			return
		}
	} else if err != nil {
		runtime.LogError(a.ctx, "Error fetching HashStats: "+err.Error())
		return
	}

	// Update the stats
	if isSuccess {
		stats.SuccessRate++
	} else {
		stats.FailureRate++
	}

	if err := a.db.Save(&stats).Error; err != nil {
		runtime.LogError(a.ctx, "Error updating HashStats: "+err.Error())
	}
}

func scraper(hash string, hashtype string) (string, error) {
	var url string
	if hashtype == "SHA1" {
		url = "https://sha1.gromweb.com/?sha1=" + hash
	} else {
		url = "https://md5.gromweb.com/?md5=" + hash
	}
	res, err := http.Get(url)
	if err != nil {
		log.Fatalf("Error fetching the URL: %v", err)
		return "", err
	}
	defer res.Body.Close()
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatalf("Error loading HTML document: %v", err)
		return "", err
	}
	word := doc.Find(".String").First().Text()
	if word != "" {
		return word, nil
	} else {
		return "", nil
	}

}
func (a *App) GetPassword(hash string) string {
	var entry Entry
	if err := a.localdb.Where("hash = ?", hash).First(&entry).Error; err == nil {
		a.UpdateStats(true)
		return entry.Password
	}

	log.Printf("Failed to fetch from localdb")

	if err := a.db.Where("hash = ?", hash).First(&entry).Error; err == nil {
		a.UpdateStats(true)
		return entry.Password
	}
	log.Printf("Failed to fetch from both databases")

	hashType := getHashType(hash)
	if hashType == "Unknown" {
		runtime.LogError(a.ctx, "Unknown hash type for: "+hash)
		a.UpdateStats(false)
		return "Failed"
	}
	result, err := scraper(hash, hashType)
	if result == "" || err != nil || result == "password" {
		runtime.LogError(a.ctx, "Error scraping password: ")
		a.UpdateStats(false)
		return "Failed"
	}
	entry.Password = result
	a.UpdateStats(true)
	return entry.Password
}

func getHashType(hash string) string {
	switch len(hash) {
	case 32:
		return "MD5"
	case 40:
		return "SHA1"
	default:
		return "Unknown"
	}
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
func (a *App) GetHashStats() HashStats {
	var stats HashStats
	if err := a.db.First(&stats).Error; err != nil {
		runtime.LogError(a.ctx, "Error fetching HashStats: "+err.Error())
	}
	return stats
}

func (a *App) StartHashing(passwords []string) {
	runtime.EventsEmit(a.ctx, "hashingStarted", len(passwords))

	passwordChan := make(chan string, len(passwords))
	resultChan := make(chan Entry, len(passwords))

	workers := 4
	var wg sync.WaitGroup
	var processedCount int64 // Tracks progress

	worker := func() {
		defer wg.Done()
		for password := range passwordChan {
			resultChan <- Entry{Password: password, Hash: computeMD5(password), Type: "MD5"}
			resultChan <- Entry{Password: password, Hash: computeSHA1(password), Type: "SHA1"}
			resultChan <- Entry{Password: password, Hash: computeSHA256(password), Type: "SHA256"}
			resultChan <- Entry{Password: password, Hash: computeRIPEMD(password), Type: "RIPEMD"}

			atomic.AddInt64(&processedCount, 1)
			runtime.EventsEmit(a.ctx, "hashingProgress", map[string]int{
				"Processed": int(atomic.LoadInt64(&processedCount)),
				"Total":     len(passwords),
			})
		}
	}

	// Start workers
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go worker()
	}

	// Producer goroutine
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

		// Emit completion event
		runtime.EventsEmit(a.ctx, "hashingCompleted")
		log.Println("Hashing and database insertion completed.")
	}()
}
func (a *App) GetMeta() [][]Meta {
	var wg sync.WaitGroup

	pieResultChan := make(chan []Meta, 1)
	donutresultchan := make(chan []Meta, 1)

	var TotalMeta [][]Meta

	wg.Add(1)
	go func() {
		defer wg.Done()

		pie := make([]Meta, 0, 4)

		hashTypes := []string{"MD5", "SHA1", "SHA256", "RIPEMD"}

		type HashCount struct {
			Type  string `gorm:"column:type"`
			Count int64  `gorm:"column:count"`
		}

		// Helper function to query a database and return the results
		queryDatabase := func(db *gorm.DB) ([]HashCount, error) {
			var hashCounts []HashCount
			err := db.Model(&Entry{}).
				Select("type", "COUNT(*) as count").
				Where("type IN (?)", hashTypes).
				Group("type").
				Find(&hashCounts).Error
			return hashCounts, err
		}

		// Query the main database
		mainHashCounts, err := queryDatabase(a.db)
		if err != nil {
			log.Printf("Error querying hash types from main DB: %v", err)
			pieResultChan <- pie
			return
		}

		// Query the local database
		localHashCounts, err := queryDatabase(a.localdb)
		if err != nil {
			log.Printf("Error querying hash types from local DB: %v", err)
			pieResultChan <- pie
			return
		}

		// Combine results from both databases
		countMap := make(map[string]int64)
		for _, hc := range mainHashCounts {
			countMap[hc.Type] += hc.Count
		}
		for _, hc := range localHashCounts {
			countMap[hc.Type] += hc.Count
		}

		// Populate the pie chart data
		for _, hashType := range hashTypes {
			count := countMap[hashType]
			pie = append(pie, Meta{
				Point: hashType,
				Value: strconv.FormatInt(count, 10),
			})
		}

		pieResultChan <- pie
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()

		donut := make([]Meta, 0, 2)

		var localdbCount, dbCount int64
		localResult := a.localdb.Model(&Entry{}).Count(&localdbCount)
		dbResult := a.db.Model(&Entry{}).Count(&dbCount)

		if localResult.Error != nil {
			log.Printf("Error querying localdb: %v", localResult.Error)
		}
		if dbResult.Error != nil {
			log.Printf("Error querying db: %v", dbResult.Error)
		}

		donut = append(donut,
			Meta{Point: "LocalDB", Value: strconv.FormatInt(localdbCount, 10)},
			Meta{Point: "Postgres", Value: strconv.FormatInt(dbCount, 10)},
		)

		donutresultchan <- donut
	}()

	wg.Wait()

	close(pieResultChan)
	close(donutresultchan)

	var bar []Meta
	stats := a.GetHashStats()
	bar = append(bar, Meta{Point: "Success Rate", Value: strconv.Itoa(stats.SuccessRate)})
	bar = append(bar, Meta{Point: "Failure Rate", Value: strconv.Itoa(stats.FailureRate)})
	TotalMeta = append(TotalMeta, <-pieResultChan)
	TotalMeta = append(TotalMeta, <-donutresultchan)
	TotalMeta = append(TotalMeta, bar)
	return TotalMeta
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
		if len(batch) >= batchSize {
			if err := a.localdb.Create(&batch).Error; err != nil {
				runtime.LogError(a.ctx, "Error inserting batch: "+err.Error())
				return false
			}
			log.Printf("Inserted batch of %d entries", len(batch))
			batch = batch[:0]
		}
	}
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
		Title:  "Raingo",
		Width:  1920,
		Height: 1080,
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
				Title:   "Raingo",
				Message: "idk",
				// Icon:    icon,
			},
		},
		Linux: &linux.Options{
			// Icon:                icon,
			WindowIsTranslucent: false,
			WebviewGpuPolicy:    linux.WebviewGpuPolicyAlways,
			ProgramName:         "Raingo",
		},
		Debug: options.Debug{
			OpenInspectorOnStartup: false,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
