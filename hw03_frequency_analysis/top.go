package hw03frequencyanalysis

import (
	"sort"
	"strings"
)

func Top10(str string) []string {
	words := strings.Fields(str)
	freq := make(map[string]int)
	result := make([]string, 0)
	for _, word := range words {
		if _, ok := freq[word]; ok {
			freq[word]++
		} else {
			freq[word] = 1
		}
	}
	for cnt := 10; cnt > 0; {
		wordsMaxFreq := getMaxFreqWords(freq)
		cnt, result = addMaxFreqWordsToRes(wordsMaxFreq, cnt, result)
		if len(freq) == 0 {
			return result
		}
	}

	return result
}

func getMaxFreqWords(freq map[string]int) []string {
	maxValue := -1
	res := make([]string, 0)
	for _, value := range freq {
		if value > maxValue {
			maxValue = value
		}
	}
	for key, value := range freq {
		if value == maxValue {
			res = append(res, key)
			delete(freq, key)
		}
	}
	return res
}

func addMaxFreqWordsToRes(wordsMaxFreq []string, cnt int, result []string) (int, []string) {
	resArr := make([]string, len(wordsMaxFreq))
	sort.Strings(wordsMaxFreq)
	if len(wordsMaxFreq) <= cnt {
		resArr = append(result, wordsMaxFreq...)
		return cnt - len(wordsMaxFreq), resArr
	}
	return 0, append(result, wordsMaxFreq[:cnt]...)
}
