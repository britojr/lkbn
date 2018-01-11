package scores

import (
	"math"

	"github.com/britojr/lkbn/data"
	"github.com/britojr/lkbn/model"
)

// ComputeBIC computes Bayesian Information Criterion:
// 	BIC = LL - ( model_size * ln(N)/2 )
func ComputeBIC(ct *model.CTree, ds *data.Dataset) float64 {
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
