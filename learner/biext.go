package learner

import (
	"log"

	"github.com/britojr/lkbn/model"
)

// BIextSearch implements the sampling strategy
type BIextSearch struct {
	*common // common variables and methods
}

// NewBIextSearch creates a instance of this stragegy
func NewBIextSearch() Learner {
	return &BIextSearch{common: newCommon()}
}

// Search searches for a network structure
func (s *BIextSearch) Search() Solution {
	ct := model.SampleUniform(s.vs, s.tw)
	s.paramLearner.Run(ct, s.ds.IntMaps())
	log.Printf("--------------------------------------------------\n")
	log.Printf("LL: %.6f\n", ct.Score())
	return ct
}
