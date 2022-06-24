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
		log.Fatalf("%s[%s]\n", err, e.script)
	}

	input_size := stat.Size()

	cmd := exec.Command(e.script, cause)
	stdin, err := cmd.StdinPipe()

	file, err := os.Open(cause)
	io.Copy(stdin, file)
	file.Close()

	if err != nil {
		log.Fatalf("%s[%s]\n", err, e.script)
	}
	out, err := cmd.Output()
	if err != nil {
		log.Fatalf("%s[%s]\n", err, e.script)
	}

	file, err = os.Create(e.output)
	if err != nil {
		log.Fatalf("%s[%s]\n", err, e.script)
	}
	file.WriteString(string(out))
	stat, err = file.Stat()

	if err != nil {
		log.Fatalf("%s[%s]\n", err, e.script)
	}
	output_size := stat.Size()

	file.Close()

	diff := time.Now().Sub(before)
	fmt.Printf("%v %d %s(%d) %s %s(%d)\n", time.Now().UnixNano(), diff.Nanoseconds(), cause, input_size, e.script, e.output, output_size)
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
		line := fscan.Text()
		fields := strings.Fields(line)
		if line[0] == '#' {
			continue
		}
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
	// anew
	// clear after use
	// line by line
	// split

	//TODO do some basic sanity checks on the graph
	// (infinite loops, multiple connections and the like)

	return ret
}

func handle_events(crystalfile string) {
	edges := read_edges()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	watcher.Add("./")
	filepath.Walk("./", func(walkPath string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fi.IsDir() {
			if err = watcher.Add(walkPath); err != nil {
				return err
			}
		}
		return nil
	})

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			//			log.Println("event:", event)
			if event.Op&fsnotify.Write == fsnotify.Write {
				if event.Name == crystalfile {
					edges = read_edges()
				}
				for _, e := range edges {
					match, err := regexp.MatchString(e.listen_regex, event.Name)
					if err != nil {
						log.Fatal(err)
					}
					if match {
						go run_edge_script(e, event.Name)
					}
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("error:", err)
		}
	}

}

func main() {
	handle_events("crystalfile")
}
