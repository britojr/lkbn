package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	yaml "gopkg.in/yaml.v2"

	"github.com/britojr/lkbn/data"
	"github.com/britojr/lkbn/model"
	"github.com/britojr/lkbn/vars"
	"github.com/britojr/utl/errchk"
)

type propStruct struct {
	LatentVars []int `yaml:"latent_vars"`
}

func main() {
	var (
		dataFile, parmFile string
		nSamples, k        int
	)
	flag.StringVar(&dataFile, "d", "", "dataset file in csv format")
	flag.StringVar(&parmFile, "p", "", "parameters file")
	flag.IntVar(&k, "k", 1, "tree-width")
	flag.IntVar(&nSamples, "s", 1, "number of samples")
	// flag.IntVar(&lv, "l", 0, "number of latent variables")

	flag.Parse()
	if len(dataFile) == 0 {
		fmt.Printf("\n error: missing dataset file\n\n")
		flag.PrintDefaults()
		os.Exit(1)
	}
	props := propStruct{}
	if len(parmFile) > 0 {
		data, err := ioutil.ReadFile(parmFile)
		errchk.Check(err, "")
		errchk.Check(yaml.Unmarshal([]byte(data), &props), "")
	}
	dname := strings.TrimSuffix(dataFile, path.Ext(dataFile))

	ds := data.NewDataset(dataFile)
	vs := ds.Variables().Copy()
	for _, card := range props.LatentVars {
		v := vars.New(len(vs), card)
		v.SetLatent(true)
		vs.Add(v)
	}
	for i := 0; i < nSamples; i++ {
		modelFOut := fmt.Sprintf("%s#%04d.ctree", dname, i)
		ct := model.SampleUniform(vs, k)
		ct.Write(modelFOut)
	}
}
