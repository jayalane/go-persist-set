// -*- tab-width: 2 -*-

// Package persistset is a non-transactional but low-impact way to
// persist a set of things to the disk.  It is intended for things
// like "my job checking 400 million objects has to be stopped and
// restarted but I want to pick up roughly where I left off" The code
// should idempotent but you want to quickly jump back to where you
// where, roughly.  So you queue up persists into a channel and
// soonish they are on disk, append only file.  No compression or
// anything
package persistset

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"sync"
)

const (
	channelSize     = 1_000_000
	defaultFileMode = 0o600
)

// DB is the structure containing set.
type DB struct {
	setLock sync.RWMutex
	setDB   map[string]int
	name    string
	done    chan bool
	setChan chan string
}

// New returns an initialized structure to persist data to the disk for.
func New(dbName string) (d *DB) { //nolint:nonamedreturns
	var err error

	d = &DB{}
	d.setLock = sync.RWMutex{}
	d.name = dbName
	d.setDB = make(map[string]int)

	// read from disk if there
	binaryFilename, err := os.Executable()
	if err != nil {
		panic(err)
	}

	filePath := path.Join(path.Dir(binaryFilename), d.name+".db")

	file, err := os.Open(filePath)
	if err != nil {
		log.Println("Warning: can't open db file, initting empty", d.name+".db", filePath, err.Error())
	} else {
		log.Println("Using db file", filePath)

		defer file.Close()

		fileReader := bufio.NewReader(file)

		err = d.initMap(fileReader)
		if err != nil {
			panic(fmt.Sprintf("Can't read DB file, remove and try again %s", err))
		}
	}

	d.setChan = make(chan string, channelSize) // should be config driven
	d.done = make(chan bool, 1)                // Not really using yet

	go func() { // async writer
		// open one file for the duration
		f, err := os.OpenFile(d.name+".db",
			os.O_APPEND|os.O_CREATE|os.O_WRONLY, defaultFileMode)
		if err != nil {
			log.Println(err)
		}

		defer f.Close()

		ff := bufio.NewWriter(f)

		for {
			select {
			case <-d.done:
				return
			case s := <-d.setChan:
				if _, err := ff.WriteString(s + "\n"); err != nil { // later length prefix
					log.Println(err)
				}

				d.setLock.Lock()

				d.setDB[s] = 1

				d.setLock.Unlock()
			}
		}
	}()

	return d
}

// initMap is reading in the file.
func (d *DB) initMap(reader io.Reader) error {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()

		if err := scanner.Err(); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			log.Println("Error reading config", err)

			return err
		}

		d.setLock.Lock()

		d.setDB[line] = 1

		d.setLock.Unlock()
	}

	return nil
}

// Close shuts down the persisted set.
func (d *DB) Close() {
	d.done <- true
}

// Add puts the string into the set eventually.
func (d *DB) Add(s string) {
	d.setChan <- s
}

// InSet returns true if the string is in the set DB
// currently it doesn't look at pending values.
func (d *DB) InSet(s string) bool {
	d.setLock.RLock()
	defer d.setLock.RUnlock()
	_, ok := d.setDB[s]

	return ok
}
