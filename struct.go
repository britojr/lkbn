package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/britojr/lkbn/data"
	"github.com/britojr/lkbn/learner"
	"github.com/britojr/lkbn/model"
	"github.com/britojr/utl/ioutl"
)

func runStructComm() {
	// Required Flags
	if dataFile == "" {
		fmt.Printf("\n error: missing dataset file\n\n")
		structComm.PrintDefaults()
		os.Exit(1)
	}
	if !verbose {
		log.SetOutput(ioutil.Discard)
	}

	structureLearning()
}

func structureLearning() {
	log.Printf(" ========== BEGIN STRUCTURE OPTIMIZATION ========== \n")
	log.Printf("Dataset file: '%v'\n", dataFile)
	log.Printf("Learning algorithm: '%v'\n", learnerAlg)
	log.Printf("Max. iterations: %v\n", numSolutions)
	log.Printf("Max. time available (sec): %v\n", timeAvailable)
	log.Printf("Parameters file: '%v'\n", parmFile)
	log.Printf("Save solution in: '%v'\n", bnetFile)
	log.Printf(" -------------------------------------------------- \n")

	log.Println("Reading parameters file")
	parms := ioutl.ReadYaml(parmFile)
	dataSet := data.NewDataset(dataFile)

	log.Println("Creating structure learning algorithm")
	alg := learner.Create(learnerAlg)
	alg.SetDataset(dataSet)
	alg.SetFileParameters(parms)
	alg.PrintParameters()

	log.Println("Searching structure")
	start := time.Now()
	m := learner.Search(alg, numSolutions, timeAvailable).(*model.CTree)
	elapsed := time.Since(start)

	log.Printf(" ========== SOLUTION ============================== \n")
	if m == nil {
		log.Printf("Couldn't find any solution in the given time!\n")
		os.Exit(0)
	}
	totScore := m.ComputeScore(dataSet)
	log.Printf("Time: %v\n", elapsed)
	log.Printf("Best Score: %.6f\n", totScore)
	log.Printf(" -------------------------------------------------- \n")

	if len(bnetFile) > 0 {
		writeSolution(bnetFile, m, alg)
	}
}

func writeSolution(fname string, bn *model.CTree, alg learner.Learner) {
	log.Printf("Printing solution: '%v'\n", fname)
	f := ioutl.CreateFile(fname)
	defer f.Close()
}
