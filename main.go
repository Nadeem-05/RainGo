package main

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"log"
	"net/http"
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
	"github.com/wailsapp/wails/v2/pkg/options/windows"
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

//go:embed build/appicon.png
var icon []byte

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

func (a *App) connectDB() {
	data, _ := envFile.ReadFile(".env")
	myenv, err := godotenv.Parse(bytes.NewReader(data))
	if err != nil {
		log.Fatalf("failed to parse .env file: %v", err)
	}
	dsn := myenv["CONNECTION_STRING"]
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

	a.db = db
	a.localdb = localdb
	log.Printf("Databases connected successfully!")
}

func (a *App) UpdateStats(isSuccess bool) {
	log.Printf("Updating stats")
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
	var processedCount int64
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
		runtime.EventsEmit(a.ctx, "hashingCompleted")
		log.Println("Hashing and database insertion completed.")
	}()
}
func (a *App) GetMeta() [][]Meta {
	var wg sync.WaitGroup
	var mu sync.Mutex

	TotalMeta := make([][]Meta, 0, 3)
	pie := make([]Meta, 0, 4)
	donut := make([]Meta, 0, 2)
	bar := make([]Meta, 0, 2)

	wg.Add(1)
	go func() {
		defer wg.Done()

		type HashCount struct {
			Type  string
			Count int64
		}

		// Use the materialized view for faster query
		query := `
			SELECT "type", count 
			FROM type_counts;
		`

		var hashCounts []HashCount
		if err := a.db.Raw(query).Scan(&hashCounts).Error; err != nil {
			log.Printf("Error querying materialized view: %v", err)
			return
		}

		mu.Lock()
		for _, hc := range hashCounts {
			pie = append(pie, Meta{
				Point: hc.Type,
				Value: strconv.FormatInt(hc.Count, 10),
			})
		}
		TotalMeta = append(TotalMeta, pie)
		mu.Unlock()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		var localCount, globalCount int64

		// Local count (SQLite fallback if needed)
		if err := a.localdb.Model(&Entry{}).Count(&localCount).Error; err != nil {
			log.Printf("Error querying local count: %v", err)
		}

		// Approximate global count (or fallback to COUNT(*))
		query := `
			SELECT reltuples::BIGINT AS approximate_count 
			FROM pg_class 
			WHERE relname = 'entries';
		`
		if a.db.Dialector.Name() == "postgres" {
			if err := a.db.Raw(query).Scan(&globalCount).Error; err != nil {
				log.Printf("Error querying approximate count: %v", err)
			}
		} else {
			// Fallback for SQLite
			if err := a.db.Model(&Entry{}).Count(&globalCount).Error; err != nil {
				log.Printf("Error querying global count: %v", err)
			}
		}

		mu.Lock()
		donut = append(donut,
			Meta{Point: "LocalDB", Value: strconv.FormatInt(localCount, 10)},
			Meta{Point: "Postgres", Value: strconv.FormatInt(globalCount, 10)},
		)
		TotalMeta = append(TotalMeta, donut)
		mu.Unlock()
	}()

	wg.Wait()

	stats := a.GetHashStats()
	bar = append(bar,
		Meta{Point: "Success Rate", Value: strconv.Itoa(stats.SuccessRate)},
		Meta{Point: "Failure Rate", Value: strconv.Itoa(stats.FailureRate)},
	)

	mu.Lock()
	TotalMeta = append(TotalMeta, bar)
	mu.Unlock()

	return TotalMeta
}

func (a *App) GetTotalEntries(usePostgres bool) int64 {
	var total int64
	if usePostgres {
		query := `
		SELECT reltuples::BIGINT AS approximate_count 
		FROM pg_class 
		WHERE relname = 'entries';
	`
		if err := a.db.Raw(query).Scan(&total).Error; err != nil {
			log.Printf("Error querying approximate count: %v", err)
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
		Windows:                          &windows.Options{},
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
				Icon:    icon,
			},
			Preferences: &mac.Preferences{
				TextInteractionEnabled: mac.Disabled,
			},
		},
		Linux: &linux.Options{
			Icon:                icon,
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
