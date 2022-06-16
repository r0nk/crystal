package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
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

func matching_files(regex string) []string {
	var files []string

	if regex[0] == '/' {
		root := "/"
	} else {
		root := "."
	}

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {

		if err != nil {
			fmt.Println(err)
			return nil
		}
		match, err := regexp.MatchString(regex, path)

		if err != nil {
			log.Fatal(err)
		}

		if !info.IsDir() && match {
			files = append(files, path)
		}

		return nil
	})

	return files
}

func handle_edge(e edge) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	for _, file := range matching_files(e.listen_regex) {
		watcher.Add(file)
		fmt.Printf("Setting up edge: %s | %s > %s\n", file, e.script, e.output)
	}

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
	// (infinite loops, multiple connections and the like)

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
