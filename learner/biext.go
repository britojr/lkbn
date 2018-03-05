package learner

import (
	"github.com/britojr/btbn/optimizer"
	"github.com/britojr/btbn/scr"
	"github.com/britojr/lkbn/factor"
	"github.com/britojr/lkbn/model"
	"github.com/britojr/lkbn/vars"
)

// BIextSearch implements the sampling strategy
type BIextSearch struct {
	*common     // common variables and methods
	bn          *model.BNet
	scoreRanker scr.Ranker
	iterAlg     optimizer.Optimizer
}

// NewBIextSearch creates a instance of this stragegy
func NewBIextSearch() Learner {
	return &BIextSearch{common: newCommon()}
}

// SetFileParameters sets properties
func (s *BIextSearch) SetFileParameters(props map[string]string) {
	s.common.SetFileParameters(props)
	if inputNetFile, ok := props[ParmInputNet]; ok {
		s.bn = model.ReadBNetXML(inputNetFile)
		lvs := s.bn.Variables().Diff(s.ds.Variables())
		for _, v := range lvs {
			v.SetLatent(true)
		}
	}
	if scoreFile, ok := props[ParmParentScores]; ok {
		s.scoreRanker = scr.CreateRanker(scr.Read(scoreFile), s.tw)
	}
	s.iterAlg = optimizer.NewIterativeSearch(s.scoreRanker)
	s.iterAlg.SetDefaultParameters()
	s.iterAlg.SetFileParameters(props)
	s.iterAlg.ValidateParameters()
}

// Search searches for a network structure
func (s *BIextSearch) Search() Solution {
	queue := []*vars.Var{findRoot(s.bn)}
	bn := model.NewBNet()
	vs := s.bn.Variables()
	for len(queue) > 0 {
		v := queue[0]
		nd := model.NewBNode(v)
		nd.SetPotential(s.bn.Node(v).Potential())
		bn.AddNode(nd)
		chs := findChildren(s.bn, v)
		for _, v := range chs {
			if v.Latent() {
				queue = append(queue, v)
				chs.Remove(v.ID())
			}
		}
		s.iterAlg.(*optimizer.IterativeSearch).SetVarsSubSet(listIDs(chs))
		bnStruct := optimizer.Search(s.iterAlg, 1, 0)
		for _, u := range chs {
			nd := model.NewBNode(u)
			bn.AddNode(nd)
			scope := vars.VarList{u}
			scope.Add(v)
			for _, id := range bnStruct.Parents(u.ID()).DumpAsInts() {
				scope.Add(vs.FindByID(id))
			}
			nd.SetPotential(factor.New(scope...))
		}
		queue = queue[1:]
	}
	// log.Printf("--------------------------------------------------\n")
	// log.Printf("LL: %.6f\n", bn.Score())
	return bn
}

func findRoot(bn *model.BNet) *vars.Var {
	for _, v := range bn.Variables() {
		if len(bn.Node(v).Parents()) == 0 {
			return v
		}
	}
	panic("no root variable")
}

func findChildren(bn *model.BNet, v *vars.Var) (chs vars.VarList) {
	for _, u := range bn.Variables() {
		if bn.Node(u).Parents().FindByID(v.ID()) != nil {
			chs.Add(u)
		}
	}
	return
}

func listIDs(vs []*vars.Var) (is []int) {
	for _, v := range vs {
		is = append(is, v.ID())
	}
	return
}
