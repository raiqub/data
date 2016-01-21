# Data

Data is a library for the [Go Programming Language][go]. It defines interfaces
shared by other packages that implement data storages.

## Status

[![Build Status](https://travis-ci.org/raiqub/data.svg?branch=master)](https://travis-ci.org/raiqub/data)
[![Coverage Status](https://coveralls.io/repos/raiqub/data/badge.svg?branch=master&service=github)](https://coveralls.io/github/raiqub/data?branch=master)
[![GoDoc](https://godoc.org/github.com/raiqub/data?status.svg)](http://godoc.org/github.com/raiqub/data)

## Features

* **Store** interface for objects that store expirable values.
* **memstore.Store** type to store expirable values in-memory.
* **mongostore.Store** type to store expirable values in MongoDB.

## Installation

This library provides two Store Implementations: in-memory and MongoDB.

### In-Memory

To install in-memory implementation of Store run the following command:

```bash
go get gopkg.in/raiqub/data.v0/memstore
```

To import this package, add the following line to your code:

```bash
import "gopkg.in/raiqub/data.v0/memstore"
```

### MongoDB

To install MongoDB implementation of Store run the following command:

```bash
go get gopkg.in/raiqub/data.v0/mongostore
```

To import this package, add the following line to your code:

```bash
import "gopkg.in/raiqub/data.v0/memstore"
```

## Examples

Examples can be found on [library documentation][doc].

## Running tests

The tests can be run via the provided Bash script:

```bash
./test.sh
```

## License

raiqub/dot is made available under the [Apache Version 2.0 License][license].

[go]: http://golang.org/
[doc]: https://godoc.org/gopkg.in/raiqub/data.v0
[license]: http://www.apache.org/licenses/LICENSE-2.0
