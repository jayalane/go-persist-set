// -*- tab-width: 2 -*-

// Package counters enables 1 line creation of stats to track your program flow; you get summaries every minute
// implemented using channels now but switching to sync based on these tests maybe; will keep both implementations
package persistSet

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

var nameFrags = []string{"hello",
	"db",
	"chris",
	"whatever/is/it/called",
	"dlv",
	"s3-bucket-control",
}

func makeName() string {
	values := []string{}
	for i := 0; i < 10; i++ {
		values = append(values, nameFrags[rand.Intn(len(nameFrags))])
	}
	return strings.Join(values, "-")
}

func TestSet(t *testing.T) {
	wg := new(sync.WaitGroup)
	d := New("test1")
	defer d.Close()
	saveName := makeName()
	fmt.Println(saveName)
	d.Add(saveName)
	for i := 0; i < 1; i++ {
		//		go func() {
		d.Add(makeName())
		//		}()
	}
	//	for i := 0; i < 1; i++ {
	//		go func() {
	//			d.Add(makeName()) // why both?
	//		}()
	//	}
	time.Sleep(1100 * time.Millisecond)
	time.Sleep(1100 * time.Millisecond)
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			if d.InSet(makeName()) {
				t.Fatal("Random new string was in set, unlikely")
			}
			if !d.InSet(saveName) {
				t.Fatal("failed to retrieve set value")
			}
			wg.Done()
		}()
	}
	wg.Wait()
	defer os.Remove("test1.db")

}
