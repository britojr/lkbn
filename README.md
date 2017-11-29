# Latent K-Tree Bayesian Networks Learner

lkbn is a tool for learning bounded tree-width Bayesian networks with latent variables

[![Build Status](https://travis-ci.org/britojr/lkbn.svg?branch=master)](https://travis-ci.org/britojr/lkbn)
[![Coverage Status](https://coveralls.io/repos/github/britojr/lkbn/badge.svg?branch=master)](https://coveralls.io/github/britojr/lkbn?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/britojr/lkbn)](https://goreportcard.com/report/github.com/britojr/lkbn)
[![GoDoc](https://godoc.org/github.com/britojr/lkbn?status.svg)](http://godoc.org/github.com/britojr/lkbn)

___

## Installation and usage

### Get, install and test:

		go get -u github.com/britojr/lkbn...
		go install github.com/britojr/lkbn...
		go test github.com/britojr/lkbn... -cover

### Usage:

		lkbn --help
		Usage: lkbn <command> [options]

		Commands:

### Examples:

...

	cd examples/

Sample 10 clique tree structures with tree-width 3:

	ctsample -d example.csv -s 10 -k 3

#### Parameters file:

YAML file containing specific parameters for the learning algorithms

##### Parameters file fields:

	treewidth:		structure's tree-width upper bound

##### Parameters file example:

	./examples/example-parms.yaml
	treewidth: 4
