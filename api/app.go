package main

import (
	"os/exec"
	"sync"
	"time"

	"github.com/astaxie/beego/session"
	"gorm.io/gorm"
	"surface-api/models"
)

const (
	pwdChangeWindow      = 5 * time.Minute
	pwdChangeMaxAttempts = 5
)

// App holds application dependencies and state
type App struct {
	db                *gorm.DB
	cfg               models.Config
	sess              *session.Manager
	pwdChangeLock     sync.Mutex
	pwdChangeAttempts map[string][]time.Time
	scrapingMu        sync.Mutex
	scrapingProcesses map[string]*exec.Cmd
}

// NewApp creates a new App instance
func NewApp(db *gorm.DB, cfg models.Config, sess *session.Manager) *App {
	return &App{
		db:                db,
		cfg:               cfg,
		sess:              sess,
		pwdChangeAttempts: make(map[string][]time.Time),
		scrapingProcesses: make(map[string]*exec.Cmd),
	}
}
