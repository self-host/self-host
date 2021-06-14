package main

func StringSliceIndex(haystack []string, needle string) int {
        for idx, a := range haystack {
                if a == needle {
                        return idx
                }
        }
        return -1
}
