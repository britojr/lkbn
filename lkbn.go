package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/britojr/lkbn/learner"
)

// Define subcommand names
const (
	ctLearnConst      = "ctlearn"
	ctLearnDescr      = "run latent k-tree model learning algorithm"
	ctParamLearnConst = "ctparam"
	ctParamLearnDescr = "run parameter learning on a given model structure"
)

// Define Flag variables
var (
	// common
	verbose bool // verbose mode

	// learn command
	dataFile      string // dataset csv file
	modelFile     string // network output file
	parmFile      string // parameters file for search algorithms
	learnerAlg    string // learner strategy
	timeAvailable int    // time available to search solution
	numSolutions  int    // number of iterations

	// param learn command
	modelFIn  string // network input file
	modelFOut string // network output file

	// Define subcommands
	ctLearnComm      *flag.FlagSet
	ctParamLearnComm *flag.FlagSet
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
	case ctParamLearnConst:
		ctParamLearnComm.Parse(os.Args[2:])
		runCTParamLearnComm()
	default:
		printDefaults()
		os.Exit(1)
	}
}

func initSubcommands() {
	// Subcommands
	ctLearnComm = flag.NewFlagSet(ctLearnConst, flag.ExitOnError)
	ctParamLearnComm = flag.NewFlagSet(ctParamLearnConst, flag.ExitOnError)

	// learn subcommand flags
	ctLearnComm.BoolVar(&verbose, "v", true, "prints detailed steps")
	ctLearnComm.StringVar(&dataFile, "d", "", "dataset file in csv format")
	ctLearnComm.StringVar(&parmFile, "p", "", "parameters file")
	ctLearnComm.StringVar(&modelFile, "b", "", "network output file")
	ctLearnComm.StringVar(&learnerAlg, "a", learner.AlgCTSampleSearch, "learner algorithm")
	ctLearnComm.IntVar(&timeAvailable, "t", 60, "available time to search solution (0->unbounded)")
	ctLearnComm.IntVar(&numSolutions, "i", 1, "max number of iterations (0->unbounded)")

	// param learn subcommand flags
	ctParamLearnComm.BoolVar(&verbose, "v", true, "prints detailed steps")
	ctParamLearnComm.StringVar(&dataFile, "d", "", "dataset file in csv format")
	ctParamLearnComm.StringVar(&parmFile, "p", "", "parameters file")
	ctParamLearnComm.StringVar(&modelFIn, "bi", "", "network input file")
	ctParamLearnComm.StringVar(&modelFOut, "bo", "", "network output file")
}

func printDefaults() {
	fmt.Printf("lkbn is a tool for learning latent k-tree models\n")
	fmt.Printf("Usage:\n\n")
	fmt.Printf("\tlkbn <command> [options]\n\n")
	fmt.Printf("Commands:\n\n")
	fmt.Printf("\t%v\t\t%v\n", ctLearnConst, ctLearnDescr)
	fmt.Printf("\t%v\t\t%v\n", ctParamLearnConst, ctParamLearnDescr)
	fmt.Println()
	fmt.Printf("For usage details of each command, run:\n\n")
	fmt.Printf("\tlkbn <command> --help\n")
	fmt.Println()
}
