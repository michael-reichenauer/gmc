package utils

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func Test(t *testing.T) {
	var names []string
	assert.Nil(t, names)

	names2 := []string{}
	assert.NotNil(t, names2)

	assert.Equal(t, len(names), len(names2))

	names3 := []string{}
	assert.True(t, &names2 != &names3)
	assert.NotEqual(t, names, names2)
	assert.Equal(t, names2, names3)
	names2 = []string{"name"}
	names3 = []string{"name"}
	assert.Equal(t, names2, names3)
	names3 = []string{"name3"}
	assert.NotEqual(t, names2, names3)
}

func TestInfiniteChannel(t *testing.T) {
	in, out := InfiniteChannel()
	lastVal := -1
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		for v := range out {
			vi := v.(int)
			assert.Equal(t, lastVal+1, vi)
			lastVal = vi
		}
		wg.Done()
	}()

	for i := 0; i < 100; i++ {
		in <- i
	}
	close(in)
	fmt.Println("Finished writing")
	wg.Wait()
	assert.Equal(t, 99, lastVal)
}
