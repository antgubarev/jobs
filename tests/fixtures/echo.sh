#!/usr/bin/env bash

trap "echo SIGINT; exit" SIGINT

for (( i=1; i<$1; i++ ))
do
  echo "step $i"
  sleep $2
done

echo "finish"