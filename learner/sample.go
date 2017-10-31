package learner

import "github.com/britojr/lkbn/model"

// SampleSearch implements the sampling strategy
type SampleSearch struct {
	*common // common variables and methods
}

// NewSampleSearch creates a instance of the sample stragegy
func NewSampleSearch() Learner {
	return &SampleSearch{common: newCommon()}
}

// Search searches for a network structure
func (s *SampleSearch) Search() Solution {
	// tk := ktree.UniformSample(s.nv, s.tw)
	// bn := daglearner.Approximated(tk, s.scoreRanker)
	// return bn
	return new(model.BNet)
}
