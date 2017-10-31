package model

import (
	"github.com/britojr/lkbn/data"
	"github.com/britojr/lkbn/factor"
	"github.com/britojr/lkbn/vars"
)

// BNet defines a Bayesian network model
type BNet struct {
	nodes map[*vars.Var]*BNode
	score float64
}

// BNode defines a BN node
type BNode struct {
	vx  *vars.Var
	cpt *factor.Factor
}

// NewBNet creates a new BNet model
func NewBNet() *BNet {
	b := new(BNet)
	b.nodes = make(map[*vars.Var]*BNode)
	return b
}

// Better ..
func (b *BNet) Better(other interface{}) bool {
	panic("not implemented")
}

// ComputeScore ..
func (b *BNet) ComputeScore(ds *data.Dataset) float64 {
	panic("not implemented")
}

// ToCTree return a ctree for this bnet
func (b *BNet) ToCTree() *CTree {
	panic("not implemented")
}

// Node return the respective node of a var
func (b *BNet) Node(v *vars.Var) *BNode {
	return b.nodes[v]
}

// Variable return node pivot variable
func (nd *BNode) Variable() *vars.Var {
	return nd.vx
}

// Potential return node potential
func (nd *BNode) Potential() *factor.Factor {
	return nd.cpt
}

// SetPotential set node potential
func (nd *BNode) SetPotential(p *factor.Factor) {
	nd.cpt = p
}
