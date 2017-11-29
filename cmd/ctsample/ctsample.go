package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/britojr/lkbn/data"
	"github.com/britojr/lkbn/model"
)

func main() {
	var (
		dataFile    string
		nSamples, k int
	)
	flag.StringVar(&dataFile, "d", "", "dataset file in csv format")
	flag.IntVar(&k, "k", 1, "tree-width")
	flag.IntVar(&nSamples, "s", 1, "number of samples")

	flag.Parse()
	if len(dataFile) == 0 {
		fmt.Printf("\n error: missing dataset file\n\n")
		flag.PrintDefaults()
		os.Exit(1)
	}
	dname := strings.TrimSuffix(dataFile, path.Ext(dataFile))

	ds := data.NewDataset(dataFile)
	for i := 0; i < nSamples; i++ {
		modelFOut := fmt.Sprintf("%s#%04d.ctree", dname, i)
		ct := model.SampleUniform(ds.Variables(), k)
		ct.Write(modelFOut)
	}
}
