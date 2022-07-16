#!/bin/bash

go run . &

sleep 5
echo asdf > nodes/from.txt
sleep 1
echo hackerone.com > nodes/domains.txt
sleep 5

echo killing...
pkill crystal
