#!/usr/bin/env bash

for i in 1 2 3 4 5
do
   go clean -testcache && make test > /dev/null
   echo $?
done
