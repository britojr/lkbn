package main

import (
	"fmt"

	"github.com/britojr/lkbn/data"
	"github.com/britojr/lkbn/emlearner"
	"github.com/britojr/lkbn/learner/lkm"
	"github.com/britojr/lkbn/vars"
)

func main() {
	paramLearner := emlearner.New()
	paramLearner.SetProperties(map[string]string{
		emlearner.ParmMaxIters:  "32",
		emlearner.ParmThreshold: "1e-1",
		emlearner.ParmInitIters: "1",
		emlearner.ParmRestarts:  "8",
		emlearner.ParmThreads:   "8",
	})
	fname := "test.csv"
	ds := data.NewDataset(fname)
	vs := ds.Variables()

	cases := []struct {
		cl1, cl2 []vars.VarList
		lvs      []*vars.Var
		ds       *data.Dataset
	}{{
		[]vars.VarList{
			[]*vars.Var{vs[0]}, []*vars.Var{vs[1]}, []*vars.Var{vs[2]}, []*vars.Var{vs[3]}, []*vars.Var{vs[4]},
			[]*vars.Var{vs[5]}, []*vars.Var{vs[6]}, []*vars.Var{vs[7]}, []*vars.Var{vs[8]}, []*vars.Var{vs[9]},
		},
		[]vars.VarList{
			[]*vars.Var{vs[10]}, []*vars.Var{vs[11]}, []*vars.Var{vs[12]}, []*vars.Var{vs[13]},
			[]*vars.Var{vs[14]}, []*vars.Var{vs[15]},
		},
		[]*vars.Var{vars.New(len(vs), 2, "", true), vars.New(len(vs)+1, 2, "", true)},
		ds,
	}}
	for _, tt := range cases {
		ct, lvs, _, _ := lkm.LearnLKM2L(tt.lvs, tt.cl1, tt.cl2, tt.ds, paramLearner)
		vs := ct.Variables()
		if len(vs) != len(tt.ds.Variables())+2 {
			fmt.Printf("latent variable wasn't created: %v", vs)
		}
		if lvs[0] != ct.Variables()[len(tt.ds.Variables())-1] ||
			lvs[1] != ct.Variables()[len(tt.ds.Variables())] {
			fmt.Printf("wrong return: %v != %v", lvs, ct.Variables()[len(tt.ds.Variables())-1:])
		}
		fmt.Println()
		fmt.Println(ct.String())
	}
}
