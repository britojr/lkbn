package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/britojr/lkbn/data"
	"github.com/britojr/lkbn/factor"
	"github.com/britojr/lkbn/model"
	"github.com/britojr/lkbn/vars"
	"github.com/britojr/utl/conv"
	"github.com/britojr/utl/errchk"
	"github.com/britojr/utl/ioutl"
)

// conversion types
const (
	biToCtree = "bi-ctree"
	biToBif   = "bi-bif"
	xmlToBif  = "xml-bif"
	biToXML   = "bi-xml"
)

// converts an LTM in bif format to ctree format
func main() {
	var inFile, outFile, convType, dataFile string
	flag.StringVar(&inFile, "i", "", "input file")
	flag.StringVar(&outFile, "o", "", "output file")
	flag.StringVar(&dataFile, "d", "", "dataset file")
	flag.StringVar(&convType, "t", biToCtree, "conversion type ("+strings.Join([]string{biToCtree, biToBif, xmlToBif, biToXML}, "|")+")")
	flag.Parse()

	if len(inFile) == 0 || len(outFile) == 0 {
		fmt.Printf("\n error: missing input/output file name\n\n")
		flag.PrintDefaults()
		os.Exit(1)
	}
	var vs vars.VarList
	if len(dataFile) != 0 {
		vs = data.NewDataset(dataFile).Variables()
	} else {
		vs = []*vars.Var{}
	}
	switch convType {
	case biToCtree:
		potentials, _ := parseLTMbif(inFile, vs)
		ct := buildCTree(potentials)
		ct.WriteYAML(outFile)
	case biToBif:
		potentials, _ := parseLTMbif(inFile, vs)
		ct := buildCTree(potentials)
		writeBif(ct, outFile)
	case xmlToBif:
		writeXMLToBif(inFile, outFile)
	case biToXML:
		potentials, _ := parseLTMbif(inFile, vs)
		ct := buildCTree(potentials)
		writeXML(ct, outFile)
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

			ixf := vars.NewIndexFor(pavs, pavs)
			for !ixf.Ended() {
				attrbMap := ixf.Attribution()
				attrbStr := make([]string, 0, len(attrbMap))
				for _, v := range pavs {
					attrbStr = append(attrbStr, varStates(v)[attrbMap[v.ID()]])
				}
				p := nd.Potential().Copy()
				p.Reduce(attrbMap).SumOut(pavs...)
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

func maxID(vs vars.VarList) int {
	if len(vs) > 0 {
		return vs[len(vs)-1].ID()
	}
	return -1
}

func parseLTMbif(fname string, vs vars.VarList) ([]*factor.Factor, vars.VarList) {
	var (
		pots    []*factor.Factor
		nstate  int
		w, name string
		latent  bool
	)
	id := maxID(vs) + 1
	fi := ioutl.OpenFile(fname)
	defer fi.Close()

	_, err := fmt.Fscanf(fi, "%s", &w)
	for err != io.EOF {
		if w == "variable" {
			fmt.Fscanf(fi, "%s", &name)
			name = strings.Trim(name, "\"")
			v := findVar(vs, name)
			if v == nil {
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
		}
		if w == "probability" {
			varOrd := make([]*vars.Var, 0, 2)
			clq := vars.VarList{}
			fmt.Fscanf(fi, "%s", &w)
			fmt.Fscanf(fi, "%s", &name)
			name = strings.Trim(name, "\"")
			varOrd = append(varOrd, findVar(vs, name))
			clq.Add(varOrd[0])
			fmt.Fscanf(fi, "%s", &w)
			if w == "|" {
				fmt.Fscanf(fi, "%s", &name)
				name = strings.Trim(name, "\"")
				varOrd = append(varOrd, findVar(vs, name))
				clq.Add(varOrd[1])
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

			if len(clq) == 1 {
				pots = append(pots, factor.New(clq...).SetValues(values))
			} else {
				// need to invert variable order
				arranged := make([]float64, len(values))
				ixf := vars.NewOrderedIndex(clq, varOrd)
				for _, v := range values {
					arranged[ixf.I()] = v
					ixf.NextRight()
				}
				pots = append(pots, factor.New(clq...).SetValues(arranged))
			}
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

// XMLBIF defines xmlbif structure
type XMLBIF struct {
	Network Net `xml:"NETWORK"`
}

// Net defines network in xmlbif pattern
type Net struct {
	Name      string     `xml:"NAME"`
	Variables []Variable `xml:"VARIABLE"`
	Probs     []Prob     `xml:"DEFINITION"`
}

// Variable a variable in xmlbif
type Variable struct {
	Name   string   `xml:"NAME"`
	States []string `xml:"OUTCOME"`
}

// Prob conditional probability in xmlbif
type Prob struct {
	For   string   `xml:"FOR"`
	Given []string `xml:"GIVEN"`
	Table string   `xml:"TABLE"`
}

func parseXMLBif(fname string) Net {
	f := ioutl.OpenFile(fname)
	defer f.Close()
	b, err := ioutil.ReadAll(f)
	errchk.Check(err, "")
	var bn XMLBIF
	err = xml.Unmarshal(b, &bn)
	errchk.Check(err, "")
	return bn.Network
}

func writeXMLToBif(inFile, outFile string) {
	xmlbn := parseXMLBif(inFile)

	f := ioutl.CreateFile(outFile)
	defer f.Close()
	fmt.Fprintf(f, "network %v {}\n", xmlbn.Name)
	vs := vars.VarList{}
	for i, v := range xmlbn.Variables {
		u := vars.New(i, len(v.States), v.Name, false)
		fmt.Fprintf(f, "variable %v {\n", u.Name())
		fmt.Fprintf(f, "  type discrete [ %v ] { %v };\n", u.NState(), strings.Join(varStates(u), ", "))
		fmt.Fprintf(f, "}\n")
		vs.Add(u)
	}
	for _, p := range xmlbn.Probs {
		if len(p.Given) > 0 {
			fmt.Fprintf(f, "probability ( %v | %v ) {\n", p.For, strings.Join(p.Given, ", "))
			xv := findVar(vs, p.For)
			pavs := []*vars.Var{}
			for _, name := range p.Given {
				pavs = append(pavs, findVar(vs, name))
			}
			ixf := vars.NewOrderedIndex(pavs, pavs)
			k := 0
			tableVals := strings.Fields(strings.Trim(p.Table, " "))
			for !ixf.Ended() {
				attrbMap := ixf.Attribution()
				attrbStr := make([]string, 0, len(attrbMap))
				for _, v := range pavs {
					attrbStr = append(attrbStr, varStates(v)[attrbMap[v.ID()]])
				}
				tableInd := strings.Join(attrbStr, ", ")
				tableVal := strings.Join(tableVals[k:k+xv.NState()], ", ")
				fmt.Fprintf(f, "  (%v) %v;\n", tableInd, tableVal)
				ixf.Next()
				k += xv.NState()
			}
		} else {
			fmt.Fprintf(f, "probability ( %v ) {\n", p.For)
			fmt.Fprintf(f, "  table %v;\n", strings.Replace(strings.Trim(p.Table, " "), " ", ", ", -1))
		}
		fmt.Fprintf(f, "}\n")
	}

	return
}

func writeXML(ct *model.CTree, fname string) {
	f := ioutl.CreateFile(fname)
	defer f.Close()
	bn := XMLBIF{Net{}}
	for _, v := range ct.Variables() {
		bn.Network.Variables = append(bn.Network.Variables, Variable{Name: v.Name(), States: varStates(v)})
	}
	for _, nd := range ct.Nodes() {
		p := Prob{}
		if nd.Parent() == nil {
			p.For = nd.Variables()[0].Name()
			p.Table = strings.Join(conv.Sftoa(nd.Potential().Values()), " ")
		} else {
			vx := nd.Variables().Diff(nd.Parent().Variables())[0]
			p.For = vx.Name()
			pavx, pavl := []*vars.Var{vx}, vars.VarList{vx}
			for _, u := range nd.Variables().Intersec(nd.Parent().Variables()) {
				p.Given = append(p.Given, u.Name())
				pavx = append(pavx, u)
				pavl.Add(u)
			}
			ixf := vars.NewOrderedIndex(pavl, pavx)
			values := nd.Potential().Values()
			tableVals := make([]float64, len(values))
			for i := 0; !ixf.Ended(); i++ {
				tableVals[ixf.I()] = values[i]
				ixf.Next()
			}
			p.Table = strings.Join(conv.Sftoa(tableVals), " ")
		}
		bn.Network.Probs = append(bn.Network.Probs, p)
	}

	data, err := xml.MarshalIndent(bn, "", "\t")
	errchk.Check(err, "")
	f.Write(data)
}
