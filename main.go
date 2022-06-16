package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/fsnotify/fsnotify"
)

type edge struct {
	listen_regex string
	output       string
	script       string
}

func run_edge_script(e edge, cause string) {
	//set environmental variable to changing file
	fmt.Printf("running edge script" + e.output)
	out, err := exec.Command(e.script, cause).Output()
	if err != nil {
		log.Fatal(err)
	}
	file, err := os.Create(e.output)
	defer file.Close()
	if err != nil {
		log.Fatal(err)
	}
	file.WriteString(string(out))
}

func handle_edge(e edge) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	//TODO do this for every file that the regex matches
	// it currently only does direct matches
	watcher.Add(e.listen_regex)
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			log.Println("event:", event)
			if event.Op&fsnotify.Write == fsnotify.Write {
				go run_edge_script(e, event.Name)
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("error:", err)
		}
	}
	//TODO handle multiple connection cases
	fmt.Printf("asdf", e)
}

func read_edges() []edge {
	var ret []edge
	//for each entry in crystalfile
	//read entry
	//add to ret
	//return ret
	var e edge
	e.listen_regex = "nodes/from.txt"
	e.output = "nodes/to.txt"
	e.script = "edges/count"
	ret = append(ret, e)
	return ret
}

func main() {
	fmt.Printf("starting up...")
	for _, edge := range read_edges() {
		go handle_edge(edge)
	}
	for { //watch edges
		time.Sleep(1 * time.Second)
	}
}
