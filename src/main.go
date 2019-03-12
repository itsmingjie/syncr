// Mingjie Jiang <jiang@mingjie.info>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/common-nighthawk/go-figure"
	"github.com/radovskyb/watcher"
)

const VER = "1.0.0"

func init() {
	PrintHeader()
}

// PrintHeader prints the header information of the program
func PrintHeader() {
	syncrLogo := figure.NewFigure("SYNCR", "", true)
	syncrLogo.Print()
	fmt.Println()

	// copyright + license information
	fmt.Println("Syncr by Mingjie Jiang <github.com/itsmingjie>")
	fmt.Println("Version ", VER)
	fmt.Println("Licensed under the Apache License, Version 2.0")
	fmt.Println()

	// instructions
	fmt.Println("Drag the source/target folders into this window")
	fmt.Println("And press [Enter] to confirm")
	fmt.Println("To terminate Syncr, press [CTRL] + [C]")
	fmt.Println()
}

// CleanDir trims off the suffix from the readline
func CleanDir(o string) string {
	o = strings.TrimSuffix(o, "\n")
	o = strings.TrimSuffix(o, "\r")
	return o
}

// DirCP copies a whole directory recursively
func DirCP(src string, dst string) error {
	var err error
	var fds []os.FileInfo
	var srcinfo os.FileInfo

	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}

	if err = os.MkdirAll(dst, srcinfo.Mode()); err != nil {
		return err
	}

	if fds, err = ioutil.ReadDir(src); err != nil {
		return err
	}
	for _, fd := range fds {
		srcfp := path.Join(src, fd.Name())
		dstfp := path.Join(dst, fd.Name())

		if fd.IsDir() {
			if err = DirCP(srcfp, dstfp); err != nil {
				fmt.Println(err)
			}
		} else {
			if err = FileCP(srcfp, dstfp); err != nil {
				fmt.Println(err)
			}
		}
	}
	return nil
}

// FileCP copies a single file from src to dst
func FileCP(src, dst string) error {
	var err error
	var srcfd *os.File
	var dstfd *os.File
	var srcinfo os.FileInfo

	if srcfd, err = os.Open(src); err != nil {
		return err
	}
	defer srcfd.Close()

	if dstfd, err = os.Create(dst); err != nil {
		return err
	}
	defer dstfd.Close()

	if _, err = io.Copy(dstfd, srcfd); err != nil {
		return err
	}
	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}
	return os.Chmod(dst, srcinfo.Mode())
}

// DirExists validates the path of the argument string
func DirExists(path string) bool {
	// check if the source dir exist
	src, err := os.Stat(path)
	if err != nil {
		fmt.Println(path, "does not exist.")
		return false
	}

	// check if the source is indeed a directory or not
	if !src.IsDir() {
		fmt.Println(path, "is not a directory.")
		return false
	}

	return true
}

func main() {
	w := watcher.New()
	reader := bufio.NewReader(os.Stdin)

	var srcDir, tgtDir string
	// take in 2 directories to start the process
	for {
		fmt.Print("Source directory: ")
		srcDir, _ = reader.ReadString('\n')
		srcDir = CleanDir(srcDir)

		if DirExists(srcDir) {
			break
		}

		fmt.Println("Please try again.")
		fmt.Println()
	}

	for {
		fmt.Print("\nTarget directory: ")
		tgtDir, _ = reader.ReadString('\n')
		tgtDir = CleanDir(tgtDir)

		if DirExists(tgtDir) {
			break
		}

		fmt.Println("Please try again.")
		fmt.Println()
	}

	// single event handling
	w.SetMaxEvents(1)

	r := regexp.MustCompile("^.*.(java)$")
	w.AddFilterHook(watcher.RegexFilterHook(r, false))

	go func() {
		fmt.Println("\n==============================")
		fmt.Println("Syncr service has started. Press [CTRL] + [C] to terminate.")
		for {
			select {
			case event := <-w.Event:
				fmt.Println(event) // Print the event's info.
				DirCP(srcDir, tgtDir)
			case err := <-w.Error:
				log.Fatalln(err)
			case <-w.Closed:
				return
			}
		}
	}()

	// Watch test_folder recursively for changes.
	if err := w.AddRecursive(srcDir); err != nil {
		log.Fatalln(err)
	}

	// Start the watching process - it'll check for changes every 100ms.
	if err := w.Start(time.Millisecond * 100); err != nil {
		log.Fatalln(err)
	}
}
