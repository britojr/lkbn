package model

import (
	"fmt"
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

// ReadCTree creates new CTree from file
func ReadCTree(fname string) (c *CTree) {
	data, err := ioutil.ReadFile(fname)
	errchk.Check(err, "")
	return CTreeFromString(string(data))
}

// Write writes CTree on file
func (c *CTree) Write(fname string) {
	f := ioutl.CreateFile(fname)
	d := []byte(c.String())
	fmt.Fprintf(f, "# Score: %v\n", c.score)
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
	return c.String() == other.String()
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
	// TODO: improve this with bfs or caching the varlist
	m := c.VarsNeighbors()
	for v := range m {
		vs.Add(v)
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

// AddChildren add children and update parent
func (cn *CTNode) AddChildren(ch *CTNode) {
	cn.children = append(cn.children, ch)
	ch.parent = cn
}

func (cn *CTNode) String() string {
	return fmt.Sprint(cn.pot.Variables())
}
