package scores

import (
	"log"
	"math"

	"github.com/britojr/lkbn/inference"
	"github.com/britojr/lkbn/model"
	"gonum.org/v1/gonum/floats"
)

// ComputeBIC computes Bayesian Information Criterion:
// 	BIC = LL - ( model_size * ln(N)/2 )
// WARNING: this method assumes that LL is computed and set
func ComputeBIC(ct *model.CTree, intMaps []map[int]int) float64 {
	modelsize := computeModelSize(ct)
	return ct.Score() - float64(modelsize)*math.Log(float64(len(intMaps)))/2.0
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

// ComputeLL computes the Log Likelihood of the given model
func ComputeLL(ct *model.CTree, intMaps []map[int]int) float64 {
	infalg := inference.NewCTreeCalibration(ct)
	var ll float64
	for _, evid := range intMaps {
		evLkhood := infalg.Run(evid)
		if evLkhood == 0 {
			log.Panicf("scores: invalid prob of evidence == 0")
		}
		ll += math.Log(evLkhood)
	}
	return ll
}

// KLDiv computes kl-divergence
func KLDiv(orgNet *model.BNet, compNet *model.CTree) (kld float64) {
	infalg := inference.NewCTreeCalibration(compNet)
	for _, v := range orgNet.Variables() {
		pcond := orgNet.Node(v).Potential().Copy()
		family := pcond.Variables()
		// qjoint := infalg.Posterior(family, nil)
		qcond, _ := infalg.Posterior(family, nil).Normalize(v)
		pjoint := orgNet.MarginalizedFamily(v)

		// kld += floats.Sum(pjoint.Times(pcond.Log().Minus(qjoint.Log())).Values())
		kld += floats.Sum(pjoint.Times(pcond.Log().Minus(qcond.Log())).Values())
	}
	return
}

// KLDivBruteForce computes kl-divergence with no simplifications
func KLDivBruteForce(orgNet *model.BNet, compNet *model.CTree) (kld float64) {
	vs := orgNet.Variables()
	// compute complete pjoint
	pjoint := orgNet.Node(vs[0]).Potential().Copy()
	for _, v := range vs[1:] {
		pjoint.Times(orgNet.Node(v).Potential())
	}
	// compute complete qjoint
	qjoint := compNet.Nodes()[0].Potential().Copy()
	for _, nd := range compNet.Nodes()[1:] {
		qjoint.Times(nd.Potential())
	}
	kld = -floats.Sum(pjoint.Times(qjoint.Log().Minus(pjoint.Log())).Values())
	return kld
}
