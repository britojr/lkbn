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
	bn, ct := s.buildExtStructures()
	bn.SetCTree(ct)
	// TODO:
	// - learn parameters for clique tree
	// - use the learned ctree to set bnet parameters
	// - compute scores
	//
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

func (s *BIextSearch) buildExtStructures() (*model.BNet, *model.CTree) {
	queue := []*vars.Var{findRoot(s.bn)}
	bn := model.NewBNet()
	ct := model.NewCTree()
	vs := s.bn.Variables()
	for len(queue) > 0 {
		v := queue[0]
		nd := model.NewBNode(v)
		nd.SetPotential(s.bn.Node(v).Potential().Copy())
		bn.AddNode(nd)
		createCTNode(ct, nd.Potential().Copy(), nd.Potential().Variables().Diff(vars.VarList{v}))
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
		ord := bnStruct.Topological()

		fmt.Println("structure")
		fmt.Println(bnStruct)
		fmt.Println("Ordering")
		fmt.Println(ord)
		fmt.Println()

		scopes := make(map[int]vars.VarList)
		fmt.Println("\tscopes:")
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
			scopes[u.ID()] = nd.Potential().Variables()
			fmt.Printf("\t[%v]: %v\n", u.Name(), scopes[u.ID()])
		}

		fmt.Println("\ttreeNodes:")
		p := factor.New()
		for _, v := range ord[:s.Treewidth()+1] {
			p.Times(bn.Node(vs.FindByID(v)).Potential())
		}
		createCTNode(ct, p, vars.VarList{v})
		fmt.Printf("\t[r]%v\n", p.Variables())
		for _, v := range ord[s.Treewidth()+1:] {
			u := vs.FindByID(v)
			p := factor.New(scopes[u.ID()]...)
			createCTNode(ct, p, scopes[u.ID()].Diff(vars.VarList{u}))
			fmt.Printf("\t[%v]%v\n", u.Name(), p.Variables())
		}

		queue = queue[1:]
	}
	return bn, ct
}

func createCTNode(ct *model.CTree, pot *factor.Factor, parents vars.VarList) {
	nd := model.NewCTNode()
	nd.SetPotential(pot)
	pa := ct.FindNodeContaining(parents)
	if pa != nil {
		pa.AddChild(nd)
	}
	ct.AddNode(nd)
}
