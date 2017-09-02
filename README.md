go-distribution
===============
[![Build Status](https://travis-ci.org/bradfordboyle/go-distribution.svg?branch=master)](https://travis-ci.org/bradfordboyle/go-distribution)
[![Coverage Status](https://coveralls.io/repos/github/bradfordboyle/go-distribution/badge.svg?branch=master)](https://coveralls.io/github/bradfordboyle/go-distribution?branch=master)

A Go version of philovivero's [distribution][] script.

Testing
-------

```sh
glide install
go test -v ./...
cd acceptance-tests && ./runTests.sh && cd -
```

Building
--------

```sh
go build -o distribution main.go
```

[distribution]: https://github.com/philovivero/distribution
