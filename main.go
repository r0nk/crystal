package main

import (
	"fmt"
	"os/exec"
)

type edge struct {
	listen_regex string
	output string
	script string
}

func run_edge_script(e edge,cause string) {
	cmd := exec.Command(e.script_file,cause)
	cmd.Run()
}

func handle_edge(e edge) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	//TODO do this for every file that the regex matches
	watcher.Add(e.listen_regex)
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			log.Println("event:", event)
			if event.Op&fsnotify.Write == fsnotify.Write {
				run_edge_script(e,event.Name)
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("error:", err)
		}
	}

	//on fs change:
		//set environmental variable to changing file
		//cmd := exec.Command(edge.script_file)
		//cmd.Run()
	//TODO handle multiple connection cases
	fmt.Printf("asdf",e);
}

func read_edges() edge[]{
	var ret edge[]
	//for each entry in crystalfile
		//read entry
		//add to ret
	//return ret
	var e edge
	e.listen_regex="nodes/from.txt"
	e.output="nodes/to.txt"
	e.script_file="edges/count"
	ret = ret.append(e)
	return ret
}

func main() {
	for _,edge := range read_edges() {
		go handle_edge(edge)
	}
	for { //watch edges
		time.Sleep(1*time.Second)
	}
}
