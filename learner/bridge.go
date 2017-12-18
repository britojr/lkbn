package learner

import (
	"fmt"

	"github.com/britojr/btbn/scr"
	"github.com/britojr/lkbn/data"
	"github.com/britojr/lkbn/emlearner"
	"github.com/britojr/lkbn/factor"
	"github.com/britojr/lkbn/model"
	"github.com/britojr/lkbn/vars"
)

// BridgeSearch implements the bridge strategy
type BridgeSearch struct {
	*common              // common variables and methods
	mutInfo *scr.MutInfo // pre-computed mutual information matrix
}

// NewCTBridgeSearch creates a instance of this stragegy
func NewCTBridgeSearch() Learner {
	return &BridgeSearch{common: newCommon()}
}

// Search searches for a network structure
func (s *BridgeSearch) Search() Solution {
	ct := model.SampleUniform(s.vs, s.tw)
	// s.paramLearner.Run(ct, s.ds.IntMaps())

	fmt.Println("Grouping variables per mutual information")
	gs := splitGroups(s.ds.Variables(), s.tw, s.mutInfo)
	for i, g := range gs {
		fmt.Printf("%v: %v\n", i, g)
	}
	fmt.Println("Computing mutual information per groups")
	gmi := groupMI(gs, s.mutInfo)
	for i := range gmi {
		fmt.Printf("%v\n", gmi[i])
	}

	fmt.Println("Finding clusters...")
	gids := make([]int, len(gs))
	for i := range gs {
		gids[i] = i
	}
	cls := make([][]int, 0)
	for len(gids) > 0 {
		cluster := highestGroupPair(gids, gmi)
		intsRemove(&gids, cluster[0])
		intsRemove(&gids, cluster[1])
		for {
			if len(gids) == 0 {
				break
			}
			x, _ := highestGroupToCluster(cluster, gids, gmi)
			cluster = append(cluster, x)
			intsRemove(&gids, x)
			// TODO: add pseudo-lcm and pseudo-ltm2l functions here
			if len(cluster) > 3 {
				break
			}
		}
		cls = append(cls, cluster)
		if len(gids) == 1 {
			cls = append(cls, []int{gids[0]})
			intsRemove(&gids, gids[0])
		}
	}

	for i := range cls {
		fmt.Printf("%v: %v\n", i, cls[i])
	}
	return ct
}

// SetDataset sets dataset
func (s *BridgeSearch) SetDataset(ds *data.Dataset) {
	s.common.SetDataset(ds)
	s.mutInfo = scr.ComputeMutInfDF(ds.DataFrame())
}

func splitGroups(vs vars.VarList, k int, mutInfo *scr.MutInfo) (gs []vars.VarList) {
	remain := vs.Copy()
	for len(remain) != 0 {
		g := highestPair(remain, mutInfo)
		remain.Remove(g[0].ID())
		remain.Remove(g[1].ID())
		for i := 2; i < k; i++ {
			if len(remain) == 0 {
				break
			}
			x, _ := highestToGroup(g, remain, mutInfo)
			g.Add(x)
			remain.Remove(x.ID())
		}
		if len(remain) == 1 {
			g.Add(remain[0])
			remain.Remove(remain[0].ID())
		}
		gs = append(gs, g)
	}
	return
}

// finds the highest scoring pair of variables
func highestPair(vs vars.VarList, mutInfo *scr.MutInfo) vars.VarList {
	maxMI := 0.0
	var x, y *vars.Var
	for i, v := range vs {
		for _, w := range vs[i+1:] {
			mi := mutInfo.Get(v.ID(), w.ID())
			if mi > maxMI {
				maxMI = mi
				x, y = v, w
			}
		}
	}
	return []*vars.Var{x, y}
}

// finds the highest mi scoring var with relation to another group of variables
func highestToGroup(vs, ws vars.VarList, mutInfo *scr.MutInfo) (*vars.Var, float64) {
	maxMI := 0.0
	var x *vars.Var
	for _, v := range vs {
		for _, w := range ws {
			mi := mutInfo.Get(v.ID(), w.ID())
			if mi > maxMI {
				maxMI = mi
				x = w
			}
		}
	}
	return x, maxMI
}

// computes mi between groups of variables
func groupMI(gs []vars.VarList, mutInfo *scr.MutInfo) [][]float64 {
	mat := make([][]float64, len(gs))
	for i := range mat {
		mat[i] = make([]float64, len(gs))
	}

	for i := 0; i < len(gs); i++ {
		for j := 0; j < i; j++ {
			_, maxMI := highestToGroup(gs[i], gs[j], mutInfo)
			mat[i][j], mat[j][i] = maxMI, maxMI
		}
	}
	return mat
}

func highestGroupPair(gids []int, gmi [][]float64) []int {
	maxMI := 0.0
	var x, y int
	for i, v := range gids {
		for _, w := range gids[i+1:] {
			mi := gmi[v][w]
			if mi > maxMI {
				maxMI = mi
				x, y = v, w
			}
		}
	}
	return []int{x, y}
}

func highestGroupToCluster(vs, ws []int, gmi [][]float64) (int, float64) {
	maxMI := 0.0
	var x int
	for _, v := range vs {
		for _, w := range ws {
			mi := gmi[v][w]
			if mi > maxMI {
				maxMI = mi
				x = w
			}
		}
	}
	return x, maxMI
}

func intsRemove(xs *[]int, y int) {
	for j, x := range *xs {
		if x == y {
			(*xs) = append((*xs)[:j], (*xs)[j+1:]...)
			break
		}
	}
}

func computeBIC(ct *model.CTree) float64 {
	// TODO: check this bic computation
	numparms := 0
	for _, nd := range ct.Nodes() {
		numparms += len(nd.Potential().Values())
	}
	return ct.Score() - float64(numparms)
}

func learnLKM1L(gs []vars.VarList, ds *data.Dataset, paramLearner emlearner.EMLearner) *model.CTree {
	// create new latent variable
	nstate := 2
	lv := vars.New(len(ds.Variables()), nstate, "", true)
	// mount structure
	f := factor.New(gs[0].Union([]*vars.Var{lv})...)
	ct := model.NewCTree()
	root := model.NewCTNode()
	root.SetPotential(f)
	ct.AddNode(root)
	for _, g := range gs[1:] {
		f := factor.New(g.Union([]*vars.Var{lv})...)
		nd := model.NewCTNode()
		nd.SetPotential(f)
		root.AddChildren(nd)
		ct.AddNode(nd)
	}

	// increase latent cardinality and learn parameters until bic stops increasing
	ct, _, _ = paramLearner.Run(ct, ds.IntMaps())
	bic := computeBIC(ct)
	for {
		nstate++
		lv.SetNState(nstate)
		ct, _, _ = paramLearner.Run(ct, ds.IntMaps())
		newbic := computeBIC(ct)
		if newbic <= bic {
			break
		}
		bic = newbic
	}
	return ct
}

func learnLKM2L(gs []vars.VarList, ds *data.Dataset, paramLearner emlearner.EMLearner) (*model.CTree, int) {
	// TODO: implement correct 2L
	// create new latent variable
	nstate := 2
	lvs := []*vars.Var{
		vars.New(len(ds.Variables()), nstate, "", true),
		vars.New(len(ds.Variables())+1, nstate, "", true),
	}
	// mount structure
	f := factor.New(gs[0].Union([]*vars.Var{lvs[0]})...)
	ct := model.NewCTree()
	root := model.NewCTNode()
	root.SetPotential(f)
	ct.AddNode(root)
	for _, g := range gs[1:] {
		f := factor.New(g.Union([]*vars.Var{lvs[0]})...)
		nd := model.NewCTNode()
		nd.SetPotential(f)
		root.AddChildren(nd)
		ct.AddNode(nd)
	}

	// increase latent cardinality and learn parameters until bic stops increasing
	ct, _, _ = paramLearner.Run(ct, ds.IntMaps())
	bic := computeBIC(ct)
	for {
		nstate++
		lvs[0].SetNState(nstate)
		ct, _, _ = paramLearner.Run(ct, ds.IntMaps())
		newbic := computeBIC(ct)
		if newbic <= bic {
			break
		}
		bic = newbic
	}
	return ct, len(lvs)
}
