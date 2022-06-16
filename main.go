package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

type edge struct {
	listen_regex string
	output       string
	script       string
}

func run_edge_script(e edge, cause string) {

	//TODO pipe the cause file into standard input for the program

	fmt.Printf("%s %s %s (%s)\n", e.listen_regex, e.script, e.output, cause)
	// set environmental variable to changing file
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
	//TODO handle multiple connection cases
	watcher.Add(e.listen_regex)
	fmt.Printf("Listening to %s\n", e.listen_regex)

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			//log.Println("event:", event)
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
}

func read_edges() []edge {
	crystalfile, err := os.Open("crystalfile")
	if err != nil {
		log.Fatal(err)
	}
	var ret []edge

	fscan := bufio.NewScanner(crystalfile)
	fscan.Split(bufio.ScanLines)

	var e edge
	for fscan.Scan() {
		fields := strings.Fields(fscan.Text())
		e.listen_regex = fields[0]
		e.script = fields[1]
		e.output = fields[2]
		ret = append(ret, e)
	}

	//TODO do some basic sanity checks on the graph
	// (infinite loops and the like)

	return ret
}

func main() {
	for _, edge := range read_edges() {
		go handle_edge(edge)
	}
	for { //TODO watch crystal file for changes
		time.Sleep(1 * time.Second)
	}
}
