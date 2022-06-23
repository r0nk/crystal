package main

import (
	"bufio"
	"fmt"
	"io"
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

	before := time.Now()

	stat, err := os.Stat(cause)
	if err != nil {
		log.Fatal(err)
	}

	input_size := stat.Size()

	cmd := exec.Command(e.script, cause)
	stdin, err := cmd.StdinPipe()

	file, err := os.Open(cause)
	io.Copy(stdin, file)
	file.Close()

	if err != nil {
		log.Fatal(err)
	}

	out, err := cmd.Output()

	if err != nil {
		log.Fatal(err)
	}

	file, err = os.Create(e.output)
	if err != nil {
		log.Fatal(err)
	}
	file.WriteString(string(out))
	stat, err = file.Stat()

	if err != nil {
		log.Fatal(err)
	}
	output_size := stat.Size()

	file.Close()

	diff := time.Now().Sub(before)
	fmt.Printf("%v %d %s(%d) %s %s(%d)\n", time.Now().UnixNano(), diff.Nanoseconds(), cause, input_size, e.script, e.output, output_size)
}

func matching_files(regex string) []string {
	var files []string
	root := "."

	if regex[0] == '/' {
		root = "/"
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

	if err != nil {
		log.Fatal(err)
	}

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
	count := 0

	var e edge
	for fscan.Scan() {
		count += 1
		fields := strings.Fields(fscan.Text())
		e.listen_regex = fields[0]
		e.script = fields[1]
		e.output = fields[2]
		ret = append(ret, e)
	}

	if count == 0 {
		log.Fatal("No edges found in ./crystalfile, exiting.")
	}

	//TODO edges should also have options as the 4th field
	// append
	// delete after use
	// line by line

	//TODO do some basic sanity checks on the graph
	// (infinite loops, multiple connections and the like)

	return ret
}

func main() {
	fmt.Printf("Starting up...\n")

	for _, edge := range read_edges() {
		go handle_edge(edge)
	}
	fmt.Printf("Finished starting up, watching and waiting\n")

	for {
		//TODO watch crystal file for changes
		//TODO watch this directory for new files that might match regex
		time.Sleep(1 * time.Second)
	}
}
