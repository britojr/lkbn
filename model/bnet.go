package model

import (
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/britojr/lkbn/factor"
	"github.com/britojr/lkbn/vars"
	"github.com/britojr/utl/conv"
	"github.com/britojr/utl/errchk"
	"github.com/britojr/utl/ioutl"
)

// BNet defines a Bayesian network model
type BNet struct {
	vs    vars.VarList
	nodes map[*vars.Var]*BNode
	score float64
	ct    *CTree
}

// NewBNet creates a new BNet model
func NewBNet() *BNet {
	b := new(BNet)
	b.vs = []*vars.Var{}
	b.nodes = make(map[*vars.Var]*BNode)
	return b
}

// AddNode adds node to current bn
func (b *BNet) AddNode(nd *BNode) {
	b.vs.Add(nd.vx)
	b.nodes[nd.vx] = nd
}

// Better ..
func (b *BNet) Better(other interface{}) bool {
	panic("not implemented")
}

// Score return ctree score
func (b *BNet) Score() float64 {
	return b.score
}

// SetScore set ctree score
func (b *BNet) SetScore(score float64) {
	b.score = score
}

// CTree return a ctree for this bnet
func (b *BNet) CTree() *CTree {
	return b.ct
}

// SetCTree sets bnet ctree
func (b *BNet) SetCTree(ct *CTree) {
	b.ct = ct
}

// Node return the respective node of a var
func (b *BNet) Node(v *vars.Var) *BNode {
	return b.nodes[v]
}

// Variables returns bnet variables
func (b *BNet) Variables() vars.VarList {
	return b.vs
}

// ReadBNetXML creates new BNet from xmlbif file
func ReadBNetXML(fname string) *BNet {
	b := NewBNet()
	xmlfile := readXMLBIF(fname)
	xmlbn := xmlfile.BNetXML
	for i, v := range xmlbn.Variables {
		u := vars.New(i, len(v.States), v.Name, false)
		b.vs.Add(u)
	}
	for _, p := range xmlbn.Probs {
		vx := b.vs.FindByName(p.For[0])
		if len(p.Given) == 0 {
			values := conv.Satof(strings.Fields(strings.Trim(p.Table, " ")))
			b.nodes[vx] = &BNode{vx, factor.New(vx).SetValues(values)}
		} else {
			pavx, pavl := []*vars.Var{vx}, vars.VarList{vx}
			for _, name := range p.Given {
				u := b.vs.FindByName(name)
				pavx = append(pavx, u)
				pavl.Add(u)
			}
			ixf := vars.NewOrderedIndex(pavx, pavl)
			tableVals := conv.Satof(strings.Fields(strings.Trim(p.Table, " ")))
			values := make([]float64, len(tableVals))
			for i := 0; !ixf.Ended(); i++ {
				values[i] = tableVals[ixf.I()]
				ixf.Next()
			}
			b.nodes[vx] = &BNode{vx, factor.New(pavl...).SetValues(values)}
		}
	}
	return b
}

// XMLStruct creates a struct that can be marshalled into xmlbif format
func (b *BNet) XMLStruct() (xmlStruct Network) {
	for _, v := range b.Variables() {
		xmlStruct.Variables = append(xmlStruct.Variables, Variable{Name: v.Name(), States: v.States()})
		nd := b.Node(v)
		p := Prob{}
		p.For = append(p.For, nd.Variable().Name())
		if len(nd.Parents()) == 0 {
			p.Table = strings.Join(conv.Sftoa(nd.Potential().Values()), " ")
		} else {
			pavx := []*vars.Var{nd.Variable()}
			for _, u := range nd.Parents() {
				p.Given = append(p.Given, u.Name())
				pavx = append(pavx, u)
			}
			ixf := vars.NewOrderedIndex(nd.Potential().Variables(), pavx)
			values := nd.Potential().Values()
			tableVals := make([]float64, len(values))
			for i := 0; !ixf.Ended(); i++ {
				tableVals[ixf.I()] = values[i]
				ixf.Next()
			}
			p.Table = strings.Join(conv.Sftoa(tableVals), " ")
		}
		xmlStruct.Probs = append(xmlStruct.Probs, p)
	}
	return
}

// Write writes BNet on file
func (b *BNet) Write(fname string) {
	f := ioutl.CreateFile(fname)
	defer f.Close()
	xmlData := XMLBIF{}
	xmlData.BNetXML = b.XMLStruct()
	ct := b.CTree()
	if ct != nil {
		xmlData.CTreeXML = ct.XMLStruct()
	}
	data, err := xml.MarshalIndent(xmlData, "", "\t")
	errchk.Check(err, "")
	f.Write([]byte(xml.Header))
	f.Write(data)
}

// MarginalizedFamily returns the marginalized family of x:  P(x, pax)
func (b *BNet) MarginalizedFamily(x *vars.Var) *factor.Factor {
	f := b.nodes[x].cpt.Copy()
	queue := b.nodes[x].Parents().Copy()
	visit := vars.VarList{}
	for len(queue) > 0 {
		pa := queue[0]
		queue = queue[1:]
		visit.Add(pa)
		f.Times(b.nodes[pa].Potential())
		queue = queue.Union(b.nodes[pa].Parents().Diff(visit))
	}
	return f.Marginalize(b.nodes[x].cpt.Variables()...)
}

func (b *BNet) Equal(b2 *BNet) bool {
	if !b.Variables().Equal(b2.Variables()) {
		return false
	}
	for _, v := range b.Variables() {
		nd := b.Node(v)
		nd2 := b2.Node(v)
		if nd == nil || nd2 == nil {
			fmt.Println("found nil")
			fmt.Println(b.Variables())
			fmt.Println(v)
			fmt.Println(nd)
			fmt.Println(nd2)
			return false
		}
		if !nd.Equal(nd2) {
			return false
		}
	}
	return true
}

func (b *BNet) String() string {
	s := ""
	for _, v := range b.Variables() {
		s += fmt.Sprintf("%v: %v\n", v, b.Node(v))
	}
	return s
}

// BNode defines a BN node
type BNode struct {
	vx  *vars.Var
	cpt *factor.Factor
}

// NewBNode creates a new node
func NewBNode(vx *vars.Var) *BNode {
	return &BNode{vx, nil}
}

// SetPotential set node potential
func (nd *BNode) SetPotential(p *factor.Factor) {
	nd.cpt = p
}

// Variable returns pivot variable
func (nd *BNode) Variable() *vars.Var {
	return nd.vx
}

// Parents returns parents variables
func (nd *BNode) Parents() vars.VarList {
	return nd.cpt.Variables().Diff(vars.VarList{nd.vx})
}

// Potential return node potential
func (nd *BNode) Potential() *factor.Factor {
	return nd.cpt
}

func (nd *BNode) Equal(nd2 *BNode) bool {
	if nd.vx.ID() != nd2.vx.ID() {
		return false
	}
	return nd.cpt.Equal(nd2.cpt)
}

func (nd *BNode) String() string {
	return fmt.Sprintf("%v|%v", nd.vx, nd.Parents())
}
