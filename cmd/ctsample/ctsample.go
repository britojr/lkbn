package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/britojr/btbn/scr"
	"github.com/britojr/lkbn/data"
	"github.com/britojr/lkbn/learner"
	"github.com/britojr/lkbn/model"
	"github.com/britojr/lkbn/vars"
	"github.com/britojr/utl/conv"
	"github.com/britojr/utl/ioutl"
)

// parms file fields
const (
	ParmTreewidth  = "treewidth"
	ParmLatentVars = "latent_vars"
	ParmNumSamples = "num_samples"
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
	mutInfo := scr.ComputeMutInf(dataFile)

	var props map[string]string
	if len(parmFile) > 0 {
		props = ioutl.ReadYaml(parmFile)
	}
	tw := 1
	if twStr, ok := props[ParmTreewidth]; ok {
		tw = conv.Atoi(twStr)
	}
	numSamples := 1
	if numSampStr, ok := props[ParmNumSamples]; ok {
		numSamples = conv.Atoi(numSampStr)
	}
	latentVars := []int{}
	if lvStr, ok := props[ParmLatentVars]; ok {
		latentVars = conv.Satoi(strings.FieldsFunc(lvStr, func(r rune) bool {
			return r == ',' || r == ' '
		}))
	}

	vs := ds.Variables().Copy()
	for _, card := range latentVars {
		v := vars.New(len(vs), card, "", true)
		vs.Add(v)
	}
	dname := strings.TrimSuffix(dataFile, path.Ext(dataFile))
	i := 0
	for i < numSamples {
		modelFOut := fmt.Sprintf("%s#%04d.ctree", dname, i)
		ct := model.SampleUniform(vs, tw)
		mi := learner.ComputeMIScore(ct, mutInfo)
		if mi > 19 {
			ct.Write(modelFOut)
			i++
			fmt.Printf("%v: %v\n", i, mi)
		}
	}
}
