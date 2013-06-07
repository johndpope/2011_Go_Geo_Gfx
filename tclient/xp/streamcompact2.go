package main

import (
	"fmt"
)

var (
	unsorted = []int {
		55, 88, 33, 44,
		11, 66, 22, 88,
		77, 99, 77, 11,
		22, 33, 44, 55,
		66, 99, 88, 99,
		11, 33, 55, 77,
	}
	usage = []int {
		500, 800, 300, 400,
		100, 600, 200, 800,
		700, 900, 700, 100,
		200, 300, 400, 500,
		600, 900, 800, 900,
		100, 300, 500, 700,
	}
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
		if usage[offset + i] == needle { cv++ } else { ci++ }
	}
	fragCountsValid[fragIndex] = cv
	fragCountsInvalid[fragIndex] = ci
	ch <- 1
}

func fragPhase2_PrefixSums () {
	var rt = 0
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
	var foffv, foffi, v, u = fragOffsetsValid[fragIndex], fragOffsetsInvalid[fragIndex], 0, 0
	for i := 0; i < numElems; i++ {
		u = usage[offset + i]
		v = unsorted[offset + i]
		if u == needle { sorted[foffv] = v; foffv++ } else { sorted[foffi] = v; foffi++ }
	}
	ch <- 1
}

func streamCompact (needle int) []int {
	var tmp = 0
	var sorted = make([]int, len(unsorted))
	fmt.Printf("BFOR (for %v) %+v\n", needle, unsorted)
	// p1
	for f := 0; f < numFrags; f++ { go fragPhase1_CountElems(f, (len(unsorted) / numFrags) * f, needle, len(unsorted) / numFrags) }
	for tmp < numFrags { tmp += <- ch }
	// p2
	fmt.Printf("VCOUNT %+v\n", fragCountsValid)
	fmt.Printf("ICOUNT %+v\n", fragCountsInvalid)
	fragPhase2_PrefixSums()
	fmt.Printf("VOFFS %+v\n", fragOffsetsValid)
	fmt.Printf("IOFFS %+v\n", fragOffsetsInvalid)
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
	for i := 100; i <= 300; i += 100 { unsorted = streamCompact(i) }
}
