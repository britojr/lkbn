package learner

import (
	"fmt"
	"log"
	"math"

	"github.com/britojr/btbn/scr"
	"github.com/britojr/lkbn/data"
	"github.com/britojr/lkbn/emlearner"
	"github.com/britojr/lkbn/factor"
	"github.com/britojr/lkbn/graph"
	"github.com/britojr/lkbn/inference"
	"github.com/britojr/lkbn/learner/lkm"
	"github.com/britojr/lkbn/model"
	"github.com/britojr/lkbn/scores"
	"github.com/britojr/lkbn/vars"
)

// BridgeSearch implements the bridge strategy
type BridgeSearch struct {
	*common              // common variables and methods
	mutInfo *scr.MutInfo // pre-computed mutual information matrix

	localLearner emlearner.EMLearner // local parameter learner
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

// SetFileParameters sets properties
func (s *BridgeSearch) SetFileParameters(props map[string]string) {
	s.common.SetFileParameters(props)
	// use a copy of paramLearner with 'reuse' parm set to false
	s.localLearner = emlearner.New()
	s.localLearner.SetProperties(props)
	s.localLearner.SetProperties(map[string]string{emlearner.ParmReuse: "false"})
}

// Search searches for a network structure
func (s *BridgeSearch) Search() Solution {

	log.Println("Grouping variables per mutual information")
	gs := groupVariables(s.ds.Variables(), s.tw, s.mutInfo)

	log.Println("Computing mutual information between groups")
	gpmi := computeGroupedMI(gs, s.mutInfo)

	log.Println("Creating clusters of groups")
	cls := clusterGroups(gs, gpmi, s.ds, s.localLearner)

	log.Println("Learning a subtree for each cluster")
	lvs, subtrees := createSubtrees(cls, s.ds, s.localLearner)

	log.Println("Connecting subtrees")
	ct := buildConnectedTree(lvs, subtrees, s.ds)

	log.Println("Learning parameters for the full model")
	ct, _, _ = s.localLearner.Run(ct, s.ds.IntMaps())

	ct.SetBIC(scores.ComputeBIC(ct, s.ds.IntMaps()))
	log.Printf("BIC: %v\n", ct.BIC())
	fmt.Printf("Final:\n%v\n", ct.Nodes())
	return ct
}

type mutInfCalc interface {
	Get(i, j int) float64
}
type dataset interface {
	Variables() vars.VarList
	IntMaps() []map[int]int
}

// splits varlist in groups of size k, grouping variables by highest MI
func groupVariables(vs vars.VarList, k int, mutInfo mutInfCalc) (gs []vars.VarList) {
	// create groups of size one, for k=1
	if k < 2 {
		for _, v := range vs {
			gs = append(gs, []*vars.Var{v})
		}
		return
	}
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
		gs = append(gs, g)
		// if just one remais, creates a group of one
		if len(remain) == 1 {
			gs = append(gs, []*vars.Var{remain[0]})
			remain.Remove(remain[0].ID())
		}
	}
	return
}

// finds the highest scoring pair of variables
func highestPair(vs vars.VarList, mutInfo mutInfCalc) vars.VarList {
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
func highestToGroup(vs, ws vars.VarList, mutInfo mutInfCalc) ([]*vars.Var, float64) {
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
func computeGroupedMI(gs []vars.VarList, mutInfo mutInfCalc) map[string]map[string]float64 {
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
	ds dataset, paramLearner emlearner.EMLearner) [][]vars.VarList {
	nstates := 2
	lvs := []*vars.Var{
		vars.New(len(ds.Variables()), nstates, "", true),
		vars.New(len(ds.Variables())+1, nstates, "", true),
	}
	cls := make([][]vars.VarList, 0)
	for len(gs) > 0 {
		cl := highestGroupPair(gs, gpMI)
		groupRemove(&gs, cl...)
		for len(gs) != 0 {
			cl1, cl2 := increaseCluster(&cl, &gs, gpMI)
			if len(cl) < 4 {
				// can't improve much by spliting a group of size lesser than 4
				continue
			}
			ct1L, _ := lkm.LearnLKM1L(cl, lvs[0], ds, paramLearner)
			ct2L, _, gs1, gs2 := lkm.LearnLKM2L(lvs, cl1, cl2, ds, paramLearner)

			fmt.Println("-----------------------------------------------------")
			fmt.Printf("b1: %v\tb2: %v\n", ct1L.BIC(), ct2L.BIC())
			fmt.Printf("cl:\n%v(%v)\n", cl, len(cl))
			fmt.Printf("cl1:\n%v(%v)\n", cl1, len(cl1))
			fmt.Printf("cl2:\n%v(%v)\n", cl2, len(cl2))
			fmt.Printf("ct1L:\n%v\n", ct1L.Nodes())
			fmt.Println()
			fmt.Printf("ct2L:\n%v\n", ct2L.Nodes())
			fmt.Println()

			if ct2L.BIC()-ct1L.BIC() > lkm.BicThreshold {
				if chooseGroupOne(cl, gs1, gs2) {
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

// adds one variable to cluster cl, returns cl and its two partitions:
// 		- cl2 contains the highest pair between cl and gs
// 		- cl1 the remaining variables already in cl
func increaseCluster(cl, gs *[]vars.VarList,
	gpMI map[string]map[string]float64) (cl1, cl2 []vars.VarList) {
	cl2 = highestGroupToCluster(*cl, *gs, gpMI)
	groupRemove(gs, cl2[1])
	cl1 = append([]vars.VarList(nil), *cl...)
	groupRemove(&cl1, cl2[0])
	*cl = append(*cl, cl2[1])
	return
}

// returns true for choosing the group one, and false otherwise
// if fails the test, should choose the group that contains the initial highest pair
// in case the pair was split, should choose the highest group
func chooseGroupOne(cl, gs1, gs2 []vars.VarList) bool {
	if groupContains(gs1, cl[0]) {
		if groupContains(gs1, cl[1]) {
			return true
		}
	} else {
		if groupContains(gs2, cl[1]) {
			return false
		}
	}
	if len(gs1) > len(gs2) {
		return true
	}
	return false
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

func groupRemove(gs *[]vars.VarList, gv ...vars.VarList) {
	for _, g := range gv {
		for i, v := range *gs {
			if g.Equal(v) {
				(*gs) = append((*gs)[:i], (*gs)[i+1:]...)
				break
			}
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

func createSubtrees(cls [][]vars.VarList,
	ds dataset, paramLearner emlearner.EMLearner) ([]*vars.Var, []*model.CTree) {
	// creates the latent variables of the model
	lvs := make([]*vars.Var, len(cls))
	subtrees := make([]*model.CTree, len(cls))
	n := len(ds.Variables())
	// creates a subtree for each cluster
	for i, cl := range cls {
		lvs[i] = vars.New(n+i, 2, "", true)
		subtrees[i], lvs[i] = lkm.LearnLKM1L(cl, lvs[i], ds, paramLearner)
	}
	return lvs, subtrees
}

func buildConnectedTree(lvs vars.VarList, subtrees []*model.CTree, ds dataset) *model.CTree {
	nodes, edges := createFullGraph(lvs, subtrees, ds)
	// select the edges corresponding to a Max Spanning Tree on the latent variables
	// and order them in bfs order starting from root
	edges = graph.RootedTree(0, graph.MaxSpanningTree(nodes, edges))
	ct := connectSubtrees(edges, subtrees)
	return ct
}

// create a full graph, with MI as weight
func createFullGraph(lvs vars.VarList, subtrees []*model.CTree, ds dataset) ([]int, []graph.WEdge) {
	lvPosts := computeAllLatentPosts(subtrees, ds)
	var edges []graph.WEdge
	nodes := make([]int, len(lvs))
	for i := range lvs {
		nodes[i] = i
		for j := 0; j < i; j++ {
			dist := computeLatentDist(lvs[i], lvs[j], ds, lvPosts)
			mi := computePairwiseMI(dist)
			edges = append(edges, graph.WEdge{
				Head: i, Tail: j, Weight: mi,
			})
		}
	}
	return nodes, edges
}

// connect the subtrees using the first one as root
func connectSubtrees(edges []graph.WEdge, subtrees []*model.CTree) *model.CTree {
	for _, e := range edges {
		i, j := e.Head, e.Tail
		subtrees[i].Root().AddChildren(subtrees[j].Root())
		// add the parent variable to the child clique
		vi := subtrees[i].Root().Variables()
		vj := subtrees[j].Root().Variables()
		if subtrees[i].Root().Parent() != nil {
			vi = vi.Diff(subtrees[i].Root().Parent().Variables())
		}
		newScope := vj.Union(vi)
		subtrees[j].Root().SetPotential(factor.New(newScope...))
	}
	ct := subtrees[0]
	ct.BfsNodes()
	return ct
}

func computeAllLatentPosts(subtrees []*model.CTree, ds dataset) map[int][]*factor.Factor {
	lvPosts := make(map[int][]*factor.Factor)
	for _, st := range subtrees {
		lvID := st.Root().Variables()[0].ID()
		lvPosts[lvID] = computeLatentPosterior(st, ds)
	}
	return lvPosts
}

// computes the posterior distribution of the latent variable for each line of the dataset
func computeLatentPosterior(subtree *model.CTree, ds dataset) []*factor.Factor {
	lvPost := make([]*factor.Factor, len(ds.IntMaps()))
	infalg := inference.NewCTreeCalibration(subtree)
	for i, evid := range ds.IntMaps() {
		infalg.Run(evid)
		lvPost[i] = infalg.CTree().Root().Potential().Copy()
	}
	return lvPost
}

// computes joint distribution of two latent variables based on their posterior distributions
func computeLatentDist(x, y *vars.Var, ds dataset,
	lvPosts map[int][]*factor.Factor) *factor.Factor {
	// P(x, y|d) = P(x|d) * P(y|d)
	lvsDist := lvPosts[x.ID()][0].Copy().Times(lvPosts[y.ID()][0])
	for i := 1; i < len(ds.IntMaps()); i++ {
		lvsDist.Plus(lvPosts[x.ID()][i].Copy().Times(lvPosts[y.ID()][i]))
	}
	lvsDist.Normalize()
	return lvsDist
}

func computePairwiseMI(f *factor.Factor) (mi float64) {
	// marginals
	fx := f.Copy().SumOut(f.Variables()[0]).Values()
	fy := f.Copy().SumOut(f.Variables()[1]).Values()
	// I(X;Y) = sum_X,Y P(X,Y) log P(X,Y)/P(X)P(Y)
	i := 0
	for _, px := range fx {
		for _, py := range fy {
			pxy := f.Values()[i]
			i++
			if pxy != 0 {
				mi += pxy * math.Log(pxy/(px*py))
			}
		}
	}
	return
}
