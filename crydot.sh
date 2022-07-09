#!/bin/bash

cat crystalfile  | tr '\/' '_' | tr -d \. | sed "s/\*/STAR/g" | awk 'BEGIN {print "digraph {" } {print "\t",$1,"->",$3,"[label=\"",$2,"\"];"} END { print "}" }'
