package winlosetie

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "winlosetie",
	Short: "Send schedules to winlosetie",
	RunE: func(c *cobra.Command, args []string) error {
		var r io.ReadCloser
		var err error

		if *infile == "" {
			r = os.Stdin
		} else {
			r, err = os.Open(*infile)
			if err != nil {
				return err
			}
		}
		defer r.Close()

		return run(r)
	},
}

type Client interface {
	Do(r *http.Request) (*http.Response, error)
}

var (
	infile *string
	client Client
	apiUrl = "http://admin.winloseortie.com/api/v1/vendor/game"
)

func InitCmd() {
	infile = Cmd.Flags().StringP("file", "f", "", "XLS file path (required)")
	client = &http.Client{}
}

func run(r io.Reader) error {
	var ids = make([]string, 0)

	csvr := csv.NewReader(r)

	for {
		rec, err := csvr.Read()
		// id, date, time, surface id
		// 793537,2023-11-11,14:10,3876
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return err
		}

		url := fmt.Sprintf("https://livebarn.com/en/video/%s/%s/%s", rec[3], rec[1], rec[2])
		s := fmt.Sprintf(`{"vendorId":75,"gameIdInternal":%s,"linkStreamVideo":"%s","videoStatus":true}`, rec[0], url)

		ids = append(ids, s)
	}

	content := "[" + strings.Join(ids, ",") + "]"
	req, err := http.NewRequest("POST", apiUrl, strings.NewReader(content))
	if err != nil {
		return err
	}
	req.Header.Add("Content-type", "application/json")
	resp, err := client.Do(req)

	if err != nil {
		return fmt.Errorf("servier returned error %w", err)
	}

	fmt.Println(resp.StatusCode, resp.Status)
	return nil
}
