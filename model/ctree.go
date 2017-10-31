package model

import (
	"log"
	"reflect"

	"github.com/britojr/btbn/ktree"
	"github.com/britojr/lkbn/factor"
	"github.com/britojr/lkbn/vars"
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

// NewCTree creates new empty CTree
func NewCTree() *CTree {
	return new(CTree)
}

// SampleUniform ..
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

// AddNode add node to tree
func (c *CTree) AddNode(nd *CTNode) {
	c.nodes = append(c.nodes, nd)
	if c.root == nil {
		c.root = nd
	}
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

// Families return map of var to family
func (c *CTree) Families() map[*vars.Var]*CTNode {
	return c.family
}

// Better return true if this model is better
func (c *CTree) Better(other interface{}) bool {
	if v, ok := other.(CTree); ok {
		return c.score > v.score
	}
	if v, ok := other.(BNet); ok {
		return c.score > v.score
	}
	log.Panicf("ctree: cannot compare to type '%v'", reflect.TypeOf(other))
	return false
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
