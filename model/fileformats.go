package model

import (
	"encoding/xml"
	"io/ioutil"

	"github.com/britojr/utl/errchk"
	"github.com/britojr/utl/ioutl"
)

// XMLBIF defines xmlbif structure
type XMLBIF struct {
	BNetXML  Network `xml:"NETWORK"`
	CTreeXML Network `xml:"CTREE"`
}

// Network defines network in xmlbif pattern
type Network struct {
	Name      string     `xml:"NAME"`
	Variables []Variable `xml:"VARIABLE"`
	Probs     []Prob     `xml:"DEFINITION"`
}

// Variable a variable in xmlbif
type Variable struct {
	Name   string   `xml:"NAME"`
	States []string `xml:"OUTCOME"`
}

// Prob conditional probability in xmlbif
type Prob struct {
	For   []string `xml:"FOR"`
	Given []string `xml:"GIVEN"`
	Table string   `xml:"TABLE"`
}

// readXMLBIF reads a xmlbif file into a xmlbif object
func readXMLBIF(fname string) XMLBIF {
	f := ioutl.OpenFile(fname)
	defer f.Close()
	b, err := ioutil.ReadAll(f)
	errchk.Check(err, "")
	var net XMLBIF
	err = xml.Unmarshal(b, &net)
	errchk.Check(err, "")
	return net
}
