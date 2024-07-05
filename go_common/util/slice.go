package util

// ReverseInt64 ...
func ReverseInt64(s []int64) {
	for left, right := 0, len(s)-1; left < right; left, right = left+1, right-1 {
		s[left], s[right] = s[right], s[left]
	}
}

// ReverseString ...
func ReverseString(s []string) {
	for left, right := 0, len(s)-1; left < right; left, right = left+1, right-1 {
		s[left], s[right] = s[right], s[left]
	}
}

// ReverseRune ...
func ReverseRune(s []rune) {
	for left, right := 0, len(s)-1; left < right; left, right = left+1, right-1 {
		s[left], s[right] = s[right], s[left]
	}
}

// ContainsString ... check whether slice contains target string
func ContainsString(slice []string, target string) bool {
	for _, v := range slice {
		if target == v {
			return true
		}
	}

	return false
}

// ContainsRune ... check whether slice contains target rune
func ContainsRune(slice []rune, target rune) bool {
	for _, v := range slice {
		if target == v {
			return true
		}
	}

	return false
}

// ConcatStringSlices ... concat a list of string slices
func ConcatStringSlices(slices [][]string) []string {
	var output []string
	for _, slice := range slices {
		output = append(output, slice...)
	}
	return output
}

// GetUnionString ... get union set of string slices
func GetUnionSetString(s1, s2 []string) []string {
	s := append(s1, s2...)
	s = RemoveDuplicateString(s)
	return s
}

// GetIntersectionSetString ... get intersection set of string slices
func GetIntersectionSetString(s1, s2 []string) []string {
	hash := ListToMapString(s1)
	inter := []string{}
	for _, e := range s2 {
		if hash[e] {
			inter = append(inter, e)
		}
	}
	inter = RemoveDuplicateString(inter)
	return inter
}
