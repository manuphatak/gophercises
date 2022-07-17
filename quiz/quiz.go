package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/samber/lo"

	"github.com/urfave/cli/v2"
)

func main() {
	rand.Seed(time.Now().UnixNano())

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
			&cli.IntFlag{
				Name:    "limit",
				Value:   30,
				Usage:   "The time limit to complete the quiz (seconds)",
				Aliases: []string{"l"},
			},
			&cli.BoolFlag{
				Name:    "shuffle",
				Value:   false,
				Usage:   "Randomize the order of the questions",
				Aliases: []string{"s"},
			},
		},
		Action: func(c *cli.Context) error {
			csvPath := c.Path("csv")
			limit := c.Int("limit")
			shuffle := c.Bool("shuffle")

			rows, err := readCsv(csvPath)
			if err != nil {
				log.Fatal(err)
			}

			if shuffle {
				rows = lo.Shuffle(rows)
			}

			results, err := collectQuizAnswers(rows, limit)
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

func collectQuizAnswers(rows [][]string, limit int) ([]bool, error) {
	results := make([]bool, len(rows))
	timer := time.NewTimer(time.Duration(limit) * time.Second)

	reader := bufio.NewReader(os.Stdin)
	for i, row := range rows {
		fmt.Printf("Question #%d: %s = ", i+1, row[0])
		answerChan := make(chan string)
		go asyncReadAnswer(reader, answerChan)

		select {
		case <-timer.C:
			fmt.Println()
			return results, nil
		case answer := <-answerChan:
			results[i] = answer == row[1]

		}

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

func asyncReadAnswer(reader *bufio.Reader, answerChan chan string) {
	answer, err := reader.ReadString('\n')

	if err != nil {
		log.Fatal(err)
	}
	answerChan <- strings.TrimSpace(answer)
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
