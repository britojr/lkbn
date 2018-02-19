package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/britojr/lkbn/factor"
	"github.com/britojr/lkbn/model"
	"github.com/britojr/lkbn/vars"
	"github.com/britojr/utl/conv"
	"github.com/britojr/utl/ioutl"
)

// converts an LTM in bif format to ctree format
func main() {
	var inFile, outFile string
	flag.StringVar(&inFile, "i", "", "input file")
	flag.StringVar(&outFile, "o", "", "output file")

	flag.Parse()
	if len(inFile) == 0 || len(outFile) == 0 {
		fmt.Printf("\n error: missing input/output file name\n\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// open files
	fi := ioutl.OpenFile(inFile)
	defer fi.Close()

	// aux vars
	var (
		vs         vars.VarList
		pots       []*factor.Factor
		nstate, id int
		w, name    string
		latent     bool
	)

	// parse
	_, err := fmt.Fscanf(fi, "%s", &w)
	for err != io.EOF {
		if w == "variable" {
			fmt.Fscanf(fi, "%s", &name)
			name = strings.Trim(name, "\"")
			latent = false
			if strings.Index(name, "variable") >= 0 {
				latent = true
			}
			for strings.Index(w, "discrete") != 0 {
				fmt.Fscanf(fi, "%s", &w)
			}
			nstate = conv.Atoi(strings.Trim(w[len("discrete"):], "[]"))
			vs.Add(vars.New(id, nstate, name, latent))
			id++
		}
		if w == "probability" {
			clq := vars.VarList{}
			fmt.Fscanf(fi, "%s", &w)
			fmt.Fscanf(fi, "%s", &name)
			name = strings.Trim(name, "\"")
			clq.Add(findVar(vs, name))
			fmt.Fscanf(fi, "%s", &w)
			if w == "|" {
				fmt.Fscanf(fi, "%s", &name)
				name = strings.Trim(name, "\"")
				clq.Add(findVar(vs, name))
			}

			for strings.Index(w, "table") != 0 {
				fmt.Fscanf(fi, "%s", &w)
			}
			values := []float64{}
			fmt.Fscanf(fi, "%s", &w)
			for w != "}" {
				w = strings.Trim(w, ";")
				values = append(values, conv.Atof(w))
				fmt.Fscanf(fi, "%s", &w)
			}
			pots = append(pots, factor.New(clq...).SetValues(values))
		}
		_, err = fmt.Fscanf(fi, "%s", &w)
	}
	ct := buildCTree(pots)
	ct.Write(outFile)
}

func findVar(vs vars.VarList, name string) (v *vars.Var) {
	for _, u := range vs {
		if u.Name() == name {
			v = u
			break
		}
	}
	return
}

func buildCTree(fs []*factor.Factor) *model.CTree {
	var r *factor.Factor
	var fi []*factor.Factor

	r, fs = getRoot(fs)
	nd := model.NewCTNode()
	nd.SetPotential(r)
	queue := []*model.CTNode{nd}

	ct := model.NewCTree()
	for len(queue) > 0 {
		nd := queue[0]
		queue = queue[1:]
		ct.AddNode(nd)
		fi, fs = getMaxIntersec(nd.Potential(), fs)
		for _, f := range fi {
			ch := model.NewCTNode()
			ch.SetPotential(f)
			nd.AddChild(ch)
			queue = append(queue, ch)
		}
	}
	return ct
}

func getRoot(fs []*factor.Factor) (*factor.Factor, []*factor.Factor) {
	f, i := fs[0], 0
	for j, g := range fs {
		if len(g.Variables()) < len(f.Variables()) {
			f, i = g, j
		}
	}
	return f, append(fs[:i], fs[i+1:]...)
}

func getMaxIntersec(f *factor.Factor, fs []*factor.Factor) (fi []*factor.Factor, fr []*factor.Factor) {
	max := 0
	for _, g := range fs {
		l := len(f.Variables().Intersec(g.Variables()))
		if l > max {
			max = l
		}
	}
	if max == 0 {
		return nil, fs
	}
	for _, g := range fs {
		l := len(f.Variables().Intersec(g.Variables()))
		if l == max {
			fi = append(fi, g)
		} else {
			fr = append(fr, g)
		}
	}
	return
}