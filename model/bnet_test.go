package model

import (
	"reflect"
	"testing"

	"github.com/britojr/lkbn/factor"
	"github.com/britojr/lkbn/vars"
	"github.com/britojr/utl/ioutl"
)

func TestReadBNetXML(t *testing.T) {
	content := `<?xml version="1.0" encoding="UTF-8"?>
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
	fname := ioutl.TempFile("bnet_test", content)
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
	for i, v := range b.Variables() {
		if b.Node(v).Variable().Name() != v.Name() {
			t.Errorf("wrong node pivot %v != %v", b.Node(v).Variable().Name(), v.Name())
		}
		if !b.Node(v).Potential().Variables().Equal(fs[i].Variables()) {
			t.Errorf("wrong node variables %v != %v", b.Node(v).Potential().Variables(), fs[i].Variables())
		}
		if !reflect.DeepEqual(b.Node(v).Potential().Values(), fs[i].Values()) {
			t.Errorf("wrong node values %v != %v", b.Node(v).Potential().Values(), fs[i].Values())
		}
	}
}
