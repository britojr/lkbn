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

// Search searches for a network structure
func (s *BridgeSearch) Search() Solution {
	ct := model.SampleUniform(s.vs, s.tw)
	// s.paramLearner.Run(ct, s.ds.IntMaps())

	fmt.Println("Grouping variables per mutual information")
	gs := splitGroups(s.ds.Variables(), s.tw, s.mutInfo)
	for i, g := range gs {
		fmt.Printf("%v: %v\n", i, g)
	}
	gmi := groupMI(gs, s.mutInfo)
	for i := range gmi {
		fmt.Printf("%v\n", gmi[i])
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
