#!/bin/bash

go run . &

sleep 1
echo asdf > nodes/from.txt
sleep 1

echo killing...
pkill crystal
