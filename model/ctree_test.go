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
		cta.WriteYAML(f.Name())
		ctb := ReadCTreeYAML(f.Name())
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

func TestVariables(t *testing.T) {
	cases := []struct {
		vs vars.VarList
		k  int
	}{{
		vars.NewList([]int{0, 1, 2, 3, 4}, []int{2, 2, 2, 2, 2}), 2,
	}, {
		vars.NewList([]int{0, 1, 2, 3, 4, 5}, []int{2, 2, 2, 2, 2, 2}), 3,
	}}
	for _, tt := range cases {
		ct := SampleUniform(tt.vs, tt.k)
		got := ct.Variables()
		if !tt.vs.Equal(got) {
			t.Errorf("wrong variables %v != %v", tt.vs, got)
		}
	}
}

func TestEqualStruct(t *testing.T) {
	ct1 := `variables:
- {name: X0,  card: 2}
- {name: X1,  card: 2}
- {name: X2,  card: 2}
- {name: X3,  card: 2}
- {name: X4,  card: 2}
- {name: X5,  card: 2}
- {name: X6,  card: 2}
- {name: X7,  card: 2}
- {name: X8,  card: 2}
- {name: X9,  card: 2}
- {name: X10, card: 2}
- {name: Y11,  card: 3, latent: true}
nodes:
- clqvars: [X1, X5, X7, Y11]
- clqvars: [X2, X3, Y11]
  parent: [X1, X5, X7, Y11]
- clqvars: [X4, X6, X8, Y11]
  parent: [X1, X5, X7, Y11]
- clqvars: [X0, X9, X10, Y11]
  parent: [X1, X5, X7, Y11]
`
	ct2 := `variables:
- {name: X0,  card: 2}
- {name: X1,  card: 2}
- {name: X2,  card: 2}
- {name: X3,  card: 2}
- {name: X4,  card: 2}
- {name: X5,  card: 2}
- {name: X6,  card: 2}
- {name: X7,  card: 2}
- {name: X8,  card: 2}
- {name: X9,  card: 2}
- {name: X10, card: 2}
- {name: Y11,  card: 3, latent: true}
nodes:
- clqvars: [X1, X5, X7, Y11]
- clqvars: [X1, X2, X3, Y11]
  parent: [X1, X5, X7, Y11]
- clqvars: [X4, X6, X8, Y11]
  parent: [X1, X5, X7, Y11]
- clqvars: [X0, X9, X10, Y11]
  parent: [X1, X5, X7, Y11]
`
	cases := []struct {
		cta, ctb *CTree
		equal    bool
	}{
		{CTreeFromString(ct1), CTreeFromString(ct1), true},
		{CTreeFromString(ct1), CTreeFromString(ct2), false},
	}
	for _, tt := range cases {
		got := tt.cta.EqualStruct(tt.ctb)
		if got != tt.equal {
			t.Errorf("wrong comp, expect: %v got: %v", tt.equal, got)
		}
	}
}

func TestEqual(t *testing.T) {
	ct1 := `variables:
- {name: A,  card: 2}
- {name: B,  card: 2}
- {name: C,  card: 2}
- {name: D,  card: 2}
- {name: E,  card: 2}
nodes:
- clqvars: [A,B,C]
  values: [0,0,0,0,1,1,1,1]
- clqvars: [A,B,D]
  values: [1,1,0,0,1,1,1,1]
  parent: [A,B,C]
- clqvars: [B,C,E]
  values: [0,0,1,1,1,1,1,1]
  parent: [A,B,C]
`
	ct2 := `variables:
- {name: A,  card: 2}
- {name: B,  card: 2}
- {name: C,  card: 2}
- {name: D,  card: 2}
- {name: E,  card: 2}
nodes:
- clqvars: [A,B,C]
  values: [1,1,1,1,1,1,1,1]
- clqvars: [A,B,D]
  values: [1,1,0,0,1,1,1,1]
  parent: [A,B,C]
- clqvars: [B,C,E]
  values: [0,0,1,1,1,1,1,1]
  parent: [A,B,C]
`
	ct3 := `variables:
- {name: A,  card: 2}
- {name: B,  card: 2}
- {name: C,  card: 2}
- {name: D,  card: 2}
- {name: E,  card: 2}
nodes:
- clqvars: [A,B,C]
  values: [1,1,1,1,1,1,1,1]
- clqvars: [A,B,D]
  values: [1,1,0,0,1,1,1,1]
  parent: [A,B,C]
- clqvars: [A,C,E]
  values: [0,0,1,1,1,1,1,1]
  parent: [A,B,C]
`
	cases := []struct {
		cta, ctb *CTree
		equal    bool
	}{
		{CTreeFromString(ct1), CTreeFromString(ct1), true},
		{CTreeFromString(ct1), CTreeFromString(ct2), false},
		{CTreeFromString(ct2), CTreeFromString(ct3), false},
	}
	for _, tt := range cases {
		got := tt.cta.Equal(tt.ctb)
		if got != tt.equal {
			t.Errorf("wrong comp, expect: %v got: %v", tt.equal, got)
		}
	}
}

func TestReadWriteXML(t *testing.T) {
	cases := []struct {
		vs vars.VarList
		k  int
	}{{
		vars.NewList([]int{0, 1, 2, 3, 4}, []int{2, 2, 2, 2, 2}), 2,
	}, {
		vars.NewList([]int{0, 1, 2, 3, 4, 5}, []int{2, 2, 2, 2, 2, 2}), 3,
	}, {
		vars.NewList(
			[]int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19},
			[]int{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2}), 5,
	}}
	for _, tt := range cases {
		cta := SampleUniform(tt.vs, tt.k)
		for _, nd := range cta.Nodes() {
			nd.Potential().RandomDistribute()
		}
		f, err := ioutil.TempFile("", "")
		errchk.Check(err, "")
		f.Close()
		cta.Write(f.Name())
		ctb := ReadCTreeXML(f.Name())

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
