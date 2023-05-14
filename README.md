# reflectutil

`reflectutil` is a Go package that provides utilities for working with
reflection in Go. This is a distillation of functions from several codebases,
cleaned up, tested, and gathered in one place for easy reuse.

The main thing this package is useful for is parsing struct tags in the common
`value,key1,key2:value2` syntax, handling annoying things like missing values,
repeated keys, spaces and quotes in parameters, etc.
