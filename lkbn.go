package main

import (
	"flag"
	"fmt"
	"os"
)

// Define subcommand names
const (
	ctLearnConst = "ctlearn"
	ctLearnDescr = "run latent k-tree model learning algorithm"
)

// Define Flag variables
var (
	// common
	verbose bool // verbose mode

	// struct command
	dataFile      string // dataset csv file
	modelFile     string // network output file
	parmFile      string // parameters file for search algorithms
	learnerAlg    string // learner strategy
	timeAvailable int    // time available to search solution
	numSolutions  int    // number of iterations

	// Define subcommands
	ctLearnComm *flag.FlagSet
)

func main() {
	initSubcommands()
	// Verify that a subcommand has been provided
	// os.Arg[0] : main command, os.Arg[1] : subcommand
	if len(os.Args) < 2 {
		printDefaults()
		os.Exit(1)
	}
	switch os.Args[1] {
	case ctLearnConst:
		ctLearnComm.Parse(os.Args[2:])
		runCTLearnComm()
	default:
		printDefaults()
		os.Exit(1)
	}
}

func initSubcommands() {
	// Subcommands
	ctLearnComm = flag.NewFlagSet(ctLearnConst, flag.ExitOnError)

	// struct subcommand flags
	ctLearnComm.BoolVar(&verbose, "v", true, "prints detailed steps")
	ctLearnComm.StringVar(&dataFile, "d", "", "dataset file in csv format")
	ctLearnComm.StringVar(&parmFile, "p", "", "parameters file")
	ctLearnComm.StringVar(&modelFile, "b", "", "network output file")
	ctLearnComm.StringVar(&learnerAlg, "a", "sample", "learner algorithm")
	ctLearnComm.IntVar(&timeAvailable, "t", 60, "available time to search solution (0->unbounded)")
	ctLearnComm.IntVar(&numSolutions, "i", 1, "max number of iterations (0->unbounded)")
}

func printDefaults() {
	fmt.Printf("lkbn is a tool for learning latent k-tree models\n")
	fmt.Printf("Usage:\n\n")
	fmt.Printf("\tlkbn <command> [options]\n\n")
	fmt.Printf("Commands:\n\n")
	fmt.Printf("\t%v\t\t%v\n", ctLearnConst, ctLearnDescr)
	fmt.Println()
	fmt.Printf("For usage details of each command, run:\n\n")
	fmt.Printf("\tlkbn <command> --help\n")
	fmt.Println()
}
