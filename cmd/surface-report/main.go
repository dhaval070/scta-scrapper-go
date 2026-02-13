package main

import (
	"calendar-scrapper/config"
	"calendar-scrapper/pkg/repository"
	"encoding/csv"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"time"

	"github.com/spf13/cobra"
	mail "github.com/xhit/go-simple-mail/v2"
)

var (
	date  string
	email bool
)
var fileDir = "var/surface-report"

var rootCmd = &cobra.Command{
	Use:   "surface-report",
	Short: "Generates surface usage report",
	Long:  `A CLI tool to generate surface usage report`,
	Run:   runReport,
}

var cfg config.Config
var repo *repository.Repository

func init() {
	rootCmd.Flags().StringVar(&date, "date", "", "Date of surface usage. (YYYY-MM-DD)")
	rootCmd.Flags().BoolVar(&email, "email", false, "send report via email")
	config.Init("config", ".")

	cfg = config.MustReadConfig()
	repo = repository.NewRepository(cfg)

	info, err := os.Stat(fileDir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			if err = os.MkdirAll(fileDir, fs.ModePerm); err != nil {
				log.Fatal(err)
			}
		} else {
			log.Fatal(err)
		}
	} else if !info.IsDir() {
		log.Fatal(fileDir + " is not a directory")
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

type SurfaceReport struct {
	SurfaceID    string `json:"surface_id"`
	LocationID   string `json:"location_id"`
	LocationName string `json:"location_name"`
	SurfaceName  string `json:"surface_name"`
	DayOfWeek    string `json:"day_of_week"`
	StartTime    string `json:"start_time"`
}

func runReport(cmd *cobra.Command, args []string) {
	var fromDt time.Time
	var toDt time.Time
	var err error

	if date == "" {
		fromDt = time.Now()
		toDt = time.Now()
		toDt = toDt.Add(24 * time.Hour)
	} else {
		fromDt, err = time.Parse("2006-01-02", date)
		if err != nil {
			log.Fatal("invalid date format")
		}
		toDt = fromDt.Add(24 * time.Hour)
	}

	whereClause := " WHERE e.datetime between ? and ?"
	dates := []any{fromDt.Format("2006-01-02"), toDt.Format("2006-01-02")}

	query := `SELECT
		e.surface_id,
		any_value(s.location_id),
		any_value(l.name) location_name,
		any_value(s.name) surface_name,
		any_value(date_format(e.datetime, "%W")) day_of_week,
		date_format(min(e.datetime), "%Y-%m-%d %T") start_time
	FROM
		events e JOIN surfaces s on e.surface_id=s.id JOIN locations l on l.id=s.location_id` +
		whereClause +
		`
	GROUP BY e.surface_id, date(e.datetime)
	ORDER BY location_name, surface_name, surface_id, day_of_week, start_time`

	var result []SurfaceReport
	if err := repo.DB.Raw(query, dates...).Scan(&result).Error; err != nil {
		log.Println(err)
		return
	}
	now := time.Now()

	filePath := fileDir + "/" + now.Format("2006-01-02") + "-surface-report.csv"
	fh, err := os.Create(filePath)
	if err != nil {
		log.Fatal("failed to create file", err)
	}
	// var b = &bytes.Buffer{}
	w := csv.NewWriter(fh)

	w.Write([]string{
		"Surface ID", "Location Name", "Surface Name", "Day of Week", "Start Time", "Pre-Check", "Post-Check", "Sound", "VOD Playback", "Notes",
	})

	for _, row := range result {
		err = w.Write([]string{
			row.SurfaceID,
			row.LocationName,
			row.SurfaceName,
			row.DayOfWeek,
			row.StartTime,
			"", "", "", "", "",
		})
		if err != nil {
			log.Fatal("csv write error ", err)
		}
	}
	w.Flush()
	if err = fh.Close(); err != nil {
		log.Fatal("error closing file", err)
	}

	log.Println("report generated")

	if email {
		if len(cfg.SurfaceReport.MailTo) == 0 {
			log.Println("no receipient configured for surface report")
			return
		}
		if err = sendEmail(filePath); err != nil {
			log.Fatal(err)
		}
		log.Println("report sent")
	}
}

func sendEmail(filePath string) error {
	server := mail.NewSMTPClient()
	server.Host = cfg.SmtpConfig.Host
	server.Port = cfg.SmtpConfig.Port
	server.Encryption = mail.EncryptionSSLTLS
	server.Authentication = mail.AuthNone

	conn, err := server.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to smtp server %w", err)
	}

	now := time.Now()

	subject := "Surface Checklist for date - " + now.Format("2006-01-02")
	email := mail.NewMSG().SetFrom("screduler@gameonstream.com").AddTo(cfg.SurfaceReport.MailTo...).SetSubject(subject)

	email.Attach(&mail.File{
		FilePath: filePath,
	})

	email.SetBody(mail.TextPlain, "surface report")
	if email.Error != nil {
		return fmt.Errorf("email error %w", email.Error)
	}

	err = email.Send(conn)
	if err != nil {
		return fmt.Errorf("email send error %w", err)
	}
	return nil
}
