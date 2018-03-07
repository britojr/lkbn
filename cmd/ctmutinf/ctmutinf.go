package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/britojr/btbn/scr"
	"github.com/britojr/lkbn/model"
	"github.com/britojr/lkbn/vars"
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
	// ds := data.NewDataset(dataFile)
	// var props map[string]string
	// if len(parmFile) > 0 {
	// 	props = ioutl.ReadYaml(parmFile)
	// }

	mutInfo := scr.ComputeMutInf(dataFile)
	dname := strings.TrimSuffix(dataFile, path.Ext(dataFile))
	ctreefs, _ := filepath.Glob(fmt.Sprintf("%v*.xml", dname))
	for _, ctFile := range ctreefs {
		ct := model.ReadCTreeXML(ctFile)
		mi := computeMIScore(ct, mutInfo)

		log.Printf("========== SOLUTION ==============================\n")
		log.Printf("Structure: %v\n", ctFile)
		log.Printf("Mutual Information: %v\n", mi)
		log.Printf("--------------------------------------------------\n")
	}
}

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
