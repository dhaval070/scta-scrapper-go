package main

import (
	"encoding/csv"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path"
	"regexp"
	"slices"

	"github.com/spf13/cobra"
)

var cmd = &cobra.Command{
	Use:   "unique",
	Short: "Report unique games full and livebarn",
	Long:  "This command generates list of unique full and livebarn matched games of given league between from start and end date",
	RunE: func(c *cobra.Command, args []string) error {
		root, err := c.Flags().GetString("root")
		if err != nil {
			return err
		}
		return report(root)
	},
}

var (
	dtFrom *string
	dtTo   *string
	league *string
)

func init() {
	dtFrom = cmd.Flags().String("from", "", "from date(yyyymmdd)")
	dtTo = cmd.Flags().String("to", "", "to date(yyyymmdd)")
	league = cmd.Flags().String("league", "", "league")
	cmd.Flags().String("root", "./var", "root directory")

	cmd.MarkFlagRequired("from")
	cmd.MarkFlagRequired("to")
	cmd.MarkFlagRequired("league")
}

func main() {
	cmd.Execute()
}

func report(root string) error {
	entries, err := os.ReadDir(root)

	if err != nil {
		return err
	}

	dirs := filter(entries)

	result := [][]string{}
	resultWithLivebarn := [][]string{}

	var prevLast string
	var prevLastWithLivebarn string

	for _, d := range dirs {
		filepath := path.Join(root, d, *league+".csv")
		rows, err := readFile(prevLast, filepath)

		if err != nil {
			return fmt.Errorf("error in path %s: %w ", filepath, err)
		}

		if len(rows) > 0 {
			prevLast = rows[len(rows)-1][0]
			result = append(result, rows...)
		}

		// for livebarn matched files
		filepath = path.Join(root, d, "with-surface", *league+".csv")
		rows, err = readFile(prevLastWithLivebarn, filepath)

		if err != nil {
			return fmt.Errorf("error in path %s: %w ", filepath, err)
		}

		if len(rows) > 0 {
			prevLastWithLivebarn = rows[len(rows)-1][0]
			resultWithLivebarn = append(resultWithLivebarn, rows...)
		}
	}

	err = writeFiles(root, *league+"-full-unique.csv", result)
	if err != nil {
		return fmt.Errorf("error writing file %w", err)
	}

	err = writeFiles(root, *league+"-livebarn-unique.csv", resultWithLivebarn)
	if err != nil {
		return fmt.Errorf("error writing file %w", err)
	}
	return nil
}

func writeFiles(root, filename string, result [][]string) error {
	filename = path.Join(root, filename)
	fh, err := os.Create(filename)
	if err != nil {
		return err
	}

	writer := csv.NewWriter(fh)
	if err := writer.WriteAll(result); err != nil {
		return err
	}
	writer.Flush()
	fh.Close()
	return nil
}

func readFile(prevLast string, path string) ([][]string, error) {
	fh, err := os.Open(path)

	if err != nil {
		log.Fatal(err)
	}
	defer fh.Close()

	reader := csv.NewReader(fh)
	rows, err := reader.ReadAll()

	if err != nil {
		return nil, fmt.Errorf("failed to readall %w", err)
	}

	if len(rows) == 0 {
		log.Println("no records in " + path)
		return nil, nil
	}

	if prevLast == "" {
		return rows, nil
	}

	var result [][]string
	for _, row := range rows {
		// prevent duplicates
		if row[0] <= prevLast {
			continue
		}
		result = append(result, row)
	}
	return result, nil
}

func filter(entries []fs.DirEntry) []string {
	filtered := []string{}

	re := regexp.MustCompile("^[0-9]{8}$")

	var months = map[string]string{}

	for _, e := range entries {
		name := e.Name()

		if !e.Type().IsDir() || !re.MatchString(name) {
			continue
		}

		if name < *dtFrom || name > *dtTo {
			continue
		}

		yyyymm := name[0:6]
		months[yyyymm] = name[6:]
	}

	for yyymm, dd := range months {
		filtered = append(filtered, yyymm+dd)
	}

	slices.Sort(filtered)

	fmt.Printf("final dates: %+v\n", filtered)
	return filtered
}
