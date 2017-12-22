package learner

import (
	"fmt"

	"github.com/britojr/btbn/scr"
	"github.com/britojr/lkbn/data"
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

// SetDataset sets dataset
func (s *BridgeSearch) SetDataset(ds *data.Dataset) {
	s.common.SetDataset(ds)
	s.mutInfo = scr.ComputeMutInfDF(ds.DataFrame())
}

// Search searches for a network structure
func (s *BridgeSearch) Search() Solution {
	// TODO: use a copy of paramLearner with 'reuse' parm set to false
	ct := model.SampleUniform(s.vs, s.tw)
	// s.paramLearner.Run(ct, s.ds.IntMaps())

	fmt.Println("Grouping variables per mutual information")
	gs := splitVarGroups(s.ds.Variables(), s.tw, s.mutInfo)
	for i, g := range gs {
		fmt.Printf("%v: %v\n", i, g)
	}
	fmt.Println("Computing mutual information per groups")
	gmi := computeGroupedMI(gs, s.mutInfo)
	for i := range gmi {
		fmt.Printf("%v\n", gmi[i])
	}
	fmt.Println("Grouping the groups of variables")
	cls := clusterGroups(gs, gmi)
	for i := range cls {
		fmt.Printf("%v: %v\n", i, cls[i])
	}

	// TODO: remove
	// learnLKM1L(, ds, paramLearner)
	return ct
}

type group struct {
	vars.VarList
	id int
}

// splits varlist in groups of size k, grouping variables by highest MI
func splitVarGroups(vs vars.VarList, k int, mutInfo *scr.MutInfo) (gs []group) {
	remain := vs.Copy()
	id := 0
	for len(remain) != 0 {
		g := highestPair(remain, mutInfo)
		for _, v := range g {
			remain.Remove(v.ID())
		}
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
		gs = append(gs, group{g, id})
		id++
	}
	return
}

// finds the highest scoring pair of variables
func highestPair(vs vars.VarList, mutInfo *scr.MutInfo) vars.VarList {
	if len(vs) < 2 {
		panic("learner: not enough variables to compute highest MI pair")
	}
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

// computes mutual information between groups of variables
func computeGroupedMI(gs []group, mutInfo *scr.MutInfo) [][]float64 {
	mat := make([][]float64, len(gs))
	for i := range mat {
		mat[i] = make([]float64, len(gs))
	}
	for i := 0; i < len(gs); i++ {
		for j := 0; j < i; j++ {
			_, maxMI := highestToGroup(gs[i].VarList, gs[j].VarList, mutInfo)
			mat[i][j], mat[j][i] = maxMI, maxMI
		}
	}
	return mat
}

func clusterGroups(gs []group, gmi [][]float64) [][]group {
	fmt.Println("Finding clusters...")
	cls := make([][]group, 0)
	for len(gs) > 0 {
		cl := highestGroupPair(gs, gmi)
		for _, g := range cl {
			groupRemove(&gs, g)
		}
		for {
			if len(gs) == 0 {
				break
			}
			g, _ := highestGroupToCluster(cl, gs, gmi)
			cl = append(cl, g)
			groupRemove(&gs, g)
			// TODO: add pseudo-lcm and pseudo-ltm2l functions here
			if len(cl) > 3 {
				break
			}
		}
		cls = append(cls, cl)
		if len(gs) == 1 {
			cls = append(cls, gs)
			gs = nil
		}
	}
	return cls
}

func highestGroupPair(gs []group, gmi [][]float64) []group {
	maxMI := 0.0
	var x, y group
	for i, v := range gs {
		for _, w := range gs[i+1:] {
			mi := gmi[v.id][w.id]
			if mi > maxMI {
				maxMI = mi
				x, y = v, w
			}
		}
	}
	return []group{x, y}
}

func highestGroupToCluster(vs, ws []group, gmi [][]float64) (group, float64) {
	maxMI := 0.0
	var x group
	for _, v := range vs {
		for _, w := range ws {
			mi := gmi[v.id][w.id]
			if mi > maxMI {
				maxMI = mi
				x = w
			}
		}
	}
	return x, maxMI
}

func groupRemove(gs *[]group, g group) {
	for j, h := range *gs {
		if h.id == g.id {
			(*gs) = append((*gs)[:j], (*gs)[j+1:]...)
			break
		}
	}
}
