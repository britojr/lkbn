package model

import (
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

// Read creates new CTree from file
func Read(fname string) *CTree {
	return new(CTree)
}

// Write writes CTree on file
func (c *CTree) Write(fname string) {

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
