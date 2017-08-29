go-distribution
===============
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
