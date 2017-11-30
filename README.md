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
Learn parameters for a given structure:

	cd examples/
	lkbn ctparam -d example.csv -p parms.yaml -bi example#0.ctree -bo example#1.ctree

Sample clique tree structures:

	ctsample -d example.csv -p parms.yaml

#### Parameters file:

YAML file containing specific parameters for the learning algorithms

##### Parameters file fields:

	em_max_iters:		max number of EM iterations

##### Parameters file example:

	./examples/parms.yaml
	# structure sampling
	treewidth     : 1
	num_samples   : 3
	latent_vars   : 7, 12

	# parameters learning
	em_max_iters  : 20
	em_threshold  : 1e-5
	em_use_parms  : true
	em_init_iters : 2
	em_restarts   : 4
