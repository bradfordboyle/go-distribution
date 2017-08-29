go-distribution
===============
[![Build Status](https://travis-ci.org/bradfordboyle/go-distribution.svg?branch=master)](https://travis-ci.org/bradfordboyle/go-distribution)

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
