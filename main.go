package main

import (
	"flag"
	"fmt"
	"github.com/jrabasco/quiz-grader/files"
	"github.com/jrabasco/quiz-grader/grader"
	"os"
)

const (
	NoFlagProvided  int = 1
	CannotParse     int = 3
	NotADir         int = 5
	GradingError    int = 7
	PrintScoreError int = 9
)

func main() {
	answersPath := flag.String("answers", "", "path to answers file")
	submissionsPath := flag.String("submissions", "", "path to submissions")
	section := flag.Int("section", 0, "section to grade")
	flag.Parse()

	if *answersPath == "" {
		fmt.Fprintln(os.Stderr, "-answers must be specified")
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(NoFlagProvided)
	}

	if *submissionsPath == "" {
		fmt.Fprintln(os.Stderr, "-submissions must be specified")
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(NoFlagProvided)
	}

	if *section == 0 {
		fmt.Fprintln(os.Stderr, "-section must be specified")
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(NoFlagProvided)
	}

	if !files.IsDir(*submissionsPath) {
		fmt.Fprintf(os.Stderr, "%s is not a directory\n", *submissionsPath)
		os.Exit(NotADir)
	}

	g, err := grader.FromFile(*answersPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, fmt.Sprintf("Could not parse answers: %v\n", err))
		os.Exit(CannotParse)
	}

	err = g.Grade(*submissionsPath, *section)
	if err != nil {
		fmt.Fprintf(os.Stderr, fmt.Sprintf("Error while grading: %v\n", err))
		os.Exit(GradingError)
	}

	err = g.PrintScores(*submissionsPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, fmt.Sprintf("Error while printing scores: %v\n", err))
		os.Exit(GradingError)
	}
}
