#!/bin/bash
go test -bench=. -run=None -benchtime=1s -failfast
#go test -bench=^Benchmark_Objects -run=^$ -benchtime=1s -failfast
#go test -bench=^Benchmark_Trans -run=^$ -benchtime=1s -failfast
