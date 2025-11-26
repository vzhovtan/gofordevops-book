package workflow

import (
	"crypto/md5"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"sync"
)

type Result struct {
	fHash string
	fPath string
}

// func Search is wrapper which add flag functionality and call GetStat func
func Search() {
	dir := flag.String("dir", ".", "directory to start search for duplicated files non-recursively")
	flag.Parse()
	GetStat(*dir, os.Stdout)
}

// func GetStat using helper function to get md5 has for files
func GetStat(filepath string, w io.Writer) {
	rslt := make(map[string][]string)
	done := make(chan struct{})
	defer close(done)
	c := getHash(done, filepath)

	for elem := range c {
		if _, ok := rslt[elem.fHash]; ok {
			rslt[elem.fHash] = append(rslt[elem.fHash], elem.fPath)
		} else {
			rslt[elem.fHash] = []string{elem.fPath}
		}
	}

	for _, elem := range rslt {
		if len(elem) > 1 {
			sort.Slice(elem, func(i, j int) bool {
				return elem[i] < elem[j]
			})
			fmt.Fprintf(w, "%v\n", elem)
		}
	}
}

// func getHash is helper function
func getHash(done chan struct{}, root string) chan Result {
	c := make(chan Result)
	go func() {
		var wg sync.WaitGroup
		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.Mode().IsRegular() {
				return nil
			}
			wg.Add(1)
			go func() {
				data, _ := os.ReadFile(path)
				select {
				case c <- Result{
					fPath: path,
					fHash: fmt.Sprintf("%s", md5.Sum(data)),
				}:
				case <-done:
				}
				wg.Done()
			}()
			select {
			case <-done:
				return errors.New("walk canceled")
			default:
				return nil
			}
		})
		go func() {
			wg.Wait()
			close(c)
		}()
		if err != nil {
			log.Fatal(err)
		}
	}()
	return c
}
