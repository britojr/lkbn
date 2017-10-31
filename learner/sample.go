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
	ct := s.sampleCTree()
	s.computeMIScore(ct)
	return ct
}

func (s *SampleSearch) sampleCTree() *model.CTree {
	return model.SampleUniform(s.vs, s.tw)
}

func (s *SampleSearch) computeMIScore(ct *model.CTree) {

}
