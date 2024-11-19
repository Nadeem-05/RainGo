package main

import (
	"context"
	"fmt"
	"log"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"gorm.io/gorm"
)

type App struct {
	ctx context.Context
	db  *gorm.DB
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	runtime.EventsOn(ctx, "StartHashing", func(data ...interface{}) {
		if len(data) > 0 {
			if passwords, ok := data[0].([]interface{}); ok {
				var passwordStrings []string
				for _, pw := range passwords {
					if pwStr, ok := pw.(string); ok {
						passwordStrings = append(passwordStrings, pwStr)
					}
				}
				go a.StartHashing(passwordStrings)
			} else {
				log.Println("Invalid passwords data received")
			}
		}
	})

	a.ctx = ctx
	a.connectDB()
}

func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}
