package main

import (
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"
)

const (
	defaultFilename      = "questions.csv"
	defaultTimeLimit     = 30
	defaultShouldShuffle = true
)

var (
	filename      string
	timeLimit     int
	shouldShuffle bool
)

type Problem struct {
	question string
	answer   string
}

func main() {
	rand.Seed(time.Now().UnixNano())

	if err := handleFlags(); err != nil {
		log.Fatalln(err)
	}
	problems, err := parseFile()
	if err != nil {
		log.Fatalln(err)
	}
	if shouldShuffle {
		shuffleProblems(problems)
	}

	startQuiz(problems)
}

func handleFlags() error {
	flag.StringVar(&filename, "f", defaultFilename,
		"csv filename in the format of 'question,answer'")
	flag.IntVar(&timeLimit, "l", defaultTimeLimit, "time limit in seconds")
	flag.BoolVar(&shouldShuffle, "s", defaultShouldShuffle, "shuffle problems")
	flag.Parse()

	if !strings.HasSuffix(filename, ".csv") {
		return errors.New("filename must have .csv extension")
	}

	return nil
}

func parseFile() ([]Problem, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	r := csv.NewReader(f)
	var problems []Problem
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if len(record) != 2 {
			return nil, errors.New("each line must have two fields")
		}

		problems = append(problems, Problem{
			question: record[0],
			answer:   normalizeString(record[1]),
		})
	}

	return problems, nil
}

func shuffleProblems(problems []Problem) {
	rand.Shuffle(len(problems), func(i, j int) {
		problems[i], problems[j] = problems[j], problems[i]
	})
}

func startQuiz(problems []Problem) {
	fmt.Println("Welcome to the quiz!")
	fmt.Printf("You will be presented with %d questions.\n", len(problems))
	fmt.Printf("Try to answer as much as possible in %d seconds\n", timeLimit)

	var (
		answer string
		score  int
	)

	timer := time.NewTimer(time.Duration(timeLimit) * time.Second)
	answerChan := make(chan string)

problems_loop:
	for i, p := range problems {
		fmt.Printf("Problem #%d: %s = ", i+1, p.question)

		go func() {
			scanNormalized(&answer)
			answerChan <- answer
		}()

		select {
		case <-timer.C:
			fmt.Println("\nYou ran out of time!")
			break problems_loop
		case <-answerChan:
			if answer == p.answer {
				score++
			}
		}
	}

	fmt.Printf("You scored %d out of %d.\n", score, len(problems))
}

func scanNormalized(s *string) {
	fmt.Scanln(s)
	*s = normalizeString(*s)
}

func normalizeString(s string) string {
	s = strings.ToLower(s)
	return strings.TrimSpace(s)
}
