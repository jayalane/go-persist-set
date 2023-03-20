// -*- tab-width: 2 -*-

// Package counters enables 1 line creation of stats to track your program flow; you get summaries every minute
package persistSet

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"sync"
)

type SetDb struct {
	setLock sync.RWMutex
	setDb   map[string]int
	name    string
	started bool
	done    chan bool
	setChan chan string
}

func New(dbName string) (d *SetDb) {
	// Bolt DB for done Objects for copy
	var err error
	d = &SetDb{}
	d.setLock = sync.RWMutex{}
	d.name = dbName
	d.setDb = make(map[string]int)
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
	d.setChan = make(chan string, 1000000) // should be config driven
	d.done = make(chan bool, 1)            // Not really using yet

	go func() { // async writer
		// open one file for the duration
		f, err := os.OpenFile(d.name+".db",
			os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
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
				if _, err := ff.WriteString(s + "\n"); err != nil { // todo length prefix
					log.Println(err)
				}
				d.setLock.Lock()
				d.setDb[s] = 1
				d.setLock.Unlock()
			}
		}
	}()
	return d
}

// initMap is reading in the file
func (d *SetDb) initMap(reader io.Reader) error {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		if err := scanner.Err(); err != nil {
			if err == io.EOF {
				break
			}
			log.Println("Error reading config", err)
			return err
		}
		d.setLock.Lock()
		d.setDb[line] = 1
		d.setLock.Unlock()
	}
	return nil
}

// Close shutdownb
func (d *SetDb) Close() {
	d.done <- true
}

// Add puts the string into the set eventuall
func (d *SetDb) Add(s string) {
	d.setChan <- s
}

// InSet returns true if the string is in the set DB
// currently it doesn't look at pending Set
func (d *SetDb) InSet(s string) bool {
	d.setLock.RLock()
	defer d.setLock.RUnlock()
	_, ok := d.setDb[s]
	return ok
}
