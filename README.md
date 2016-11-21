# Gogen [![CircleCI](https://img.shields.io/circleci/project/github/RedSparr0w/node-csgo-parser.svg)](https://github.com/coccyx/gogen) [![Coveralls](https://img.shields.io/coveralls/jekyll/jekyll.svg)](http://github.com/coccyx/gogen) [![Go Report Card](https://goreportcard.com/badge/github.com/coccyx/gogen)](https://goreportcard.com/report/github.com/coccyx/gogen) 

Gogen is an open source data generator.  Gogen can be used for any type of data, and it has first class support for time series data.  Primarily,
it's been used and tested to generate log and metric data for testing time series systems.

## Features

* Generates complex time series data via token substitutions in original raw events
* Support for arbitrary key/value datasets, tokens can replace in any field
* Many token types: static, randomly generated, different types of choices from lists, or custom scripts
* Three generation modes: random substitution of tokens from a sample file, replaying a sample in time series order, or custom generation scripts
* Extensible via custom Lua scripts
* Easy configuration via YAML or JSON files
* Easy sharing of configurations via a centralized service
* Simple getting started experience as one statically linked binary, compiled on multiple platforms