package flags

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

type Fentry struct {
	fname string
	ftype bool
	fsize int64
}

func Run() {
	fflag := flag.Bool("f", false, "includes files into the output")
	flag.Parse()
	if len(flag.Args()) < 1 {
		fmt.Println("Not enough args")
		os.Exit(1)
	}
	start := flag.Args()[0]
	var last bool
	var lastlist []bool
	err := BuildDirTree(os.Stdout, start, last, fflag, lastlist)
	if err != nil {
		log.Fatal(err.Error())
	}
}

func BuildDirTree(out io.Writer, s string, last bool, fflag *bool, lastlist []bool) error {
	lastlist = append(lastlist, last)
	dirlist := []Fentry{}
	dirfiles, err := os.ReadDir(s)
	if err != nil {
		return err
	}
	for _, file := range dirfiles {
		info, err := file.Info()
		if err != nil {
			return err
		}
		entry := Fentry{
			file.Name(),
			file.IsDir(),
			info.Size(),
		}
		if *fflag {
			dirlist = append(dirlist, entry)
		} else if file.IsDir() {
			dirlist = append(dirlist, entry)
		}
	}
	err = walkDirTree(out, s, lastlist, fflag, dirlist)
	if err != nil {
		return err
	}
	return nil
}

func walkDirTree(out io.Writer, s string, lastlist []bool, fflag *bool, dirlist []Fentry) error {
	pathsep := string(os.PathSeparator)
	for idx, v := range dirlist {
		last := false
		depth := strings.Count(s, pathsep)
		if idx == len(dirlist)-1 {
			last = true
		}
		if v.ftype {
			fmt.Fprintln(out, printEntry(v, last, depth, lastlist))
			path := s + pathsep + v.fname
			BuildDirTree(out, path, last, fflag, lastlist)
		} else {
			// specific for MacOS
			if v.fname == ".DS_Store" {
				continue
			} else {
				fmt.Fprintln(out, printEntry(v, last, depth, lastlist))
			}
		}
	}
	return nil
}

func printEntry(v Fentry, last bool, depth int, lastlist []bool) (s string) {
	pfx := "├───"
	filesize := "(empty)"

	if v.fsize != 0 {
		filesize = "(" + strconv.FormatInt(v.fsize, 10) + "b" + ")"
	}

	if last {
		pfx = "└───"
	}

	switch {
	case depth == 0:
		if v.ftype {
			s = fmt.Sprintf("%s%s", pfx, v.fname)
		} else {
			s = fmt.Sprintf("%s%s %s", pfx, v.fname, filesize)
		}
	case depth != 0:
		var indent string
		for _, x := range lastlist[1:] {
			if x {
				indent += "\t"
			} else {
				indent += "│\t"
			}
		}
		if v.ftype {
			s = fmt.Sprintf("%s%s%s", indent, pfx, v.fname)
		} else {
			s = fmt.Sprintf("%s%s%s %s", indent, pfx, v.fname, filesize)
		}
	}
	return s
}
