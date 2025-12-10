package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"calendar-scrapper/config"

	"github.com/golang-migrate/migrate/v4"
	mysqlMigrate "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/spf13/cobra"
	mysqlDriver "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var rootCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Database migration tool",
	Long:  `Run database migrations up or down with optional level control.`,
}

var upCmd = &cobra.Command{
	Use:   "up [level]",
	Short: "Run migrations up",
	Long:  `Apply pending migrations. Optionally specify number of migrations to apply.`,
	Args:  cobra.MaximumNArgs(1),
	Run:   runUp,
}

var downCmd = &cobra.Command{
	Use:   "down <level>",
	Short: "Run migrations down",
	Long:  `Rollback migrations. Specify number of migrations to rollback.`,
	Args:  cobra.ExactArgs(1),
	Run:   runDown,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show current migration version",
	Long:  `Display the current database migration version.`,
	Run:   runVersion,
}

func init() {
	rootCmd.AddCommand(upCmd)
	rootCmd.AddCommand(downCmd)
	rootCmd.AddCommand(versionCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func getMigrate() (*migrate.Migrate, error) {
	config.Init("config", ".")
	cfg := config.MustReadConfig()
	
	// Ensure multiStatements=true is in DSN for multiple SQL statements per migration
	dsn := cfg.DbDSN
	if !strings.Contains(dsn, "multiStatements=true") {
		if strings.Contains(dsn, "?") {
			dsn += "&multiStatements=true"
		} else {
			dsn += "?multiStatements=true"
		}
	}
	
	db, err := gorm.Open(mysqlDriver.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	driver, err := mysqlMigrate.WithInstance(sqlDB, &mysqlMigrate.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to create migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://database/migrations",
		"mysql",
		driver,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create migration instance: %w", err)
	}

	return m, nil
}

func runUp(cmd *cobra.Command, args []string) {
	m, err := getMigrate()
	if err != nil {
		log.Fatalf("Migration setup failed: %v", err)
	}
	defer m.Close()

	if len(args) == 0 {
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Migration up failed: %v", err)
		}
		fmt.Println("Migrations applied successfully")
	} else {
		level, err := strconv.Atoi(args[0])
		if err != nil {
			log.Fatalf("Invalid level: %v", err)
		}
		if err := m.Steps(level); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Migration up failed: %v", err)
		}
		fmt.Printf("Applied %d migration(s)\n", level)
	}
}

func runDown(cmd *cobra.Command, args []string) {
	m, err := getMigrate()
	if err != nil {
		log.Fatalf("Migration setup failed: %v", err)
	}
	defer m.Close()

	level, err := strconv.Atoi(args[0])
	if err != nil {
		log.Fatalf("Invalid level: %v", err)
	}
	if err := m.Steps(-level); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Migration down failed: %v", err)
	}
	fmt.Printf("Rolled back %d migration(s)\n", level)
}

func runVersion(cmd *cobra.Command, args []string) {
	m, err := getMigrate()
	if err != nil {
		log.Fatalf("Migration setup failed: %v", err)
	}
	defer m.Close()

	version, dirty, err := m.Version()
	if err != nil {
		log.Fatalf("Failed to get version: %v", err)
	}

	fmt.Printf("Current version: %d\n", version)
	if dirty {
		fmt.Println("Warning: Database is in dirty state")
	}
}
