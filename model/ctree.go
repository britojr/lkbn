package model

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/britojr/btbn/ktree"
	"github.com/britojr/lkbn/factor"
	"github.com/britojr/lkbn/vars"
	"github.com/britojr/utl/conv"
	"github.com/britojr/utl/errchk"
	"github.com/britojr/utl/ioutl"
	yaml "gopkg.in/yaml.v2"
)

// CTree defines a structure in clique tree format
// a CTree is a way to group the potentials of the model according to its cliques
// the potentials assossiated with each clique are pointers to the same factors present in the model
type CTree struct {
	nodes      []*CTNode
	root       *CTNode
	score, bic float64
	// family map[*vars.Var]*CTNode
}

type codedTree struct {
	Variables []struct {
		Name   string
		Card   int
		Latent bool
	}
	Nodes []struct {
		ClqVars []string
		Values  []float64
		Parent  []string
	}
}

// NewCTree creates new empty CTree
func NewCTree() *CTree {
	return new(CTree)
}

// ReadCTreeYAML creates new CTree from yaml file
func ReadCTreeYAML(fname string) (c *CTree) {
	data, err := ioutil.ReadFile(fname)
	errchk.Check(err, "")
	return CTreeFromString(string(data))
}

// WriteYAML writes CTree on file
func (c *CTree) WriteYAML(fname string) {
	f := ioutl.CreateFile(fname)
	d := []byte(c.String())
	fmt.Fprintf(f, "# Score: %v\n", c.score)
	fmt.Fprintf(f, "# BIC: %v\n", c.bic)
	f.Write(d)
	f.Close()
}

// CTreeFromString creates new CTree from string
func CTreeFromString(strct string) (c *CTree) {
	t := codedTree{}
	errchk.Check(yaml.Unmarshal([]byte(strct), &t), "")
	c = new(CTree)
	vm := make(map[string]*vars.Var)
	for i, tv := range t.Variables {
		v := vars.New(i, tv.Card, tv.Name, tv.Latent)
		vm[tv.Name] = v
	}
	for _, tnd := range t.Nodes {
		nd := new(CTNode)
		var vs vars.VarList
		for _, tv := range tnd.ClqVars {
			vs.Add(vm[tv])
		}
		nd.pot = factor.New(vs...)
		if len(tnd.Values) > 0 {
			nd.pot.SetValues(tnd.Values)
		}
		c.nodes = append(c.nodes, nd)
		if len(tnd.Parent) == 0 {
			c.root = nd
		}
	}
	for i, tnd := range t.Nodes {
		if len(tnd.Parent) == 0 {
			continue
		}
		var vs vars.VarList
		for _, tv := range tnd.Parent {
			vs.Add(vm[tv])
		}
		pa := c.FindNode(vs)
		c.nodes[i].parent = pa
		pa.children = append(pa.children, c.nodes[i])
	}
	return
}

func (c *CTree) String() string {
	t := codedTree{}
	vm := make(map[int]*vars.Var)
	t.Nodes = make([]struct {
		ClqVars []string
		Values  []float64
		Parent  []string
	}, len(c.nodes))
	for i, nd := range c.nodes {
		clqvars := []string(nil)
		for _, v := range nd.Variables() {
			clqvars = append(clqvars, v.Name())
			vm[v.ID()] = v
		}
		t.Nodes[i].ClqVars = clqvars
		t.Nodes[i].Values = nd.pot.Values()
		if nd.parent == nil {
			continue
		}
		pavars := []string(nil)
		for _, v := range nd.parent.Variables() {
			pavars = append(pavars, v.Name())
		}
		t.Nodes[i].Parent = pavars
	}

	t.Variables = make([]struct {
		Name   string
		Card   int
		Latent bool
	}, len(vm))
	for i := range t.Variables {
		t.Variables[i].Name = vm[i].Name()
		t.Variables[i].Card = vm[i].NState()
		t.Variables[i].Latent = vm[i].Latent()
	}

	d, err := yaml.Marshal(&t)
	errchk.Check(err, "")
	return string(d)
}

// XMLStruct creates a struct that can be marshalled into xmlbif format
func (c *CTree) XMLStruct() (ctStruct Network) {
	vs := c.Variables()
	for _, v := range vs {
		ctStruct.Variables = append(ctStruct.Variables, Variable{Name: v.Name(), States: v.States()})
	}
	for _, nd := range c.Nodes() {
		p := Prob{}
		if nd.Parent() == nil {
			for _, x := range nd.Variables() {
				p.For = append(p.For, x.Name())
			}
			p.Table = strings.Join(conv.Sftoa(nd.Potential().Values()), " ")
		} else {
			pavx := []*vars.Var{}
			for _, x := range nd.Variables().Diff(nd.Parent().Variables()) {
				p.For = append(p.For, x.Name())
				pavx = append(pavx, x)
			}
			for _, u := range nd.Variables().Intersec(nd.Parent().Variables()) {
				p.Given = append(p.Given, u.Name())
				pavx = append(pavx, u)
			}
			ixf := vars.NewOrderedIndex(nd.Variables(), pavx)
			values := nd.Potential().Values()
			tableVals := make([]float64, len(values))
			for i := 0; !ixf.Ended(); i++ {
				tableVals[ixf.I()] = values[i]
				ixf.Next()
			}
			p.Table = strings.Join(conv.Sftoa(tableVals), " ")
		}
		ctStruct.Probs = append(ctStruct.Probs, p)
	}
	return
}

// SampleUniform uniformly samples a ktree
func SampleUniform(vs vars.VarList, k int) *CTree {
	n := len(vs)
	children, clqs := ktree.UniformSampleAdj(n, k)
	nodes := []*CTNode(nil)
	for i := range clqs {
		var cvars vars.VarList
		for _, v := range clqs[i] {
			cvars.Add(vs[v])
		}
		nd := new(CTNode)
		nd.pot = factor.New(cvars...)
		nodes = append(nodes, nd)
	}
	for i := range children {
		for _, v := range children[i] {
			nodes[i].children = append(nodes[i].children, nodes[v])
			nodes[v].parent = nodes[i]
		}
	}
	ct := new(CTree)
	ct.nodes = nodes
	ct.root = nodes[0]
	return ct
}

// Copy creates a copy of this
func (c *CTree) Copy() (o *CTree) {
	if c.root == nil {
		panic("ctree: no root")
	}
	o = new(CTree)
	o.root = copyNode(c.root)
	o.BfsNodes()
	return
}

func copyNode(nd *CTNode) (o *CTNode) {
	o = new(CTNode)
	o.pot = nd.pot.Copy()
	for i, ch := range nd.children {
		o.children = append(o.children, copyNode(ch))
		o.children[i].parent = o
	}
	return
}

// BfsNodes sets node slice in bfs order
func (c *CTree) BfsNodes() {
	c.nodes = []*CTNode{c.root}
	i := 0
	for i < len(c.nodes) {
		pa := c.nodes[i]
		i++
		for _, ch := range pa.children {
			c.nodes = append(c.nodes, ch)
		}
	}
}

// Equal compares cliques and values of two ctrees
func (c *CTree) Equal(other *CTree) bool {
	if !c.EqualStruct(other) {
		return false
	}
	for _, nd1 := range c.Nodes() {
		nd2 := other.FindNode(nd1.Potential().Variables())
		if nd2 == nil {
			panic("ctree: can't find node")
		}
		for i, v := range nd1.Potential().Values() {
			if nd2.Potential().Values()[i] != v {
				return false
			}
		}
	}
	return true
}

// EqualStruct compares the structure of two ctrees
func (c *CTree) EqualStruct(other *CTree) bool {
	m1, m2 := c.varIDtoNeighbors(), other.varIDtoNeighbors()
	for k, vs := range m1 {
		if !vs.Equal(m2[k]) {
			return false
		}
	}
	return true
}

// varIDtoNeighbors returns a mapping from variables to their neighbors
func (c *CTree) varIDtoNeighbors() map[int]vars.VarList {
	m := make(map[int]vars.VarList)
	for _, nd := range c.nodes {
		vs := nd.Variables()
		for i := 0; i < len(vs); i++ {
			for j := i + 1; j < len(vs); j++ {
				if mv, ok := m[vs[i].ID()]; ok {
					m[vs[i].ID()] = mv.Add(vs[j])
				} else {
					m[vs[i].ID()] = []*vars.Var{vs[j]}
				}
				if mv, ok := m[vs[j].ID()]; ok {
					m[vs[j].ID()] = mv.Add(vs[i])
				} else {
					m[vs[j].ID()] = []*vars.Var{vs[i]}
				}
			}
		}
	}
	return m
}

// VarsNeighbors returns a mapping from variables to their neighbors
func (c *CTree) VarsNeighbors() map[*vars.Var]vars.VarList {
	// TODO: rethink this method (VarsToNeighbors)
	// maybe cache the map (apply the same to nodes slice)
	// set to null if a change in the structure occurs
	// and compute when asked (then save in cache)
	m := make(map[*vars.Var]vars.VarList)
	for _, nd := range c.nodes {
		vs := nd.Variables()
		for i := 0; i < len(vs); i++ {
			for j := i + 1; j < len(vs); j++ {
				if mv, ok := m[vs[i]]; ok {
					m[vs[i]] = mv.Add(vs[j])
				} else {
					m[vs[i]] = []*vars.Var{vs[j]}
				}
				if mv, ok := m[vs[j]]; ok {
					m[vs[j]] = mv.Add(vs[i])
				} else {
					m[vs[j]] = []*vars.Var{vs[i]}
				}
			}
		}
	}
	return m
}

// Variables returns a list of all the variables
func (c *CTree) Variables() (vs vars.VarList) {
	// TODO: improve by caching the varlist
	for _, nd := range c.nodes {
		vs = vs.Union(nd.Variables())
	}
	return vs
}

// AddNode add node to tree
func (c *CTree) AddNode(nd *CTNode) {
	c.nodes = append(c.nodes, nd)
	if c.root == nil {
		c.root = nd
	}
}

// Score return ctree score
func (c *CTree) Score() float64 {
	return c.score
}

// SetScore set ctree score
func (c *CTree) SetScore(score float64) {
	c.score = score
}

// BIC return ctree bic score
func (c *CTree) BIC() float64 {
	return c.bic
}

// SetBIC set ctree bic score
func (c *CTree) SetBIC(bic float64) {
	c.bic = bic
}

// Len return number of nodes in the tree
func (c *CTree) Len() int {
	return len(c.nodes)
}

// Root return root node
func (c *CTree) Root() *CTNode {
	return c.root
}

// Nodes return list of nodes
func (c *CTree) Nodes() []*CTNode {
	return c.nodes
}

// FindNode return the node that contains exactly the given variables
func (c *CTree) FindNode(vs vars.VarList) *CTNode {
	for _, nd := range c.nodes {
		if nd.Variables().Equal(vs) {
			return nd
		}
	}
	return nil
}

// FindNodeContaining returns a node that contains the given variables
func (c *CTree) FindNodeContaining(vs vars.VarList) *CTNode {
	for _, nd := range c.nodes {
		if nd.Variables().Contains(vs) {
			return nd
		}
	}
	return nil
}

// CTNode defines a clique tree node
type CTNode struct {
	children []*CTNode
	parent   *CTNode
	pot      *factor.Factor
}

// NewCTNode creates new empty CTNode
func NewCTNode() *CTNode {
	return new(CTNode)
}

// Variables return node variables
func (cn *CTNode) Variables() vars.VarList {
	return cn.pot.Variables()
}

// Potential return node potential
func (cn *CTNode) Potential() *factor.Factor {
	return cn.pot
}

// SetPotential set node potential
func (cn *CTNode) SetPotential(p *factor.Factor) {
	cn.pot = p
}

// Children return node children
func (cn *CTNode) Children() []*CTNode {
	return cn.children
}

// Parent return node parent
func (cn *CTNode) Parent() *CTNode {
	return cn.parent
}

// AddChild add child node and updates parent
func (cn *CTNode) AddChild(ch *CTNode) {
	cn.children = append(cn.children, ch)
	ch.parent = cn
}

// RemoveChild removes child node and updates parent
func (cn *CTNode) RemoveChild(ch *CTNode) {
	j := -1
	for i, nd := range cn.children {
		if nd == ch {
			j = i
			break
		}
	}
	if j >= 0 {
		cn.children = append(cn.children[:j], cn.children[j+1:]...)
		ch.parent = nil
	}
}

func (cn *CTNode) String() string {
	return fmt.Sprint(cn.pot.Variables())
}
