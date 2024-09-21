package main

import (
	"fmt"
	"strings"
)

func main() {
	fmt.Println(minimumRecolors("WBBWWWWBBWWBBBBWWBBWWBBBWWBBBWWWBWBWW", 15))
}
func minimumRecolors(blocks string, k int) []int {
	var b string
	var counts []int
	for i := 0; i < k; i++ {
		b += "B"
	}

	for a := 0; len(blocks)-a > k; a++ {
		count := 0
		s := blocks
		for {
			if strings.Contains(s, b) {
				break
			} else if strings.Contains(s, "W") {
				s = strings.Replace(s, "W", "B", 1)
				count++
			} else {
				break
			}
		}
		fmt.Println(s)
		counts = append(counts, count)
		blocks = strings.Replace(blocks, "B", "A", 1)
		if count >= len(blocks) {
			break
		}
	}
	return counts
}
