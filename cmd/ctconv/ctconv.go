package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/britojr/lkbn/factor"
	"github.com/britojr/lkbn/model"
	"github.com/britojr/lkbn/vars"
	"github.com/britojr/utl/conv"
	"github.com/britojr/utl/ioutl"
)

// conversion types
const (
	biToCtree = "bi-ctree"
	biToBif   = "bi-bif"
)

// converts an LTM in bif format to ctree format
func main() {
	var inFile, outFile, convType string
	flag.StringVar(&inFile, "i", "", "input file")
	flag.StringVar(&outFile, "o", "", "output file")
	flag.StringVar(&convType, "t", biToCtree, "conversion type ("+strings.Join([]string{biToCtree, biToBif}, "|")+")")
	flag.Parse()

	if len(inFile) == 0 || len(outFile) == 0 {
		fmt.Printf("\n error: missing input/output file name\n\n")
		flag.PrintDefaults()
		os.Exit(1)
	}
	potentials, _ := parseLTMbif(inFile)
	ct := buildCTree(potentials)
	switch convType {
	case biToCtree:
		ct.Write(outFile)
	case biToBif:
		writeBif(ct, outFile)
	default:
		fmt.Printf("\n error: invalid conversion option: (%v)\n\n", convType)
		flag.PrintDefaults()
		os.Exit(1)
	}
}

func writeBif(ct *model.CTree, fname string) {
	f := ioutl.CreateFile(fname)
	defer f.Close()
	fmt.Fprintf(f, "network unknown {}\n")
	vs := ct.Variables()
	for _, v := range vs {
		fmt.Fprintf(f, "variable %v {\n", v.Name())
		fmt.Fprintf(f, "  type discrete [ %v ] { %v };\n", v.NState(), strings.Join(varStates(v), ", "))
		fmt.Fprintf(f, "}\n")
	}
	nds := ct.Nodes()
	for _, nd := range nds {
		if nd.Parent() != nil {
			xvs := nd.Variables().Diff(nd.Parent().Variables())
			pavs := nd.Variables().Intersec(nd.Parent().Variables())
			fmt.Fprintf(f, "probability ( %v | %v ) {\n", strings.Join(varNames(xvs), ", "), strings.Join(varNames(pavs), ", "))

			ixf := vars.NewIndexFor(xvs, xvs)
			for !ixf.Ended() {
				attrbMap := ixf.Attribution()
				attrbStr := make([]string, 0, len(attrbMap))
				for _, v := range xvs {
					attrbStr = append(attrbStr, varStates(v)[attrbMap[v.ID()]])
				}
				p := nd.Potential().Copy()
				p.Reduce(attrbMap).SumOut(xvs...)
				tableInd := strings.Join(attrbStr, ", ")
				tableVal := strings.Join(conv.Sftoa(p.Values()), ", ")
				fmt.Fprintf(f, "  (%v) %v;\n", tableInd, tableVal)
				ixf.Next()
			}
		} else {
			fmt.Fprintf(f, "probability ( %v ) {\n", strings.Join(varNames(nd.Variables()), ", "))
			fmt.Fprintf(f, "  table %v;\n", strings.Join(conv.Sftoa(nd.Potential().Values()), ", "))
		}
		fmt.Fprintf(f, "}\n")
	}
}

func varStates(v *vars.Var) (s []string) {
	for i := 0; i < v.NState(); i++ {
		s = append(s, strconv.Itoa(i))
	}
	return
}

func varNames(vs vars.VarList) (s []string) {
	for _, v := range vs {
		s = append(s, v.Name())
	}
	return
}

func parseLTMbif(fname string) ([]*factor.Factor, vars.VarList) {
	var (
		vs         vars.VarList
		pots       []*factor.Factor
		nstate, id int
		w, name    string
		latent     bool
	)
	fi := ioutl.OpenFile(fname)
	defer fi.Close()

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

			// need to invert variable order
			arranged := make([]float64, len(values))
			ixf := vars.NewIndexFor(clq, clq)
			for _, v := range values {
				arranged[ixf.I()] = v
				ixf.NextRight()
			}
			pots = append(pots, factor.New(clq...).SetValues(arranged))
		}
		_, err = fmt.Fscanf(fi, "%s", &w)
	}
	return pots, vs
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
