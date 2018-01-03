package learner

import (
	"github.com/britojr/lkbn/data"
	"github.com/britojr/lkbn/emlearner"
	"github.com/britojr/lkbn/factor"
	"github.com/britojr/lkbn/model"
	"github.com/britojr/lkbn/vars"
)

var bicThreshold = 0.1

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
		if newbic-bic > bicThreshold {
			ct = newct
			bic = newbic
		} else {
			break
		}
	}
	ct.SetBIC(bic)
	return ct
}

// creates a LKM model with two latent variables
// as starting point, the first latent variable is parent of group 1 and the second of group 2
func learnLKM2L(gs1, gs2 []vars.VarList, ds *data.Dataset,
	paramLearner emlearner.EMLearner) (*model.CTree, []vars.VarList, []vars.VarList) {
	// create new latent variables and mount structure
	nstate := 2
	lvs := []*vars.Var{
		vars.New(len(ds.Variables()), nstate, "", true),
		vars.New(len(ds.Variables())+1, nstate, "", true),
	}
	ct := createLKM2LStruct(gs1, gs2, -1, lvs)

	// learn parameters and compute score
	ct, _, _ = paramLearner.Run(ct, ds.IntMaps())
	bic := computeBIC(ct)

	// evaluate node relocations
	// TODO: it could yield better results by evaluating all neighbors before changing
	for i := range gs1 {
		if len(gs1) <= 2 {
			break
		}
		newct := createLKM2LStruct(gs1, gs2, i, lvs)
		newct, _, _ = paramLearner.Run(newct, ds.IntMaps())
		newbic := computeBIC(newct)
		if newbic-bic > bicThreshold {
			ct = newct
			bic = newbic
			gs2 = append(gs2, gs1[i])
			gs1 = append(gs1[:i], gs1[i+1:]...)
		} else {
			break
		}
	}

	// increase latent cardinality and learn parameters until bic stops increasing
	for i, lv := range lvs {
		lvs2 := append([]*vars.Var(nil), lvs...)
		for {
			lvs2[i] = vars.New(lv.ID(), lvs2[i].NState()+1, "", true)
			newct := createLKM2LStruct(gs1, gs2, -1, lvs2)
			newct, _, _ = paramLearner.Run(newct, ds.IntMaps())
			newbic := computeBIC(newct)
			if newbic-bic > bicThreshold {
				ct = newct
				bic = newbic
				lvs[i] = lvs2[i]
			} else {
				break
			}
		}
	}
	ct.SetBIC(bic)
	return ct, gs1, gs2
}

func createLKM2LStruct(gs1, gs2 []vars.VarList, reloc int, lvs []*vars.Var) *model.CTree {
	ct := model.NewCTree()
	root := model.NewCTNode()
	root.SetPotential(factor.New(lvs...))
	ct.AddNode(root)
	for i, g := range gs1 {
		if i == reloc {
			continue
		}
		nd := model.NewCTNode()
		nd.SetPotential(factor.New(g.Union([]*vars.Var{lvs[0]})...))
		root.AddChildren(nd)
		ct.AddNode(nd)
	}
	for _, g := range gs2 {
		nd := model.NewCTNode()
		nd.SetPotential(factor.New(g.Union([]*vars.Var{lvs[1]})...))
		root.AddChildren(nd)
		ct.AddNode(nd)
	}
	if reloc >= 0 && reloc < len(gs1) {
		nd := model.NewCTNode()
		nd.SetPotential(factor.New(gs1[reloc].Union([]*vars.Var{lvs[1]})...))
		root.AddChildren(nd)
		ct.AddNode(nd)
	}
	return ct
}

//
// // creates a copy of the given ctree replacing v1 by v2
// func copyReplace(ct *model.CTree, v1, v2 *vars.Var) *model.CTree {
// 	ot := ct.Copy()
// 	for _, nd := range ot.Nodes() {
// 		vs := nd.Variables()
// 		vs.Remove(v1.ID())
// 		nd.SetPotential(factor.New(vs.Add(v2)...))
// 	}
// 	return ot
// }
