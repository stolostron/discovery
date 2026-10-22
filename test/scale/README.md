# Testing

This directory contains code to test loads on the discovery operator while monitoring resource consumption.

## Scale test

The scale test generates an increasingly large number of DiscoveredClusters. Progressive load is applied to the operator and resource consumption is continuously recorded and saved in `results/`. To run the scale tests first install the operator and mock server, then run the following:

```shell
go run test/scale/scale.go
```
