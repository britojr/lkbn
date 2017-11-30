package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/britojr/btbn/scr"
	"github.com/britojr/lkbn/data"
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
		mi := computeMIScore(ct, mutInfo)
		if mi > 18 {
			ct.Write(modelFOut)
			i++
		}
	}
}

// TODO: remove this
func computeMIScore(ct *model.CTree, mutInfo *scr.MutInfo) (mi float64) {
	m := ct.VarsNeighbors()
	for v, ne := range m {
		for _, w := range ne {
			if w.ID() < v.ID() {
				break
			}
			mi += linkMI(v, w, m, mutInfo)
		}
	}
	return
}

func linkMI(v, w *vars.Var, m map[*vars.Var]vars.VarList, mutInfo *scr.MutInfo) float64 {
	if !v.Latent() && !w.Latent() {
		return mutInfo.Get(v.ID(), w.ID())
	}
	if v.ID() > w.ID() {
		v, w = w, v
	}
	ne := m[w].Diff(m[v])
	var max float64
	for _, u := range ne {
		if u.ID() == v.ID() {
			continue
		}
		mi := linkMI(v, u, m, mutInfo)
		if mi > max {
			max = mi
		}
	}
	return max
}
