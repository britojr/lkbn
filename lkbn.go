package main

import (
	"flag"
	"fmt"
	"os"
)

// Define subcommand names
const (
	structConst = "struct"
	structDescr = "run latent k-tree Bayesian networks learning algorithm"
)

// Define Flag variables
var (
	// common
	verbose bool // verbose mode

	// struct command
	dataFile      string // dataset csv file
	bnetFile      string // network output file
	parmFile      string // parameters file for search algorithms
	learnerAlg    string // structure learner algorithm
	timeAvailable int    // time available to search solution
	numSolutions  int    // number of iterations

	// Define subcommands
	structComm *flag.FlagSet
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
	case structConst:
		structComm.Parse(os.Args[2:])
		runStructComm()
	default:
		printDefaults()
		os.Exit(1)
	}
}

func initSubcommands() {
	// Subcommands
	structComm = flag.NewFlagSet(structConst, flag.ExitOnError)

	// struct subcommand flags
	structComm.BoolVar(&verbose, "v", true, "prints detailed steps")
	structComm.StringVar(&dataFile, "d", "", "dataset file in csv format")
	structComm.StringVar(&parmFile, "p", "", "parameters file")
	structComm.StringVar(&bnetFile, "b", "", "network output file")
	structComm.StringVar(&learnerAlg, "a", "sample", "structure learner algorith")
	structComm.IntVar(&timeAvailable, "t", 60, "available time to search solution (0->unbounded)")
	structComm.IntVar(&numSolutions, "i", 1, "max number of iterations (0->unbounded)")
}

func printDefaults() {
	fmt.Printf("lkbn is a tool for learning latent k-tree Bayesian networks\n")
	fmt.Printf("Usage:\n\n")
	fmt.Printf("\tlkbn <command> [options]\n\n")
	fmt.Printf("Commands:\n\n")
	fmt.Printf("\t%v\t\t%v\n", structConst, structDescr)
	fmt.Println()
	fmt.Printf("For usage details of each command, run:\n\n")
	fmt.Printf("\tlkbn <command> --help\n")
	fmt.Println()
}
