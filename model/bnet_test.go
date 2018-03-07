package model

import (
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/britojr/lkbn/factor"
	"github.com/britojr/lkbn/vars"
	"github.com/britojr/utl/errchk"
	"github.com/britojr/utl/ioutl"
	"gonum.org/v1/gonum/floats"
)

const tol = 1e-8

var xmlbifNet1 = `<?xml version="1.0" encoding="UTF-8"?>
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
	<GIVEN>node0</GIVEN>
	<GIVEN>node1</GIVEN>
	<TABLE>0.3264 0.6736 0.3579 0.6421 0.5806 0.4194 0.5468 0.4532 </TABLE>
	<!--
	<GIVEN>node1</GIVEN>
	<GIVEN>node0</GIVEN>
	<TABLE>0.3264 0.6736 0.5806 0.4194 0.3579 0.6421 0.5468 0.4532 </TABLE>
	-->
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

func TestReadBNetXML(t *testing.T) {
	fname := ioutl.TempFile("bnet_test", xmlbifNet1)
	vs := []*vars.Var{
		vars.New(0, 2, "node0", false), vars.New(1, 2, "node1", false), vars.New(2, 2, "node2", false),
		vars.New(3, 2, "node3", false), vars.New(4, 2, "node4", false),
	}
	fs := []*factor.Factor{
		factor.New(vs[0]).SetValues([]float64{0.587, 0.413}),
		factor.New(vs[0], vs[1]).SetValues([]float64{0.8052, 0.6901, 0.1948, 0.3099}),
		factor.New(vs[0], vs[1], vs[2]).SetValues([]float64{0.3264, 0.3579, 0.5806, 0.5468, 0.6736, 0.6421, 0.4194, 0.4532}),
		factor.New(vs[0], vs[1], vs[3], vs[4]).SetValues([]float64{
			0.948, 0.0517, 0.8135, 0.7964, 0.052, 0.9483, 0.1865, 0.2036, 0.5574, 0.3416, 0.2107, 0.5683, 0.4426, 0.6584, 0.7893, 0.4317,
		}),
		factor.New(vs[0], vs[4]).SetValues([]float64{0.0249, 0.6028, 0.9751, 0.3972}),
	}
	b := ReadBNetXML(fname)
	if b == nil || len(b.nodes) == 0 {
		t.Errorf("return empty bnet: %v\n", b)
	}
	if !b.Variables().Equal(vs) {
		t.Errorf("wrong variables:\n%v\n%v\n", vs, b.Variables())
	}
	for i, v := range vs {
		if b.Node(v).Variable().Name() != v.Name() {
			t.Errorf("wrong node pivot %v != %v", b.Node(v).Variable().Name(), v.Name())
		}
		if !b.Node(v).Potential().Variables().Equal(fs[i].Variables()) {
			t.Errorf("wrong node variables\n%v\n!=\n%v\n", b.Node(v).Potential().Variables(), fs[i].Variables())
		}
		if !reflect.DeepEqual(b.Node(v).Potential().Values(), fs[i].Values()) {
			t.Errorf("wrong node values\n%v\n!=\n%v\n", b.Node(v).Potential().Values(), fs[i].Values())
		}
	}
}

func TestBNetReadWrite(t *testing.T) {
	vs := []*vars.Var{
		vars.New(0, 2, "node0", false), vars.New(1, 2, "node1", false), vars.New(2, 2, "node2", false),
		vars.New(3, 2, "node3", false), vars.New(4, 2, "node4", false),
	}
	fs := []*factor.Factor{
		factor.New(vs[0]).SetValues([]float64{0.587, 0.413}),
		factor.New(vs[0], vs[1]).SetValues([]float64{0.8052, 0.6901, 0.1948, 0.3099}),
		factor.New(vs[0], vs[1], vs[2]).SetValues([]float64{0.3264, 0.3579, 0.5806, 0.5468, 0.6736, 0.6421, 0.4194, 0.4532}),
		factor.New(vs[0], vs[1], vs[3], vs[4]).SetValues([]float64{
			0.948, 0.0517, 0.8135, 0.7964, 0.052, 0.9483, 0.1865, 0.2036, 0.5574, 0.3416, 0.2107, 0.5683, 0.4426, 0.6584, 0.7893, 0.4317,
		}),
		factor.New(vs[0], vs[4]).SetValues([]float64{0.0249, 0.6028, 0.9751, 0.3972}),
	}
	b1 := NewBNet()
	for i, f := range fs {
		nd := NewBNode(vs[i])
		nd.SetPotential(f)
		b1.AddNode(nd)
	}
	fp, err := ioutil.TempFile("", "")
	errchk.Check(err, "")
	fp.Close()

	b1.Write(fp.Name())
	b2 := ReadBNetXML(fp.Name())
	if b2 == nil {
		t.Errorf("error reading structure: got nil!\n")
	}
	if !b1.Equal(b2) {
		t.Errorf("problem saving/reading structure:\n%v\n!=\n%v\n", b1, b2)
	}
}

func TestBNEqual(t *testing.T) {
	vs := []*vars.Var{
		vars.New(0, 2, "node0", false), vars.New(1, 2, "node1", false), vars.New(2, 2, "node2", false),
		vars.New(3, 2, "node3", false), vars.New(4, 2, "node4", false),
	}
	fs := []*factor.Factor{
		factor.New(vs[0]).SetValues([]float64{0.587, 0.413}),
		factor.New(vs[0], vs[1]).SetValues([]float64{0.8052, 0.6901, 0.1948, 0.3099}),
		factor.New(vs[0], vs[1], vs[2]).SetValues([]float64{0.3264, 0.3579, 0.5806, 0.5468, 0.6736, 0.6421, 0.4194, 0.4532}),
		factor.New(vs[0], vs[1], vs[3], vs[4]).SetValues([]float64{
			0.948, 0.0517, 0.8135, 0.7964, 0.052, 0.9483, 0.1865, 0.2036, 0.5574, 0.3416, 0.2107, 0.5683, 0.4426, 0.6584, 0.7893, 0.4317,
		}),
		factor.New(vs[0], vs[4]).SetValues([]float64{0.0249, 0.6028, 0.9751, 0.3972}),
	}
	b := []*BNet{NewBNet(), NewBNet(), NewBNet(), NewBNet()}
	for i, f := range fs {
		nd := NewBNode(vs[i])
		nd.SetPotential(f.Copy())
		b[0].AddNode(nd)
	}
	for i, f := range fs {
		nd := NewBNode(vs[i])
		nd.SetPotential(f.Copy())
		b[1].AddNode(nd)
	}
	for i, f := range fs {
		nd := NewBNode(vs[i])
		nd.SetPotential(f.Copy().RandomDistribute())
		b[2].AddNode(nd)
	}
	for i := range vs {
		nd := NewBNode(vs[i])
		nd.SetPotential(factor.New(vs[:i]...))
		b[3].AddNode(nd)
	}

	if !b[0].Equal(b[1]) {
		t.Errorf("b0 should equal b1\n%v\n!=\n%v\n", b[0], b[1])
	}
	if !b[1].Equal(b[0]) {
		t.Errorf("b1 should equal b0 (symmetry)\n%v\n!=\n%v\n", b[1], b[0])
	}
	if b[1].Equal(b[2]) {
		t.Errorf("b1 should differ b2\n%v\n!=\n%v\n", b[1], b[2])
	}
	if b[1].Equal(b[3]) {
		t.Errorf("b1 should differ b3\n%v\n!=\n%v\n", b[1], b[3])
	}
	for i, bx := range b {
		if !bx.Equal(bx) {
			t.Errorf("b[%v] should be equal to itself\n%v\n", i, bx)
		}
	}
}

func TestBNetParents(t *testing.T) {
	fname := ioutl.TempFile("bnet_test", xmlbifNet1)
	vs := []*vars.Var{
		vars.New(0, 2, "node0", false), vars.New(1, 2, "node1", false), vars.New(2, 2, "node2", false),
		vars.New(3, 2, "node3", false), vars.New(4, 2, "node4", false),
	}
	parents := map[int]vars.VarList{
		0: vars.VarList{},
		1: vars.VarList{vs[0]},
		2: vars.VarList{vs[0], vs[1]},
		3: vars.VarList{vs[0], vs[1], vs[4]},
		4: vars.VarList{vs[0]},
	}
	b := ReadBNetXML(fname)
	for i, v := range b.Variables() {
		if !b.Node(v).Parents().Equal(parents[i]) {
			t.Errorf("wrong parents %v != %v\n", b.Node(v).Parents(), parents[i])
		}
	}
}

func TestMarginalizedFamily(t *testing.T) {
	fname := ioutl.TempFile("bnet_test", xmlbifNet1)
	b := ReadBNetXML(fname)
	vs := b.Variables()
	// compute the complete joint
	f := b.Node(vs[0]).Potential().Copy()
	for _, v := range vs[1:] {
		f.Times(b.Node(v).Potential())
	}
	// marginalize families
	famMargs := map[int]*factor.Factor{
		0: f.Copy().Marginalize(vs[0]),
		1: f.Copy().Marginalize(vs[0], vs[1]),
		2: f.Copy().Marginalize(vs[0], vs[1], vs[2]),
		3: f.Copy().Marginalize(vs[0], vs[1], vs[3], vs[4]),
		4: f.Copy().Marginalize(vs[0], vs[4]),
	}
	// p0134 := f.Copy().Marginalize(vs[0], vs[1], vs[3], vs[4])
	// p012, err := f.Copy().Marginalize(vs[0], vs[1], vs[2]).Normalize(vs[2])
	// errchk.Check(err, "")
	// qjoint := p0134.Copy().Times(p012)
	// fmt.Printf("\np0134: %v\n", strings.Join(conv.Sftoa(p0134.Values()), ", "))
	// fmt.Printf("\np012: %v\n", strings.Join(conv.Sftoa(p012.Values()), ", "))
	//
	// if !floats.EqualApprox(f.Values(), qjoint.Values(), tol) {
	// 	t.Fatalf("wrong joint\n%v\n!=\n%v\n", f.Values(), qjoint.Values())
	// }

	for i, v := range vs {
		got := b.MarginalizedFamily(v)
		if !famMargs[i].Variables().Equal(got.Variables()) {
			t.Errorf("wrong variables for the family of %v:\n%v\n!=\n%v\n", v, famMargs[i].Variables(), got.Variables())
		}
		if !floats.EqualApprox(famMargs[i].Values(), got.Values(), tol) {
			t.Errorf("wrong values for the family of %v:\n%v\n!=\n%v\n", v, famMargs[i].Values(), got.Values())
		}
	}
}
