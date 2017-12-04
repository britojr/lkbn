package learner

import (
	"github.com/britojr/btbn/scr"
	"github.com/britojr/lkbn/model"
	"github.com/britojr/lkbn/vars"
)

// SampleSearch implements the sampling strategy
type SampleSearch struct {
	*common              // common variables and methods
	mutInfo *scr.MutInfo // pre-computed mutual information matrix
}

// NewCTSampleSearch creates a instance of the sample stragegy
func NewCTSampleSearch() Learner {
	return &SampleSearch{common: newCommon()}
}

// Search searches for a network structure
func (s *SampleSearch) Search() Solution {
	ct := model.SampleUniform(s.vs, s.tw)
	s.paramLearner.Run(ct, s.ds.IntMaps())
	return ct
}

// ComputeMIScore computes linked mutual information score
// TODO: move this to 'selected strategy'
func ComputeMIScore(ct *model.CTree, mutInfo *scr.MutInfo) (mi float64) {
	m := ct.VarsNeighbors()
	for v, ne := range m {
		for _, w := range ne {
			if w.ID() < v.ID() {
				break
			}
			mi += linkMI(v, w, m, mutInfo)
		}
	}
	return
}

func linkMI(v, w *vars.Var, m map[*vars.Var]vars.VarList, mutInfo *scr.MutInfo) float64 {
	if !v.Latent() && !w.Latent() {
		return mutInfo.Get(v.ID(), w.ID())
	}
	if v.ID() > w.ID() {
		v, w = w, v
	}
	ne := m[w].Diff(m[v])
	var max float64
	for _, u := range ne {
		if u.ID() == v.ID() {
			continue
		}
		mi := linkMI(v, u, m, mutInfo)
		if mi > max {
			max = mi
		}
	}
	return max
}
