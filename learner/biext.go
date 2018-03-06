package learner

import (
	"fmt"

	"github.com/britojr/btbn/optimizer"
	"github.com/britojr/btbn/scr"
	"github.com/britojr/lkbn/factor"
	"github.com/britojr/lkbn/model"
	"github.com/britojr/lkbn/vars"
	"github.com/britojr/utl/conv"
)

// BIextSearch implements the sampling strategy
type BIextSearch struct {
	*common     // common variables and methods
	bn          *model.BNet
	scoreRanker scr.Ranker
	iterAlg     optimizer.Optimizer

	maxTimeCl int
	maxIterCl int

	props map[string]string // save parameters map
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
	s.props = props
	// s.initOptimizer()
	if maxTimeCl, ok := props[ParmMaxTimeCluster]; ok {
		s.maxTimeCl = conv.Atoi(maxTimeCl)
	}
	if maxIterCl, ok := props[ParmMaxIterCluster]; ok {
		s.maxIterCl = conv.Atoi(maxIterCl)
	}
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
		fmt.Println(chs)
		s.initOptimizer() // TODO: create a new optimizer because unwanted colateral from timeout
		s.iterAlg.(*optimizer.IterativeSearch).SetVarsSubSet(listIDs(chs))
		bnStruct := optimizer.Search(s.iterAlg, s.maxIterCl, s.maxTimeCl)
		for _, u := range chs {
			nd := model.NewBNode(u)
			parents := bnStruct.Parents(u.ID()).DumpAsInts()
			scope := vars.VarList{u}
			scope.Add(v)
			for _, id := range parents {
				scope.Add(vs.FindByID(id))
			}
			nd.SetPotential(factor.New(scope...))
			bn.AddNode(nd)
		}
		queue = queue[1:]
	}
	// log.Printf("--------------------------------------------------\n")
	// log.Printf("LL: %.6f\n", bn.Score())
	return bn
}

func (s *BIextSearch) initOptimizer() {
	s.iterAlg = optimizer.NewIterativeSearch(s.scoreRanker)
	s.iterAlg.SetDefaultParameters()
	s.iterAlg.SetFileParameters(s.props)
	s.iterAlg.ValidateParameters()
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

func listIDs(vs []*vars.Var) []int {
	is := make([]int, 0, len(vs))
	for _, v := range vs {
		is = append(is, v.ID())
	}
	return is
}
