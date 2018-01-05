package learner

import (
	"fmt"

	"github.com/britojr/btbn/scr"
	"github.com/britojr/lkbn/data"
	"github.com/britojr/lkbn/emlearner"
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

	fmt.Println("Grouping variables per mutual information")
	gs := groupVariables(s.ds.Variables(), s.tw, s.mutInfo)
	for i, g := range gs {
		fmt.Printf("%v: %v\n", i, g)
	}
	fmt.Println("Computing mutual information per groups")
	gpmi := computeGroupedMI(gs, s.mutInfo)
	for i := range gpmi {
		fmt.Printf("%v: %v\n", i, gpmi[i])
	}
	fmt.Println("Grouping the groups of variables")
	// TODO: use a copy of paramLearner with 'reuse' parm set to false
	cls := clusterGroups(gs, gpmi, s.ds, s.paramLearner)
	for i := range cls {
		fmt.Printf("%v: %v\n", i, cls[i])
	}

	// creates the latent variables of the model
	lvs := make([]*vars.Var, len(cls))
	// creates a subtree for each cluster
	// TODO: update here:
	// - the LKM1L needs to have the same structure pattern as 2L, that is, the root clique is fully latent
	// also, it is necessary to add an input parameter to inform the desired latent variable
	// - then learn and store all subtrees
	// - for each subtree, run inference for each data case saving the latent variable post marginal
	for i, cl := range cls {
		lvs[i] = vars.New(s.nv+i, 2, "", true)
		learnLKM1L(cl, s.ds, s.paramLearner)
	}

	// TODO: remove
	ct := model.SampleUniform(s.vs, s.tw)
	return ct
}

// splits varlist in groups of size k, grouping variables by highest MI
func groupVariables(vs vars.VarList, k int, mutInfo *scr.MutInfo) (gs []vars.VarList) {
	remain := vs.Copy()
	for len(remain) != 0 {
		g := highestPair(remain, mutInfo)
		for _, v := range g {
			remain.Remove(v.ID())
		}
		for i := 0; i < k-2; i++ {
			if len(remain) == 0 {
				break
			}
			x, _ := highestToGroup(g, remain, mutInfo)
			g.Add(x[1])
			remain.Remove(x[1].ID())
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
func highestToGroup(vs, ws vars.VarList, mutInfo *scr.MutInfo) (vars.VarList, float64) {
	maxMI := 0.0
	xs := make([]*vars.Var, 2)
	for _, v := range vs {
		for _, w := range ws {
			mi := mutInfo.Get(v.ID(), w.ID())
			if mi > maxMI {
				maxMI = mi
				xs[0] = v
				xs[1] = w
			}
		}
	}
	return xs, maxMI
}

// computes mutual information between groups of variables
func computeGroupedMI(gs []vars.VarList, mutInfo *scr.MutInfo) map[string]map[string]float64 {
	gpMI := make(map[string]map[string]float64)
	for i := range gs {
		gpki := groupKey(gs[i])
		gpMI[gpki] = make(map[string]float64)
		for j := 0; j < i; j++ {
			gpkj := groupKey(gs[j])
			_, maxMI := highestToGroup(gs[i], gs[j], mutInfo)
			gpMI[gpki][gpkj], gpMI[gpkj][gpki] = maxMI, maxMI
		}
	}
	return gpMI
}

func groupKey(gp vars.VarList) string {
	k := fmt.Sprint(gp)
	k = k[1 : len(k)-1]
	return k
}

func clusterGroups(gs []vars.VarList, gpMI map[string]map[string]float64,
	ds *data.Dataset, paramLearner emlearner.EMLearner) [][]vars.VarList {
	fmt.Println("Finding clusters...")
	cls := make([][]vars.VarList, 0)
	for len(gs) > 0 {
		cl := highestGroupPair(gs, gpMI)
		for _, g := range cl {
			groupRemove(&gs, g)
		}
		for {
			if len(gs) == 0 {
				break
			}
			cl2 := highestGroupToCluster(cl, gs, gpMI)
			groupRemove(&gs, cl2[1])
			cl1 := append([]vars.VarList(nil), cl...)
			groupRemove(&cl1, cl2[0])
			cl = append(cl, cl2[1])
			ct1L := learnLKM1L(cl, ds, paramLearner)
			ct2L, gs1, gs2 := learnLKM2L(cl1, cl2, ds, paramLearner)
			fmt.Println("-----------------------------------")
			fmt.Printf("b1: %v\tb2: %v\n", ct1L.BIC(), ct2L.BIC())
			fmt.Printf("cl:\n%v(%v)\n", cl, len(cl))
			fmt.Printf("cl1:\n%v(%v)\n", cl1, len(cl1))
			fmt.Printf("cl2:\n%v(%v)\n", cl2, len(cl2))
			fmt.Printf("ct1L:\n%v\n", ct1L.Nodes())
			fmt.Println()
			fmt.Printf("ct2L:\n%v\n", ct2L.Nodes())
			fmt.Println()

			if ct2L.BIC()-ct1L.BIC() > bicThreshold {
				// if fails the test, should keep the group that contains the highest pair
				if groupContains(gs1, cl[0]) {
					if groupContains(gs1, cl[1]) {
						cl = gs1
						gs = append(gs, gs2...)
						break
					}
				} else {
					if groupContains(gs2, cl[1]) {
						cl = gs2
						gs = append(gs, gs1...)
						break
					}
				}
				if len(gs1) > len(gs2) {
					cl = gs1
					gs = append(gs, gs2...)
				} else {
					cl = gs2
					gs = append(gs, gs1...)
				}
				break
			}
		}
		cls = append(cls, cl)
		if len(gs) == 1 {
			// TODO: think what to do when just one group is left alone
			cls = append(cls, gs)
			gs = nil
		}
	}
	return cls
}

// finds the highest scoring pair of groups
func highestGroupPair(gs []vars.VarList, gpMI map[string]map[string]float64) []vars.VarList {
	maxMI := 0.0
	xs := make([]vars.VarList, 2)
	for i, v := range gs {
		gpki := groupKey(v)
		for _, w := range gs[:i] {
			gpkj := groupKey(w)
			mi := gpMI[gpki][gpkj]
			if mi > maxMI {
				maxMI = mi
				xs[0], xs[1] = v, w
			}
		}
	}
	return xs
}

// finds the highest mi scoring pair of groups between two distinct group lists
func highestGroupToCluster(vs, ws []vars.VarList, gpMI map[string]map[string]float64) []vars.VarList {
	maxMI := 0.0
	xs := make([]vars.VarList, 2)
	for _, v := range vs {
		gpki := groupKey(v)
		for _, w := range ws {
			gpkj := groupKey(w)
			mi := gpMI[gpki][gpkj]
			if mi > maxMI {
				maxMI = mi
				xs[0], xs[1] = v, w
			}
		}
	}
	return xs
}

func groupRemove(gs *[]vars.VarList, g vars.VarList) {
	for i, v := range *gs {
		if g.Equal(v) {
			(*gs) = append((*gs)[:i], (*gs)[i+1:]...)
			break
		}
	}
}

func groupContains(gs []vars.VarList, g vars.VarList) bool {
	for _, v := range gs {
		if g.Equal(v) {
			return true
		}
	}
	return false
}
