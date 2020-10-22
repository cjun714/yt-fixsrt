package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
)

type subSegment struct {
	durationStr string
	content     []string
}

func main() {
	if len(os.Args) != 2 {
		log.Fatal(".srt directory path is not specified")
		fmt.Println("usage: yt-fixsrt <srt_directory>")
	}

	dirPath := os.Args[1]

	fs, e := ioutil.ReadDir(dirPath)
	if e != nil {
		log.Fatal(e)
	}

	srtPathes := make([]string, 0)
	for _, fi := range fs {
		name := fi.Name()
		if strings.HasSuffix(name, "srt") {
			srtPathes = append(srtPathes, dirPath+"/"+name)
		}
	}

	for _, path := range srtPathes {
		if e = fixSrt(path); e != nil {
			log.Fatal(e)
		}
	}
}

func fixSrt(path string) error {
	sub, e := readSub(path)
	if e != nil {
		return e
	}

	newSub := removeRedundantSubs(sub)

	// backup old sub file
	bakPath := path + ".bak"
	e = os.Rename(path, bakPath)
	if e != nil {
		return e
	}

	if e = saveSub(path, newSub); e != nil {
		return e
	}

	return nil
}

func readSub(path string) ([]subSegment, error) {
	sub := make([]subSegment, 100)

	bs, e := ioutil.ReadFile(path)
	if e != nil {
		return nil, e
	}

	scanner := bufio.NewScanner(bytes.NewReader(bs))
	for scanner.Scan() {
		var seg subSegment

		line := scanner.Text()
		line = strings.Trim(line, " ")

		// read sub number
		_, e := strconv.Atoi(line)
		if e != nil {
			return nil, e
		}

		// read duration
		if !scanner.Scan() {
			break
		}
		seg.durationStr = scanner.Text()
		seg.content = make([]string, 0)

		first := true
		// read sub content
		for scanner.Scan() {
			str := scanner.Text()
			if len(str) == 0 && !first {
				break
			}

			str = strings.Trim(str, " ")
			if str != "" {
				seg.content = append(seg.content, scanner.Text())
			}

			if first {
				first = false
			}
		}

		sub = append(sub, seg)
	}

	if e = scanner.Err(); e != nil {
		log.Fatal(e)
	}

	return sub, nil
}

func removeRedundantSubs(sub []subSegment) []subSegment {
	newSub := make([]subSegment, 0)
	count := len(sub)
	for i := count - 1; 0 < i; i-- {
		if i != 0 {
			seg, preSeg := &sub[i], &sub[i-1]
			if len(seg.content) != 0 && len(preSeg.content) != 0 {
				if seg.content[0] == preSeg.content[len(preSeg.content)-1] {
					seg.content = seg.content[1:]
				}
			}
		}
	}

	for i := 0; i < count; i++ {
		if len(sub[i].content) != 0 {
			newSub = append(newSub, sub[i])
		}
	}
	return newSub
}

func saveSub(path string, sub []subSegment) error {
	log.Println("save ", path)

	f, e := os.Create(path)
	if e != nil {
		return e
	}
	defer f.Close()

	for idx, seg := range sub {
		// write sub number
		_, e = f.WriteString(strconv.Itoa(idx) + "\n")
		if e != nil {
			return e
		}

		// write duration
		_, e = f.WriteString(seg.durationStr + "\n")
		if e != nil {
			return e
		}

		// write sub content
		for _, str := range seg.content {
			_, e = f.WriteString(str + "\n")
			if e != nil {
				return e
			}
		}

		// write newline
		if idx < len(sub)-1 {
			_, e = f.WriteString("\n")
			if e != nil {
				return e
			}
		}
	}

	return e
}
