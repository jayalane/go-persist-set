// -*- tab-width: 2 -*-

package persistset

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

var nameFrags = []string{
	"hello",
	"db",
	"chris",
	"whatever/is/it/called",
	"dlv",
	"s3-bucket-control",
}

func makeName() string {
	values := []string{}
	for range 10 {
		values = append(values, nameFrags[rand.Intn(len(nameFrags))]) //nolint:gosec
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

	for range 1 {
		d.Add(makeName())
	}

	time.Sleep(1100 * time.Millisecond)
	time.Sleep(1100 * time.Millisecond)

	for range 100 {
		wg.Add(1)

		go func() {
			if d.InSet(makeName()) {
				t.Log("Random new string was in set, unlikely")
				t.Fail()
			}

			if !d.InSet(saveName) {
				t.Log("failed to retrieve set value")
				t.Fail()
			}

			wg.Done()
		}()
	}

	wg.Wait()

	defer os.Remove("test1.db")
}
