package scores

import (
	"testing"

	"github.com/britojr/lkbn/model"
	"github.com/britojr/utl/floats"
	"github.com/britojr/utl/ioutl"
)

const tol = 1e-10

func TestComputeModelSize(t *testing.T) {
	struct1 := `
variables:
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
- {name: Y12,  card: 5, latent: true}
nodes:
- clqvars: [Y11, Y12]
- clqvars: [X1, X5, X7, Y12]
  parent: [Y11, Y12]
- clqvars: [X2, X3, Y12]
  parent: [Y11, Y12]
- clqvars: [X4, X6, X8, Y11]
  parent: [Y11, Y12]
- clqvars: [X0, X9, X10, Y12]
  parent: [Y11, Y12]
`
	cases := []struct {
		ct   *model.CTree
		want int
	}{{
		model.CTreeFromString(struct1),
		120,
	}}
	for _, tt := range cases {
		got := computeModelSize(tt.ct)
		if got != tt.want {
			t.Errorf("wrong model size: %v != %v", tt.want, got)
		}
	}
}

var ctreeN5K3 = `
variables:
- {name: node0,  card: 2}
- {name: node1,  card: 2}
- {name: node2,  card: 2}
- {name: node3,  card: 2}
- {name: node4,  card: 2}
nodes:
- clqvars: [node0, node1, node3, node4]
  values: [1.1157054432479998E-02, 8.882308761788003E-03, 2.3162421377399996E-03, 6.144352496990398E-02, 6.119903275199999E-04, 1.6292250287821203E-01, 5.310131022599999E-04, 1.5708063390096E-02, 2.56896382210776E-01, 3.8671336423776E-02, 2.3493122640931997E-02, 2.8890730545011997E-02, 2.0398697302922397E-01, 7.453515193622401E-02, 8.800722211906799E-02, 2.1946381094987994E-02]
  parent: []
- clqvars: [node0, node1, node2]
  values: [3.264E-01, 3.579E-01, 5.806E-01, 5.468E-01, 6.736E-01, 6.421E-01, 4.194E-01, 4.532E-01]
  parent: [node0, node1, node3, node4]
`
var bnetN5K3 = `<?xml version="1.0" encoding="UTF-8"?>
<BIF VERSION="0.3">
<NETWORK>
<NAME>InternalNetwork</NAME>
<VARIABLE TYPE="nature">
	<NAME>node0</NAME>
	<OUTCOME>state0</OUTCOME>
	<OUTCOME>state1</OUTCOME>
	<PROPERTY>position = (45, 30)</PROPERTY>
</VARIABLE>
<VARIABLE TYPE="nature">
	<NAME>node1</NAME>
	<OUTCOME>state0</OUTCOME>
	<OUTCOME>state1</OUTCOME>
	<PROPERTY>position = (15, 150)</PROPERTY>
</VARIABLE>
<VARIABLE TYPE="nature">
	<NAME>node2</NAME>
	<OUTCOME>state0</OUTCOME>
	<OUTCOME>state1</OUTCOME>
	<PROPERTY>position = (44, 270)</PROPERTY>
</VARIABLE>
<VARIABLE TYPE="nature">
	<NAME>node3</NAME>
	<OUTCOME>state0</OUTCOME>
	<OUTCOME>state1</OUTCOME>
	<PROPERTY>position = (15, 390)</PROPERTY>
</VARIABLE>
<VARIABLE TYPE="nature">
	<NAME>node4</NAME>
	<OUTCOME>state0</OUTCOME>
	<OUTCOME>state1</OUTCOME>
	<PROPERTY>position = (135, 164)</PROPERTY>
</VARIABLE>
<DEFINITION>
	<FOR>node0</FOR>
	<TABLE>0.587 0.413 </TABLE>
</DEFINITION>
<DEFINITION>
	<FOR>node1</FOR>
	<GIVEN>node0</GIVEN>
	<TABLE>0.8052 0.1948 0.6901 0.3099 </TABLE>
</DEFINITION>
<DEFINITION>
	<FOR>node2</FOR>
	<GIVEN>node1</GIVEN>
	<GIVEN>node0</GIVEN>
	<TABLE>0.3264 0.6736 0.5806 0.4194 0.3579 0.6421 0.5468 0.4532 </TABLE>
</DEFINITION>
<DEFINITION>
	<FOR>node3</FOR>
	<GIVEN>node4</GIVEN>
	<GIVEN>node0</GIVEN>
	<GIVEN>node1</GIVEN>
	<TABLE>0.948 0.052 0.5574 0.4426 0.0517 0.9483 0.3416 0.6584 0.8135 0.1865 0.2107 0.7893 0.7964 0.2036 0.5683 0.4317 </TABLE>
</DEFINITION>
<DEFINITION>
	<FOR>node4</FOR>
	<GIVEN>node0</GIVEN>
	<TABLE>0.0249 0.9751 0.6028 0.3972 </TABLE>
</DEFINITION>
</NETWORK>
</BIF>
`

func TestKLDiv(t *testing.T) {
	orgFile := ioutl.TempFile("kldiv_test", bnetN5K3)
	compFile := ioutl.TempFile("kldiv_test", ctreeN5K3)
	orgNet := model.ReadBNetXML(orgFile)
	compNet := model.ReadCTree(compFile)
	want := 0.0
	got := KLDiv(orgNet, compNet)
	if !floats.AlmostEqual(want, got, tol) {
		t.Errorf("wrong result for the same distribution: %v != %v", want, got)
	}
}

func TestKLDivBruteForce(t *testing.T) {
	orgFile := ioutl.TempFile("kldiv_test", bnetN5K3)
	compFile := ioutl.TempFile("kldiv_test", ctreeN5K3)
	orgNet := model.ReadBNetXML(orgFile)
	compNet := model.ReadCTree(compFile)
	want := 0.0
	got := KLDivBruteForce(orgNet, compNet)
	if !floats.AlmostEqual(want, got, tol) {
		t.Errorf("wrong result for the same distribution: %v != %v", want, got)
	}
}
