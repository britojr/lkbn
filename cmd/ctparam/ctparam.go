package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/britojr/lkbn/data"
	"github.com/britojr/lkbn/emlearner"
	"github.com/britojr/lkbn/model"
	"github.com/britojr/utl/ioutl"
)

func main() {
	var dataFile, parmFile string
	flag.StringVar(&dataFile, "d", "", "dataset file in csv format")
	flag.StringVar(&parmFile, "p", "", "parameters file")

	flag.Parse()
	if len(dataFile) == 0 {
		fmt.Printf("\n error: missing dataset file\n\n")
		flag.PrintDefaults()
		os.Exit(1)
	}
	ds := data.NewDataset(dataFile)

	var props map[string]string
	if len(parmFile) > 0 {
		props = ioutl.ReadYaml(parmFile)
	}

	dname := strings.TrimSuffix(dataFile, path.Ext(dataFile))
	ctreefs, _ := filepath.Glob(fmt.Sprintf("%v*.ctree", dname))

	eml := emlearner.New()
	eml.SetProperties(props)
	eml.PrintProperties()
	for _, ctFile := range ctreefs {
		ct := model.ReadCTree(ctFile)
		start := time.Now()
		ct, ll, it := eml.Run(ct, ds.IntMaps())
		elapsed := time.Since(start)

		log.Printf("========== SOLUTION ==============================\n")
		log.Printf("Structure: %v\n", ctFile)
		log.Printf("Time: %v\n", elapsed)
		log.Printf("Iterations: %v\n", it)
		log.Printf("LogLikelihood: %.6f\n", ll)
		log.Printf("--------------------------------------------------\n")

		ct.Write(ctFile)
	}
}