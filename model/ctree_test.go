package model

import (
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/britojr/lkbn/vars"
	"github.com/britojr/utl/errchk"
)

func TestReadWrite(t *testing.T) {
	cases := []struct {
		vs vars.VarList
		k  int
	}{{
		vars.NewList([]int{0, 1, 2, 3, 4}, []int{2, 2, 2, 2, 2}), 2,
	}, {
		vars.NewList([]int{0, 1, 2, 3, 4, 5}, []int{2, 2, 2, 2, 2, 2}), 3,
	}}
	for _, tt := range cases {
		cta := SampleUniform(tt.vs, tt.k)
		// fmt.Println(cta)
		n := len(cta.VarsNeighbors())
		if n != len(tt.vs) {
			t.Errorf("sample wrong number of variables %v != %v", len(tt.vs), n)
		}
		// for _, nd := range cta.Nodes() {
		// 	nd.Potential().RandomDistribute()
		// }
		f, err := ioutil.TempFile("", "")
		errchk.Check(err, "")
		f.Close()
		cta.Write(f.Name())
		ctb := ReadCTree(f.Name())
		// fmt.Println(ctb)

		queue := []*CTNode{cta.Root()}
		for len(queue) > 0 {
			pa := queue[0]
			queue = queue[1:]
			nd := ctb.FindNode(pa.Variables())
			if nd == nil {
				t.Errorf("can't find node %v", pa.Variables())
			} else {
				if !nd.Potential().Variables().Equal(pa.Potential().Variables()) {
					t.Errorf("potentials with different variables %v != %v",
						nd.Potential().Variables(), pa.Potential().Variables(),
					)
				}
				if !reflect.DeepEqual(nd.Potential().Values(), pa.Potential().Values()) {
					t.Errorf("potentials with different values %v != %v",
						nd.Potential().Values(), pa.Potential().Values(),
					)
				}
			}
			for _, ch := range pa.Children() {
				nd := ctb.FindNode(ch.Variables())
				if nd == nil {
					t.Errorf("can't find children %v", ch.Variables())
				}
			}
			queue = append(queue, pa.Children()...)
		}
	}
}

func TestCopy(t *testing.T) {
	cases := []struct {
		vs vars.VarList
		k  int
	}{{
		vars.NewList([]int{0, 1, 2, 3, 4}, []int{2, 2, 2, 2, 2}), 2,
	}, {
		vars.NewList([]int{0, 1, 2, 3, 4, 5}, []int{2, 2, 2, 2, 2, 2}), 3,
	}}
	for _, tt := range cases {
		cta := SampleUniform(tt.vs, tt.k)
		ctb := cta.Copy()
		mcta, mctb := cta.VarsNeighbors(), ctb.VarsNeighbors()
		for v, vs := range mcta {
			if !mctb[v].Equal(vs) {
				t.Errorf("wrong neighbors of %v: %v != %v", v, vs, mctb[v])
			}
		}
		// if !cta.Equal(ctb) {
		// 	t.Errorf("different copy!\n%v\n!=\n%v\n", cta, ctb)
		// }
	}
}
