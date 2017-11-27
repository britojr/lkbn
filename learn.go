package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/britojr/lkbn/data"
	"github.com/britojr/lkbn/emlearner"
	"github.com/britojr/lkbn/learner"
	"github.com/britojr/lkbn/model"
	"github.com/britojr/utl/ioutl"
)

func runCTLearnComm() {
	// Required Flags
	if dataFile == "" {
		fmt.Printf("\n error: missing dataset file\n\n")
		ctLearnComm.PrintDefaults()
		os.Exit(1)
	}
	if !verbose {
		log.SetOutput(ioutil.Discard)
	}

	runLearner()
}

func runLearner() {
	log.Printf("=========== BEGIN MODEL LEARNING =================\n")
	log.Printf("Dataset file: '%v'\n", dataFile)
	log.Printf("Learning algorithm: '%v'\n", learnerAlg)
	log.Printf("Max. iterations: %v\n", numSolutions)
	log.Printf("Max. time available (sec): %v\n", timeAvailable)
	log.Printf("Parameters file: '%v'\n", parmFile)
	log.Printf("Save solution in: '%v'\n", modelFile)
	log.Printf("--------------------------------------------------\n")

	var props map[string]string
	if len(parmFile) > 0 {
		log.Println("Reading parameters file")
		props = ioutl.ReadYaml(parmFile)
	}
	dataSet := data.NewDataset(dataFile)

	log.Println("Initializong learning algorithm")
	alg := learner.Create(learnerAlg)
	alg.SetDataset(dataSet)
	alg.SetFileParameters(props)
	alg.ValidateParameters()
	alg.PrintParameters()

	log.Println("Searching structure")
	start := time.Now()
	m := learner.Search(alg, numSolutions, timeAvailable)
	elapsed := time.Since(start)

	log.Printf("========== SOLUTION ==============================\n")
	if m == nil {
		log.Printf("Couldn't find any solution in the given time!\n")
		os.Exit(0)
	}
	log.Printf("Time: %v\n", elapsed)
	log.Printf("Best Score: %.6f\n", m.Score())
	log.Printf("--------------------------------------------------\n")

	if len(modelFile) > 0 {
		writeSolution(modelFile, m.(model.Model), alg)
	}
}

func writeSolution(fname string, m model.Model, alg learner.Learner) {
	log.Printf("Printing solution: '%v'\n", fname)
	f := ioutl.CreateFile(fname)
	defer f.Close()
}

func runCTParamLearnComm() {
	// Required Flags
	if dataFile == "" {
		fmt.Printf("\n error: missing dataset file\n\n")
		ctParamLearnComm.PrintDefaults()
		os.Exit(1)
	}
	if modelFIn == "" {
		fmt.Printf("\n error: missing model structure file\n\n")
		ctParamLearnComm.PrintDefaults()
		os.Exit(1)
	}
	if !verbose {
		log.SetOutput(ioutil.Discard)
	}

	runParamLearner()
}

func runParamLearner() {
	log.Printf("=========== BEGIN MODEL LEARNING =================\n")
	log.Printf("Dataset file: '%v'\n", dataFile)
	log.Printf("Parameters file: '%v'\n", parmFile)
	log.Printf("Read structure from: '%v'\n", modelFIn)
	log.Printf("Save solution in: '%v'\n", modelFOut)
	log.Printf("--------------------------------------------------\n")

	var props map[string]string
	if len(parmFile) > 0 {
		log.Println("Reading parameters file")
		props = ioutl.ReadYaml(parmFile)
	}
	dataSet := data.NewDataset(dataFile)

	log.Println("Reading model structure")
	ct := model.Read(modelFIn)
	log.Println("Initializong parameter learner")
	eml := emlearner.New()
	eml.SetProperties(props)
	log.Println("Learning parameters")
	start := time.Now()
	m, ll := eml.Run(ct, dataSet.IntMaps())
	elapsed := time.Since(start)

	log.Printf("========== SOLUTION ==============================\n")
	log.Printf("Time: %v\n", elapsed)
	log.Printf("LogLikelihood: %.6f\n", ll)
	log.Printf("--------------------------------------------------\n")

	if len(modelFOut) > 0 {
		writeSolution(modelFOut, m, nil)
	}
}
