package main

import (
	"flag"
	"log"
	"surface-api/models"

	_ "surface-api/docs"

	"github.com/astaxie/beego/session"
	_ "github.com/astaxie/beego/session/mysql"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// @title Surface API
// @version 1.0
// @description API for managing hockey rink schedules, surfaces, locations, and events
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@example.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8000
// @BasePath /
// @schemes http https

// @securityDefinitions.apikey CookieAuth
// @in cookie
// @name gosession

var db *gorm.DB
var cfg models.Config
var sess *session.Manager
var app *App
var configFile = flag.String("config", "config.yaml", "Path to config file")

func init() {
	flag.Parse()
	gin.SetMode(gin.ReleaseMode)

	viper.SetConfigFile(*configFile)
	viper.SetDefault("port", "8000")
	viper.SetDefault("mode", "production")

	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}

	err := viper.Unmarshal(&cfg)

	if err != nil {
		panic(err)
	}

	db, err = gorm.Open(mysql.Open(cfg.DB_DSN))
	if err != nil {
		panic(err)
	}
	log.Println(cfg)

	db.Exec(`SET SESSION sql_mode=(SELECT REPLACE(@@sql_mode,'ONLY_FULL_GROUP_BY',''))`)

	sess, err = session.NewManager("mysql", &session.ManagerConfig{
		CookieName:      "gosession",
		Gclifetime:      3600,
		ProviderConfig:  cfg.DB_DSN,
		EnableSetCookie: true,
	})
	if err != nil {
		log.Fatal(err)
	}
	go sess.GC()
	app = NewApp(db, cfg, sess)
}

func main() {
	r := gin.Default()

	if cfg.Mode == "local" {
		corsCfg := cors.DefaultConfig()
		corsCfg.AllowCredentials = true
		corsCfg.AllowOrigins = []string{"http://localhost:5173"}
		r.Use(cors.New(corsCfg))
	}
	r.Use(app.AuthMiddleware())

	r.GET("/site-locations/:site", app.getSiteLoc)
	r.GET("/mhr-locations", app.getMHRLoc)
	r.POST("/mhr-set-location", app.setMHRLoc)
	r.POST("/mhr-set-surface", app.setMHRSurface)
	r.POST("/mhr-unset-mapping", app.unsetMHRMapping)
	r.GET("/mappings/:site", app.getMappings)
	r.GET("/sites", app.getSites)
	r.GET("/surfaces", app.getSurfaces)
	r.GET("/provinces", app.getProvinces)
	r.POST("/set-surface", app.setSurface)
	r.POST("/set-location", app.setLocation)
	r.POST("/set-mapping", app.setMapping)
	r.POST("/unset-mapping", app.unsetMapping)
	r.POST("/login", app.login)
	r.GET("/logout", app.logout)
	r.GET("/session", app.checkSession)
	r.GET("/report", app.surfaceReport)
	r.GET("/report/download", app.downloadReportCSV)
	r.GET("/rink-report", app.rinkReport)
	r.GET("/events-by-date", app.getEventsByDateRange)
	r.GET("/ramp-mappings/:province", app.rampMappings)
	r.GET("/ramp-provinces", app.rampProvinces)
	r.POST("/set-ramp-mapping", app.SetRampMappings)
	r.GET("/locations", app.getLocations)
	r.GET("/events", app.getEvents)
	r.PUT("/events/:id", app.updateEvent)

	// User management routes
	r.GET("/users", app.listUsers)
	r.POST("/users", app.addUser)
	r.DELETE("/users/:username", app.deleteUser)
	r.PUT("/users/:username/password", app.changePassword)

	// Sites config CRUD routes
	r.GET("/sites-config", app.getSitesConfig)
	r.GET("/parser-types", app.getParserTypes)
	r.GET("/sites-config/:id", app.getSitesConfigByID)
	r.POST("/sites-config", app.createSitesConfig)
	r.PUT("/sites-config/:id", app.updateSitesConfig)
	r.DELETE("/sites-config/:id", app.deleteSitesConfig)

	// Swagger documentation endpoint
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	log.Println("starting server on ", cfg.Port)

	if err := r.Run(":" + cfg.Port); err != nil {
		panic(err)
	}
}
