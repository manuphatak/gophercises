package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/samber/lo"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:                   "quiz",
		Usage:                  "An interactive quiz game",
		Compiled:               time.Now(),
		HideHelpCommand:        true,
		UseShortOptionHandling: true,
		EnableBashCompletion:   true,
		ArgsUsage:              " ",
		Flags: []cli.Flag{
			&cli.PathFlag{
				Name:      "csv",
				Value:     "problems.csv",
				Usage:     "A csv file in the format of 'question,answer'",
				TakesFile: true,
				Aliases:   []string{"c"},
			},
		},
		Action: func(c *cli.Context) error {
			csvPath := c.Path("csv")
			rows, err := readCsv(csvPath)
			if err != nil {
				log.Fatal(err)
			}
			results, err := collectQuizAnswers(rows)
			if err != nil {
				log.Fatal(err)
			}

			printSummary(results)
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func collectQuizAnswers(rows [][]string) ([]bool, error) {
	results := []bool{}

	reader := bufio.NewReader(os.Stdin)
	for i, row := range rows {

		fmt.Printf("Question #%d: %s = ", i, row[0])
		answer, err := reader.ReadString('\n')

		if err != nil {
			return nil, err
		}

		results = append(results, strings.TrimSpace(answer) == row[1])

	}
	return results, nil
}

func readCsv(csvPath string) ([][]string, error) {
	csvFile, err := os.Open(csvPath)
	if err != nil {
		return nil, err
	}
	defer csvFile.Close()

	csvReader := csv.NewReader(csvFile)

	rows, err := csvReader.ReadAll()
	if err != nil {
		return nil, err
	}

	return rows, nil
}

func printSummary(results []bool) {
	correct := lo.Count(results, true)
	count := len(results)

	if correct == count {
		fmt.Print("🎉 ")
	} else if correct == 0 {
		fmt.Print("❌ ")
	}

	fmt.Printf("You scored %d out of %d\n", correct, count)
}
