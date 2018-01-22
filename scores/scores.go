package scores

import (
	"log"
	"math"

	"github.com/britojr/lkbn/inference"
	"github.com/britojr/lkbn/model"
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
