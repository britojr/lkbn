package model

import (
	"io/ioutil"

	"github.com/britojr/btbn/ktree"
	"github.com/britojr/lkbn/factor"
	"github.com/britojr/lkbn/vars"
	"github.com/britojr/utl/errchk"
	"github.com/britojr/utl/ioutl"
	yaml "gopkg.in/yaml.v2"
)

// CTree defines a structure in clique tree format
// a CTree is a way to group the potentials of the model according to its cliques
// the potentials assossiated with each clique are pointers to the same factors present in the model
type CTree struct {
	nodes  []*CTNode
	root   *CTNode
	family map[*vars.Var]*CTNode
	score  float64
}

type codedTree struct {
	Variables []struct {
		Name string
		Card int
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

// Read creates new CTree from file
func Read(fname string) (c *CTree) {
	data, err := ioutil.ReadFile(fname)
	errchk.Check(err, "")
	return FromString(string(data))
}

// Write writes CTree on file
func (c *CTree) Write(fname string) {
	f := ioutl.CreateFile(fname)
	d := []byte(c.String())
	f.Write(d)
	f.Close()
}

// FromString creates new CTree from string
func FromString(strct string) (c *CTree) {
	t := codedTree{}
	errchk.Check(yaml.Unmarshal([]byte(strct), &t), "")
	c = new(CTree)
	vm := make(map[string]*vars.Var)
	for i, tv := range t.Variables {
		v := vars.New(i, tv.Card)
		v.SetName(tv.Name)
		vm[tv.Name] = v
	}
	for _, tnd := range t.Nodes {
		nd := new(CTNode)
		var vs vars.VarList
		for _, tv := range tnd.ClqVars {
			vs = append(vs, vm[tv])
		}
		nd.pot = factor.New(vs...)
		nd.pot.SetValues(tnd.Values)
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
			vs = append(vs, vm[tv])
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
		Name string
		Card int
	}, len(vm))
	for i := range t.Variables {
		t.Variables[i].Name = vm[i].Name()
		t.Variables[i].Card = vm[i].NState()
	}

	d, err := yaml.Marshal(&t)
	errchk.Check(err, "")
	return string(d)
}

// SampleUniform uniformly samples a ktree
func SampleUniform(vs vars.VarList, k int) *CTree {
	n := len(vs)
	children, clqs := ktree.UniformSampleAdj(n, k)
	nodes := []*CTNode(nil)
	for i := range clqs {
		cvars := make([]*vars.Var, 0, len(clqs[i]))
		for _, v := range clqs[i] {
			cvars = append(cvars, vs[v])
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

// Equal compares cliques and values of two ctrees
func (c *CTree) Equal(other *CTree) bool {
	return c.String() == other.String()
}

// VarsNeighbors returns a mapping from variables to their neighbors
func (c *CTree) VarsNeighbors() map[*vars.Var]vars.VarList {
	m := make(map[*vars.Var]vars.VarList)
	for _, nd := range c.nodes {
		vs := nd.Variables()
		for i := 0; i < len(vs); i++ {
			for j := i + 1; j < len(vs); j++ {
				if _, ok := m[vs[i]]; ok {
					m[vs[i]] = m[vs[i]].Add(vs[j])
				} else {
					m[vs[i]] = []*vars.Var{vs[j]}
				}
				if _, ok := m[vs[j]]; ok {
					m[vs[j]] = m[vs[j]].Add(vs[i])
				} else {
					m[vs[j]] = []*vars.Var{vs[i]}
				}
			}
		}
	}
	return m
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

// Families return map of var to family
func (c *CTree) Families() map[*vars.Var]*CTNode {
	// return c.family
	panic("not implemented")
}

// ToCTree return a ctree for this
func (c *CTree) ToCTree() *CTree {
	return c
}

// CTNode defines a clique tree node
type CTNode struct {
	children []*CTNode
	parent   *CTNode
	pot      *factor.Factor
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
