package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/britojr/lkbn/data"
	"github.com/britojr/lkbn/model"
	"github.com/britojr/lkbn/scores"
)

// computes the scores of a given structure
func main() {
	var dataFile, modelFIn string
	flag.StringVar(&dataFile, "d", "", "dataset file in csv format")
	flag.StringVar(&modelFIn, "b", "", "network input file")

	flag.Parse()
	if len(dataFile) == 0 {
		fmt.Printf("\n error: missing dataset file\n\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	log.Printf("=========== COMPUTE MODEL SCORES =================\n")
	log.Printf("Dataset file: '%v'\n", dataFile)
	log.Printf("Read structure from: '%v'\n", modelFIn)
	log.Printf("--------------------------------------------------\n")

	log.Println("Reading dataset file")
	dataSet := data.NewDataset(dataFile)

	log.Println("Reading model structure")
	ct := model.ReadCTree(modelFIn)
	// check variable ordering
	for i, v := range dataSet.Variables() {
		if ct.Variables()[i].Name() != v.Name() {
			log.Printf("error: wrong variable ordering:\n%v\n%v\n", dataSet.Variables(), ct.Variables())
			os.Exit(1)
		}
	}

	log.Println("Computing scores")
	start := time.Now()
	ct.SetScore(scores.ComputeLL(ct, dataSet.IntMaps()))
	ct.SetBIC(scores.ComputeBIC(ct, dataSet.IntMaps()))
	elapsed := time.Since(start)

	log.Printf("========== RESULT ================================\n")
	log.Printf("Time: %v\n", elapsed)
	log.Printf("LogLikelihood: %.6f\n", ct.Score())
	log.Printf("BIC: %.6f\n", ct.BIC())
	log.Printf("--------------------------------------------------\n")
}
