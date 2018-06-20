// Plane permutations are those which avoid the pattern 213'54
// Avoiding 213'54 is equivalent to avoiding 2-14-3
// This generating tree based enumeration is based on the following paper:
// https://arxiv.org/pdf/1702.04529.pdf

package main

import (
	"fmt"
	"sync"
)

// struct to mimick set in Python
// code from https://play.golang.org/p/_FvECoFvhq
type ArraySet struct {
	set map[[20]int]bool
	mux sync.Mutex
}

func NewArraySet() *ArraySet {
	return &ArraySet{set: make(map[[20]int]bool)}
}

func (set *ArraySet) Add(p [20]int) bool {
	_, found := set.set[p]
	set.set[p] = true
	return !found //False if it existed already
}

func (set *ArraySet) Get(i [20]int) bool {
	_, found := set.set[i]
	return found //true if it existed already
}

func (set *ArraySet) Remove(i [20]int) {
	delete(set.set, i)
}

// -----

// func to mimick min in Python
func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

func expansion(curLevel *ArraySet, level int, p chan []int, c chan bool) {
	// handles the creations of permutation concurrently
	// new permutation are added to p
	// curLevel and level is used to generate new permutation

	//fmt.Printf("Starting exapansion of level: %d \n", level)

	// To make sure that the channel	 p is closed after all permutations are calculated
	var wgExpansion sync.WaitGroup

	for perm := range curLevel.set {
		for a := 1; a <= level+1; a++ {
			//fmt.Printf("Starting a Perm Generation of a: %d \n", a)

			// No Reason?
			var permArr [20]int
			copy(permArr[:], perm[:])

			wgExpansion.Add(1)
			c <- true
			go localExp(permArr[:], a, level, p, &wgExpansion, c)

			//time.Sleep(1 *time.Second)
		}
	}

	wgExpansion.Wait()
	//fmt.Printf("!!!!! Done expansion of level: %d \n", level)
	close(p)
}

func localExp(perm []int, a int, level int, p chan []int, wgExpansion *sync.WaitGroup, c chan bool) {

	//defer fmt.Printf("Done a new perm generation: %d\n", a)
	defer wgExpansion.Done()

	// Local expansion as described in the paper
	newPerm := make([]int, level+1)
	for i, k := range perm {
		if k == 0 {
			break
		}

		if k < a {
			newPerm[i] = k
		} else {
			newPerm[i] = k + 1
		}
	}

	newPerm[level] = a
	// Adding a new permutation to channel p. Will be consumed by checkPlane() through for
	p <- newPerm

	<-c
}

func isPlane(newLevel *ArraySet, perm []int, wg *sync.WaitGroup) {
	// If perm is a plane, then perm is added to newLevel else we nothing is done

	n := len(perm)
	steps := make([]int, 0, n)
	for k := 0; k < n-1; k++ {
		if perm[k] < perm[k+1]-1 {
			steps = append(steps, k)
		}
	}

	for _, s := range steps {
		m, M := perm[s], perm[s+1]
		two := 1000
		prefix, suffix := perm[:s], perm[s+2:]
		for _, k := range prefix {
			if (k > m) && (k < M-1) {
				two = min(k, two)
			}
		}

		for _, k := range suffix {
			if (k > two) && (k < M) {
				// perm is not a plane
				//fmt.Println("Done checking the new perm: False")
				wg.Done()
				return
			}
		}
	}

	// perm is a plane

	var permArr [20]int
	copy(permArr[:], perm)

	//// To prevent race condition
	newLevel.mux.Lock()
	newLevel.Add(permArr)
	newLevel.mux.Unlock()

	//fmt.Println("Done checking the new perm: True")
	wg.Done()
	return
}

func checkPlane(newLevel *ArraySet, p chan []int) {
	// checks all the permutation(elements) in channel p util the channel p is closed
	// Permutations in p that are  a plane are added to newLevel

	// Useless it is?
	// To make sure that all the planes are checked before returning from here
	var wgCheckPlace sync.WaitGroup

	for newPerm := range p {
		//fmt.Println("Checking a new perm")
		wgCheckPlace.Add(1)
		go isPlane(newLevel, newPerm, &wgCheckPlace)
	}

	wgCheckPlace.Wait()

}

func main() {

	curLevel := NewArraySet()
	var arr1, arr2, arr3, arr4, arr5, arr6 [20]int
	copy(arr1[:], []int{1, 2, 3})
	copy(arr2[:], []int{1, 3, 2})
	copy(arr3[:], []int{2, 1, 3})
	copy(arr4[:], []int{3, 1, 2})
	copy(arr5[:], []int{2, 3, 1})
	copy(arr6[:], []int{3, 2, 1})

	curLevel.Add(arr1)
	curLevel.Add(arr2)
	curLevel.Add(arr3)
	curLevel.Add(arr4)
	curLevel.Add(arr5)
	curLevel.Add(arr6)

	level := 3

	for level < 9 {
		newLevel := NewArraySet()
		p := make(chan []int)

		c := make(chan bool, procs)

		go expansion(curLevel, level, p, c)

		checkPlane(newLevel, p)

		fmt.Println(len(newLevel.set))
		curLevel = newLevel
		level += 1
	}
}
