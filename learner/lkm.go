package learner

import (
	"github.com/britojr/lkbn/data"
	"github.com/britojr/lkbn/emlearner"
	"github.com/britojr/lkbn/factor"
	"github.com/britojr/lkbn/model"
	"github.com/britojr/lkbn/vars"
)

func computeBIC(ct *model.CTree) float64 {
	// TODO: replace temporary approximation of BIC by correct equation
	numparms := 0
	for _, nd := range ct.Nodes() {
		numparms += len(nd.Potential().Values())
	}
	return ct.Score() - float64(numparms)
}

// creates a new tree structure with lv as parent of every clique
func createLKM1LStruct(gs []vars.VarList, lv *vars.Var) *model.CTree {
	ct := model.NewCTree()
	root := model.NewCTNode()
	root.SetPotential(factor.New(gs[0].Union([]*vars.Var{lv})...))
	ct.AddNode(root)
	for _, g := range gs[1:] {
		nd := model.NewCTNode()
		nd.SetPotential(factor.New(g.Union([]*vars.Var{lv})...))
		root.AddChildren(nd)
		ct.AddNode(nd)
	}
	return ct
}

func learnLKM1L(gs []vars.VarList, ds *data.Dataset, paramLearner emlearner.EMLearner) *model.CTree {
	// create new latent variable
	nstate := 2
	lv := vars.New(len(ds.Variables()), nstate, "", true)
	ct := createLKM1LStruct(gs, lv)

	// increase latent cardinality and learn parameters until bic stops increasing
	ct, _, _ = paramLearner.Run(ct, ds.IntMaps())
	bic := computeBIC(ct)
	for {
		nstate++
		lv2 := vars.New(len(ds.Variables()), nstate, "", true)
		newct := createLKM1LStruct(gs, lv2)
		newct, _, _ = paramLearner.Run(newct, ds.IntMaps())
		newbic := computeBIC(newct)
		if newbic <= bic {
			break
		}
		ct = newct
		bic = newbic
	}
	ct.SetBIC(bic)
	return ct
}

// creates a copy of the given ctree replacing v1 by v2
func copyReplace(ct *model.CTree, v1, v2 *vars.Var) *model.CTree {
	ot := ct.Copy()
	for _, nd := range ot.Nodes() {
		vs := nd.Variables()
		vs.Remove(v1.ID())
		nd.SetPotential(factor.New(vs.Add(v2)...))
	}
	return ot
}

func learnLKM2L(gs []vars.VarList, ds *data.Dataset, paramLearner emlearner.EMLearner) (*model.CTree, int) {
	// TODO: implement correct 2L
	// create new latent variable
	nstate := 2
	lvs := []*vars.Var{
		vars.New(len(ds.Variables()), nstate, "", true),
		vars.New(len(ds.Variables())+1, nstate, "", true),
	}
	// mount structure
	f := factor.New(gs[0].Union([]*vars.Var{lvs[0]})...)
	ct := model.NewCTree()
	root := model.NewCTNode()
	root.SetPotential(f)
	ct.AddNode(root)
	for _, g := range gs[1:] {
		f := factor.New(g.Union([]*vars.Var{lvs[0]})...)
		nd := model.NewCTNode()
		nd.SetPotential(f)
		root.AddChildren(nd)
		ct.AddNode(nd)
	}

	// increase latent cardinality and learn parameters until bic stops increasing
	ct, _, _ = paramLearner.Run(ct, ds.IntMaps())
	bic := computeBIC(ct)
	for {
		nstate++
		lvs[0].SetNState(nstate)
		ct, _, _ = paramLearner.Run(ct, ds.IntMaps())
		newbic := computeBIC(ct)
		if newbic <= bic {
			break
		}
		bic = newbic
	}
	return ct, len(lvs)
}
