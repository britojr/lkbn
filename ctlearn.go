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
	"github.com/britojr/lkbn/scores"
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
	runCTLearner()
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
	runCTParamLearner()
}

func runCTLearner() {
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
	log.Println("Reading dataset file")
	dataSet := data.NewDataset(dataFile)

	log.Println("Initializing learning algorithm")
	alg := learner.Create(learnerAlg)
	alg.SetDataset(dataSet)
	alg.SetFileParameters(props)
	alg.ValidateParameters()
	alg.PrintParameters()

	log.Println("Searching structure")
	start := time.Now()
	m, it := learner.Search(alg, numSolutions, timeAvailable)
	elapsed := time.Since(start)

	log.Printf("========== SOLUTION ==============================\n")
	if m == nil {
		log.Printf("Couldn't find any solution in the given time!\n")
		os.Exit(0)
	}
	log.Printf("Time: %v\n", elapsed)
	log.Printf("Iterations: %v\n", it)
	log.Printf("LogLikelihood: %.6f\n", m.Score())
	log.Printf("--------------------------------------------------\n")

	if len(modelFile) > 0 {
		writeSolution(modelFile, m.(*model.CTree), alg)
	}
}

func runCTParamLearner() {
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
	log.Println("Reading dataset file")
	dataSet := data.NewDataset(dataFile)

	log.Println("Reading model structure")
	ct := model.ReadCTree(modelFIn)
	log.Println("Initializong parameter learner")
	eml := emlearner.New()
	eml.SetProperties(props)
	eml.PrintProperties()
	log.Println("Learning parameters")
	start := time.Now()
	m, ll, it := eml.Run(ct, dataSet.IntMaps())
	elapsed := time.Since(start)
	m.SetBIC(scores.ComputeBIC(m, dataSet))

	log.Printf("========== SOLUTION ==============================\n")
	log.Printf("Time: %v\n", elapsed)
	log.Printf("Iterations: %v\n", it)
	log.Printf("LogLikelihood: %.6f\n", ll)
	log.Printf("BIC: %.6f\n", m.BIC())
	log.Printf("--------------------------------------------------\n")

	if len(modelFOut) > 0 {
		writeSolution(modelFOut, m, nil)
	}
}

func writeSolution(fname string, m *model.CTree, alg learner.Learner) {
	log.Printf("Printing solution: '%v'\n", fname)
	m.Write(fname)
}
