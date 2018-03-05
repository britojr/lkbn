package learner

import (
	"log"

	"github.com/britojr/btbn/scr"
	"github.com/britojr/lkbn/model"
)

// BIextSearch implements the sampling strategy
type BIextSearch struct {
	*common     // common variables and methods
	ct          *model.CTree
	scoreRanker scr.Ranker
}

// NewBIextSearch creates a instance of this stragegy
func NewBIextSearch() Learner {
	return &BIextSearch{common: newCommon()}
}

// SetFileParameters sets properties
func (s *BIextSearch) SetFileParameters(props map[string]string) {
	s.common.SetFileParameters(props)
	if scoreFile, ok := props[ParmParentScores]; ok {
		s.scoreRanker = scr.CreateRanker(scr.Read(scoreFile), s.tw)
	}
	if inputNetFile, ok := props[ParmInputNet]; ok {
		s.ct = model.ReadCTree(inputNetFile)
	}
}

// Search searches for a network structure
func (s *BIextSearch) Search() Solution {
	ct := model.SampleUniform(s.vs, s.tw)
	log.Printf("--------------------------------------------------\n")
	log.Printf("LL: %.6f\n", ct.Score())
	return ct
}
