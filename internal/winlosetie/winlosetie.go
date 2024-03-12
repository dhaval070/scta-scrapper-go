package winlosetie

import (
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

		r, err = os.Open(*infile)
		if err != nil {
			return err
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
	site   *string
	client Client
	apiUrl = "http://admin.winloseortie.com/api/v1/vendor/game"
)

func InitCmd() {
	infile = Cmd.Flags().StringP("file", "f", "", "input file path (required)")
	site = Cmd.Flags().StringP("site", "s", "", "site name(e.g. nyhl)")
	client = &http.Client{}
	Cmd.MarkFlagRequired("file")
}

func run(r io.Reader) error {
	var ids = make([]string, 0)
	var records []DataRec
	var err error

	records, err = formatInput(r, *site)

	for _, rec := range records {
		url := fmt.Sprintf("https://livebarn.com/en/video/%s/%s/%s", rec.SurfaceID, rec.Date, rec.Time)
		s := fmt.Sprintf(`{"vendorId":75,"gameIdInternal":%d,"linkStreamVideo":"%s","videoStatus":true}`, rec.Id, url)

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
