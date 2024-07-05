package util

import (
	"math"
	"strconv"
	"strings"
	"unicode"
)

// ScoredString ...
type ScoredString struct {
	Value string
	Score float32
}

// ScoredStrings ...
type ScoredStrings []ScoredString

// RemoveEmptyStr ...
func RemoveEmptyStr(tokens []string) []string {
	result := make([]string, 0, len(tokens))
	for _, t := range tokens {
		t = strings.Trim(t, " \t\n")
		if t != "" && t != " " {
			result = append(result, t)
		}
	}
	return result
}

// IsNumber ... check wether string is a number
func IsNumber(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

// IsLetter ... return true if strings contain only letter, false otherwise
func IsLetter(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) {
			return false
		}
	}
	return true
}

// SetEqual test whether two sets are the same
func SetEqual(s1, s2 []string) bool {
	if len(s1) != len(s2) {
		return false
	}
	m2 := ListToMapString(s2)
	for _, t := range s1 {
		if !m2[t] {
			return false
		}
	}
	return true
}

// SetDiff ...
func SetDiff(s1, s2 []string) []string {
	check := make(map[string]bool)
	for _, t := range s2 {
		check[t] = true
	}
	result := []string{}
	for _, t := range s1 {
		if !check[t] {
			result = append(result, t)
		}
	}
	return result
}

// RemoveDuplicateString ...
func RemoveDuplicateString(strs []string) []string {
	exist := make(map[string]bool)
	result := []string{}
	for _, str := range strs {
		if !exist[str] {
			result = append(result, str)
			exist[str] = true
		}
	}
	return result
}

// Min ...
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// MinInt64 ...
func MinInt64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

// MinFloat32 ...
func MinFloat32(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}

// Max ...
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// MaxInt64 ...
func MaxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

// MaxFloat32 ...
func MaxFloat32(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}

// MaxFloat32List ... return the maximum element in the float32 list
func MaxFloat32List(a []float32) float32 {
	max := -float32(math.MaxFloat32)
	for _, i := range a {
		max = MaxFloat32(max, i)
	}
	return max
}

// Combinations ...
func Combinations(words []string, l int) [][]string {
	indices := make([]int, len(words))
	for i := range indices {
		indices[i] = i
	}
	resultInt := combinationsInt(indices, l)

	result := make([][]string, len(resultInt))
	for resultIndex, intlist := range resultInt {
		temp := make([]string, len(intlist))
		for index, i := range intlist {
			temp[index] = words[i]
		}
		result[resultIndex] = temp
	}
	return result
}

func combinationsInt(iterable []int, r int) [][]int {
	pool := iterable
	n := len(pool)
	var finalResults [][]int
	if r > n {
		return finalResults
	}

	indices := make([]int, r)
	for i := range indices {
		indices[i] = i
	}

	result := make([]int, r)
	for i, el := range indices {
		result[i] = pool[el]
	}

	tmp := make([]int, len(result))
	copy(tmp, result)
	finalResults = append(finalResults, tmp)

	for {
		i := r - 1
		for ; i >= 0 && indices[i] == i+n-r; i-- {
		}

		if i < 0 {
			return finalResults
		}

		indices[i]++
		for j := i + 1; j < r; j++ {
			indices[j] = indices[j-1] + 1
		}

		for ; i < len(indices); i++ {
			result[i] = pool[indices[i]]
		}
		tmp := make([]int, len(result))
		copy(tmp, result)
		finalResults = append(finalResults, tmp)

	}

}

// Permutations ...
func Permutations(iterable []int, r int) [][]int {
	pool := iterable
	n := len(pool)
	var finalResults [][]int

	if r > n {
		return finalResults
	}

	indices := make([]int, n)
	for i := range indices {
		indices[i] = i
	}

	cycles := make([]int, r)
	for i := range cycles {
		cycles[i] = n - i
	}

	result := make([]int, r)
	for i, el := range indices[:r] {
		result[i] = pool[el]
	}

	tmp := make([]int, len(result))
	copy(tmp, result)
	finalResults = append(finalResults, tmp)

	for n > 0 {
		i := r - 1
		for ; i >= 0; i-- {
			cycles[i]--
			if cycles[i] == 0 {
				index := indices[i]
				for j := i; j < n-1; j++ {
					indices[j] = indices[j+1]
				}
				indices[n-1] = index
				cycles[i] = n - i
			} else {
				j := cycles[i]
				indices[i], indices[n-j] = indices[n-j], indices[i]

				for k := i; k < r; k++ {
					result[k] = pool[indices[k]]
				}

				tmp := make([]int, len(result))
				copy(tmp, result)
				finalResults = append(finalResults, tmp)

				break
			}
		}

		if i < 0 {
			return finalResults
		}

	}
	return finalResults

}

// GetAllPermutationString ... Get all unique permutation of string slice
func GetAllPermutationString(strSlice []string) [][]string {
	var res [][]string
	recursivePermutation(&res, strSlice, 0)
	return res
}

func recursivePermutation(res *[][]string, strSlice []string, start int) {
	if start >= len(strSlice) {
		t := make([]string, len(strSlice))
		copy(t, strSlice)
		*res = append(*res, t)
		return
	}
	for i := start; i < len(strSlice); i++ {
		strSlice[start], strSlice[i] = strSlice[i], strSlice[start]
		recursivePermutation(res, strSlice, start+1)
		strSlice[start], strSlice[i] = strSlice[i], strSlice[start]
	}
}

// GetFirstN ...
func GetFirstN(scores []string, count int) []string {
	if len(scores) <= count {
		return scores
	}
	result := []string{}
	for _, v := range scores {
		if len(result) >= count {
			return result
		}
		result = append(result, v)
	}
	return result
}

// GetFirstNInt64 ...
func GetFirstNInt64(scores []int64, count int) []int64 {
	if len(scores) <= count {
		return scores
	}
	result := []int64{}
	for _, v := range scores {
		if len(result) >= count {
			return result
		}
		result = append(result, v)
	}
	return result
}

// GetFirstNInt32 ...
func GetFirstNInt32(scores []int32, count int) []int32 {
	if len(scores) <= count {
		return scores
	}
	result := []int32{}
	for _, v := range scores {
		if len(result) >= count {
			return result
		}
		result = append(result, v)
	}
	return result
}

// GetFirstNInt ...
func GetFirstNInt(scores []int, count int) []int {
	if len(scores) <= count {
		return scores
	}
	result := []int{}
	for _, v := range scores {
		if len(result) >= count {
			return result
		}
		result = append(result, v)
	}
	return result
}

// Sum ...
func Sum(data []int32) int32 {
	s := int32(0)
	for _, d := range data {
		s += d
	}
	return s
}

// SumInt ...
func SumInt(data []int) int {
	s := 0
	for _, d := range data {
		s += d
	}
	return s
}

// GetKeys ... get keys from map[string]float32 whose value is higher than threshold
func GetKeys(dict map[string]float32, threshold float32) []string {
	array := []string{}
	for k, v := range dict {
		if v > threshold {
			array = append(array, k)
		}
	}
	return array
}

// GetFirstKey ... Get first key from map
func GetFirstKey(mp map[string]float32) string {
	for k := range mp {
		return k
	}
	return ""
}

// GetShortestKey ... Get the key whose length is the shortest from map
func GetShortestKey(mp map[string]float32) string {
	shortest := GetFirstKey(mp)
	for k := range mp {
		if len(shortest) > len(k) {
			shortest = k
		}
	}
	return shortest
}

// GetFirstPair ... Get first pair from map
func GetFirstPair(m map[string]float32) (string, float32) {
	for k, v := range m {
		return k, v
	}
	return "", 0
}

// GetFirstString ... Get first string
func GetFirstString(s []string) string {
	if len(s) > 0 {
		return s[0]
	}
	return ""
}

// JaccardSimString ... calculate jaccard similarity between two string slices
func JaccardSimString(s1, s2 []string) float32 {
	unionSet := GetUnionSetString(s1, s2)
	intersectSet := GetIntersectionSetString(s1, s2)
	jaccarSim := float32(len(intersectSet)) / float32(len(unionSet))
	return jaccarSim
}
