package lkm

import (
	"math"

	"github.com/britojr/lkbn/data"
	"github.com/britojr/lkbn/emlearner"
	"github.com/britojr/lkbn/factor"
	"github.com/britojr/lkbn/model"
	"github.com/britojr/lkbn/vars"
)

// BicThreshold defines the minimum difference to accept a better bic score
var BicThreshold = 0.1

// LearnLKM1L creates a LKM model with one latent variable
func LearnLKM1L(gs []vars.VarList, lv *vars.Var, ds *data.Dataset,
	paramLearner emlearner.EMLearner) (*model.CTree, *vars.Var) {
	// create initial  structure
	ct := createLKMStruct([]*vars.Var{lv}, gs, nil, -1)
	ct, _, _ = paramLearner.Run(ct, ds.IntMaps())
	ct.SetBIC(computeBIC(ct, ds))
	lvs := []*vars.Var{lv}

	ct, lvs = applyStateInsertion(ct, ds, paramLearner, lvs, gs, nil)
	return ct, lvs[0]
}

// LearnLKM2L creates a LKM model with two latent variables
// as starting point, the first latent variable is parent of group 1 and the second of group 2
func LearnLKM2L(lvs vars.VarList, gs1, gs2 []vars.VarList, ds *data.Dataset,
	paramLearner emlearner.EMLearner) (*model.CTree, vars.VarList, []vars.VarList, []vars.VarList) {
	// create initial structure and learn parameters
	ct := createLKMStruct(lvs, gs1, gs2, -1)
	ct, _, _ = paramLearner.Run(ct, ds.IntMaps())
	ct.SetBIC(computeBIC(ct, ds))

	// TODO: correctly apply NI search if necessary
	// TODO: check if its better to use SI before or after NR
	// ct, lvs = applyStateInsertion(ct, ds, paramLearner, lvs, gs1, gs2)
	ct, gs1, gs2 = applyNodeRelocation(ct, ds, paramLearner, lvs, gs1, gs2)
	ct, lvs = applyStateInsertion(ct, ds, paramLearner, lvs, gs1, gs2)
	return ct, lvs, gs1, gs2
}

// increase latent cardinality until bic stops increasing
func applyStateInsertion(ct *model.CTree, ds *data.Dataset, paramLearner emlearner.EMLearner,
	lvs vars.VarList, gs1, gs2 []vars.VarList) (*model.CTree, vars.VarList) {
	for {
		newct, newlvs := bestModelSI(ct, ds, paramLearner, lvs, gs1, gs2)
		if newct.BIC()-ct.BIC() > BicThreshold {
			ct = newct
			lvs = newlvs
		} else {
			break
		}
	}
	return ct, lvs
}

// relocate nodes until bic stops increasing
func applyNodeRelocation(ct *model.CTree, ds *data.Dataset, paramLearner emlearner.EMLearner,
	lvs vars.VarList, gs1, gs2 []vars.VarList) (*model.CTree, []vars.VarList, []vars.VarList) {
	for {
		newct, reloc := bestModelNR(ct, ds, paramLearner, lvs, gs1, gs2)
		if newct != nil && newct.BIC()-ct.BIC() > BicThreshold {
			ct = newct
			gs2 = append(gs2, gs1[reloc])
			gs1 = append(gs1[:reloc], gs1[reloc+1:]...)
		} else {
			break
		}
	}
	return ct, gs1, gs2
}

func bestModelSI(ct *model.CTree, ds *data.Dataset, paramLearner emlearner.EMLearner,
	lvs vars.VarList, gs1, gs2 []vars.VarList) (bestct *model.CTree, bestlvs vars.VarList) {
	// TODO: improve this with local EM
	for i, lv := range lvs {
		// TODO: define some max cardinality criteria here
		// if lv.NState() >= 5 {
		// 	break
		// }
		newlvs := append([]*vars.Var(nil), lvs...)
		newlvs[i] = vars.New(lv.ID(), newlvs[i].NState()+1, "", true)
		newct := createLKMStruct(newlvs, gs1, gs2, -1)
		newct, _, _ = paramLearner.Run(newct, ds.IntMaps())
		newct.SetBIC(computeBIC(newct, ds))
		if bestct == nil || newct.BIC() > bestct.BIC() {
			bestct = newct
			bestlvs = newlvs
		}
	}
	return
}

func bestModelNR(ct *model.CTree, ds *data.Dataset, paramLearner emlearner.EMLearner,
	lvs vars.VarList, gs1, gs2 []vars.VarList) (bestct *model.CTree, reloc int) {
	// TODO: improve this with local EM
	// ngs1, ngs2 := append([]*vars.Var(nil), gs1...), append([]*vars.Var(nil), gs2...)
	if len(gs1) <= 2 {
		return
	}
	for i := range gs1 {
		newct := createLKMStruct(lvs, gs1, gs2, i)
		newct, _, _ = paramLearner.Run(newct, ds.IntMaps())
		newct.SetBIC(computeBIC(newct, ds))
		if bestct == nil || newct.BIC() > bestct.BIC() {
			bestct = newct
			reloc = i
		}
	}
	return
}

// creates a new ctree structure with latent variables as parents of every clique
// it can be used with one or two latent variables
func createLKMStruct(lvs []*vars.Var, gs1, gs2 []vars.VarList, reloc int) *model.CTree {
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

// computes Bayesian Information Criterion:
// 	BIC = LL - ( model_size * ln(N)/2 )
func computeBIC(ct *model.CTree, ds *data.Dataset) float64 {
	modelsize := computeModelSize(ct)
	return ct.Score() - float64(modelsize)*math.Log(float64(len(ds.IntMaps())))/2.0
}

func computeModelSize(ct *model.CTree) (modelsize int) {
	for _, nd := range ct.Nodes() {
		modelsize += len(nd.Potential().Values())
		if nd.Parent() != nil {
			modelsize -= nd.Potential().Variables().Intersec(nd.Parent().Variables()).NStates()
		} else {
			modelsize--
		}
	}
	return
}
