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

	// increase latent cardinality and learn parameters until bic stops increasing
	ct, _, _ = paramLearner.Run(ct, ds.IntMaps())
	bic := computeBIC(ct, ds)
	for {
		newlv := vars.New(lv.ID(), lv.NState()+1, "", true)
		newct := createLKMStruct([]*vars.Var{newlv}, gs, nil, -1)
		newct, _, _ = paramLearner.Run(newct, ds.IntMaps())
		newbic := computeBIC(newct, ds)
		if newbic-bic > BicThreshold {
			ct = newct
			bic = newbic
			lv = newlv
		} else {
			break
		}
		// TODO: define some max cardinality criteria here
		// if lv.NState() >= 5 {
		// 	break
		// }
	}
	ct.SetBIC(bic)
	return ct, lv
}

// LearnLKM2L creates a LKM model with two latent variables
// as starting point, the first latent variable is parent of group 1 and the second of group 2
func LearnLKM2L(lvs vars.VarList, gs1, gs2 []vars.VarList, ds *data.Dataset,
	paramLearner emlearner.EMLearner) (*model.CTree, vars.VarList, []vars.VarList, []vars.VarList) {
	// create initial structure and learn parameters
	ct := createLKMStruct(lvs, gs1, gs2, -1)
	ct, _, _ = paramLearner.Run(ct, ds.IntMaps())
	bic := computeBIC(ct, ds)

	// TODO: it could yield better results by evaluating all neighbors before changing
	// evaluate node relocations
	for i := range gs1 {
		if len(gs1) <= 2 {
			break
		}
		newct := createLKMStruct(lvs, gs1, gs2, i)
		newct, _, _ = paramLearner.Run(newct, ds.IntMaps())
		newbic := computeBIC(newct, ds)
		if newbic-bic > BicThreshold {
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
		newlvs := append([]*vars.Var(nil), lvs...)
		for {
			newlvs[i] = vars.New(lv.ID(), newlvs[i].NState()+1, "", true)
			newct := createLKMStruct(newlvs, gs1, gs2, -1)
			newct, _, _ = paramLearner.Run(newct, ds.IntMaps())
			newbic := computeBIC(newct, ds)
			if newbic-bic > BicThreshold {
				ct = newct
				bic = newbic
				lvs[i] = newlvs[i]
			} else {
				break
			}
		}
	}
	ct.SetBIC(bic)
	return ct, lvs, gs1, gs2
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
