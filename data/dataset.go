// Package data is used to handle datasets
package data

import (
	"bufio"
	"strings"

	"github.com/britojr/lkbn/vars"
	"github.com/britojr/utl/conv"
	"github.com/britojr/utl/errchk"
	"github.com/britojr/utl/ioutl"
	"github.com/kniren/gota/dataframe"
)

// Dataset extends a dataframe to also deal with variables
type Dataset struct {
	vs vars.VarList
	df dataframe.DataFrame

	intMaps []map[int]int // cached intMaps
}

// NewDataset return a new dataset type
func NewDataset(fname, hdrFile string, hasHdr ...bool) (d *Dataset) {
	// TODO: add more options to dataset reading
	f := ioutl.OpenFile(fname)
	defer f.Close()
	d = new(Dataset)
	if len(hasHdr) > 0 {
		d.df = dataframe.ReadCSV(bufio.NewReader(f), dataframe.HasHeader(hasHdr[0]))
	} else {
		d.df = dataframe.ReadCSV(bufio.NewReader(f))
	}
	schema, names := []int{}, []string{}
	if len(hdrFile) > 0 {
		r := ioutl.OpenFile(hdrFile)
		defer r.Close()
		var lines [][]string
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			if len(scanner.Text()) == 0 {
				continue
			}
			lines = append(lines, strings.Split(scanner.Text(), ","))
		}
		if len(lines) == 1 {
			schema = conv.Satoi(lines[0])
		} else {
			if len(lines) > 1 {
				names = lines[0]
				schema = conv.Satoi(lines[1])
			}
		}
	}
	if len(names) > 0 {
		d.df.SetNames(names...)
	}
	for id, name := range d.df.Names() {
		var nstate int
		if id < len(schema) {
			nstate = schema[id]
		} else {
			nstate = int(d.df.Col(name).Max()) + 1
		}
		v := vars.New(id, nstate, name, false)
		d.vs = append(d.vs, v)
	}
	return
}

// IntMaps return a slice of intmaps of the dataset
func (d *Dataset) IntMaps() []map[int]int {
	// TODO: check this
	if len(d.intMaps) == 0 {
		r, c := d.df.Dims()
		var err error
		d.intMaps = make([]map[int]int, r)
		for i := 0; i < r; i++ {
			d.intMaps[i] = make(map[int]int)
			for j := 0; j < c; j++ {
				d.intMaps[i][j], err = d.df.Elem(i, j).Int()
				errchk.Check(err, "dataset: invalid data")
			}
		}
	}
	return d.intMaps
}

// Variables return dataset variables
func (d *Dataset) Variables() vars.VarList {
	return d.vs
}

// DataFrame returns dataframe
func (d *Dataset) DataFrame() dataframe.DataFrame {
	return d.df
}
