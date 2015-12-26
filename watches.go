/*
watches - An n-way file tree differencer
Copyright 2015 Chris Kennelly (chris@ckennelly.com)

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package main

import (
	"crypto/sha256"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
)

type stringSlice []string

func (x *stringSlice) String() string {
	return fmt.Sprintf("%v", *x)
}

func (x *stringSlice) Set(value string) error {
	*x = append(*x, value)
	return nil
}

func intMin(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func hash(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", fmt.Errorf("cannot open %v: %v", filename, err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return "", fmt.Errorf("cannot stat %v: %v", filename, err)
	}
	s := sha256.New()
	const chunk = 1 << 24
	for i := int64(0); i < info.Size(); i += chunk {
		blockSize := intMin(chunk, info.Size()-i)
		buf := make([]byte, blockSize)

		_, err := file.Read(buf)
		if err != nil && err != io.EOF {
			return "", fmt.Errorf("cannot read %v @ %v: %v", filename, i, err)
		}
		io.WriteString(s, string(buf))
	}

	return fmt.Sprintf("%x", s.Sum(nil)), nil
}

func main() {
	var searchPaths stringSlice
	flag.Var(&searchPaths, "search", "Search path (Multiple uses permitted)")
	flag.Parse()

	if len(searchPaths) == 0 {
		log.Fatal("No search paths specified.")
	}

	for _, dir := range searchPaths {
		if f, err := os.Stat(dir); err != nil || !f.IsDir() {
			log.Fatal("%v does not exist or is not a directory.", dir)
		}
	}

	checked := make(map[string]bool)
	for _, dir := range searchPaths {
		filepath.Walk(dir, func(walkedPath string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}
			stub := walkedPath[len(dir):]
			if _, ok := checked[stub]; ok {
				return nil
			}

			type message struct {
				Base string
				Hash string
				Err  error
			}

			in := make(chan string)
			out := make(chan message)
			for _, base := range searchPaths {
				go func() {
					b := <-in
					h, err := hash(path.Join(b, stub))
					out <- message{
						Base: b,
						Hash: h,
						Err:  err,
					}
				}()
				in <- base
			}

			hashes := make(map[string][]string)
			for i := 0; i < len(searchPaths); i++ {
				m := <-out

				if m.Err != nil {
					log.Printf("Unable to hash %v: %v", path.Join(m.Base, stub), err)
				} else {
					hashes[m.Hash] = append(hashes[m.Hash], m.Base)
				}
			}

			if len(hashes) > 1 {
				log.Printf("Mismatch %v: %v", stub, hashes)
			}

			checked[stub] = true
			return nil
		})
	}
}
