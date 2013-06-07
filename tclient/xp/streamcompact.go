package main

import (
	"fmt"
)

var (
	unsorted = []int {
		5, 8, 3, 4,
		1, 6, 2, 8,
		7, 9, 7, 1,
		2, 3, 4, 5,
		6, 9, 8, 9,
		1, 3, 5, 7,
	}
	rt = 0
	numFrags = 8
	numVerts = 4
	fragCountsValid = make([]int, numFrags)
	fragCountsInvalid = make([]int, numFrags)
	fragOffsetsValid = make([]int, numFrags)
	fragOffsetsInvalid = make([]int, numFrags)
	ch = make(chan int)
)

func fragPhase1_CountElems (fragIndex, offset, needle, numElems int) {
	var cv, ci = 0, 0
	for i := 0;  i < numElems; i++ {
		if unsorted[offset + i] == needle { cv++ } else { ci++ }
	}
	fragCountsValid[fragIndex] = cv
	fragCountsInvalid[fragIndex] = ci
	ch <- 1
}

func fragPhase2_PrefixSums () {
	for f := 0; f < numFrags; f++ {
		fragOffsetsValid[f] = rt
		rt = rt + fragCountsValid[f]
	}
	for f := 0; f < numFrags; f++ {
		fragOffsetsInvalid[f] = rt
		rt = rt + fragCountsInvalid[f]
	}
}

func fragPhase3_MoveElems (fragIndex, offset, needle, numElems int, sorted []int) {
	var foffv, foffi, v = fragOffsetsValid[fragIndex], fragOffsetsInvalid[fragIndex], 0
	for i := 0; i < numElems; i++ {
		v = unsorted[offset + i]
		if v == needle { sorted[foffv] = v; foffv++ } else { sorted[foffi] = v; foffi++ }
	}
	ch <- 1
}

func streamCompact (needle int) []int {
	var tmp = 0
	var sorted = make([]int, len(unsorted))
	rt = 0
	fmt.Printf("BFOR (for %v) %+v\n", needle, unsorted)
	// p1
	for f := 0; f < numFrags; f++ { go fragPhase1_CountElems(f, (len(unsorted) / numFrags) * f, needle, len(unsorted) / numFrags) }
	for tmp < numFrags { tmp += <- ch }
	// p2
	fragPhase2_PrefixSums()
	// p3
	tmp = 0
	for f := 0; f < numFrags; f++ { go fragPhase3_MoveElems(f, (len(unsorted) / numFrags) * f, needle, len(unsorted) / numFrags, sorted) }
	for tmp < numFrags { tmp += <- ch }
	fmt.Printf("DONE (for %v) %+v\n", needle, sorted)
	return sorted
}

func step (v1, v2 float32) float32 {
	if v2 < v1 { return 0 }
	return 1
}

func mix (v1, v2, a float32) float32 {
	return v1 * (1 - a) + (v2 * a)
}

func main () {
	var do = false
	var test float32 = 9
	var tmp float32
	var comps = []float32 { 4, 90, 27, 0, 5, 9, 3, -81, 18, 4.5 }
	if do {
		for _, c := range comps {
			tmp = mix(test, c, test - c)
			fmt.Printf("mix(%v,%v,%v) = %v\n", c, test, -1, tmp)
			fmt.Printf("step(%v,%v) = %v\n", test, tmp, step(test, tmp))
			fmt.Println("----------------------------")
		}
	}
	for i := 9; i >= 1; i-- { unsorted = streamCompact(i) }
}
