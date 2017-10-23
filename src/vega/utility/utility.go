package utility

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
)

func ParseRanges(ranges string) []string {
	numberSet := make(map[string]bool)
	numbers := []string{}
	byComma := strings.Split(ranges, ",")
	for _, r := range byComma {
		byDash := strings.Split(r, "-")
		if len(byDash) > 1 {
			x, y := byDash[0], byDash[1]
			i, _ := strconv.Atoi(x)
			j, _ := strconv.Atoi(y)
			if j < i {
				i, j = j, i
			}
			for k := i; k <= j; k++ {
				numberSet[fmt.Sprintf("%v", k)] = true
			}
		} else {
			numberSet[byDash[0]] = true
		}
	}

	for number, _ := range numberSet {
		numbers = append(numbers, number)
	}

	sort.Strings(numbers)
	return numbers
}

func ParseRangesUint16(ranges string) []uint16 {
	asStrings := ParseRanges(ranges)
	uints := []uint16{}

	for _, strnum := range asStrings {
		str, _ := strconv.ParseUint(strnum, 10, 16)
		uints = append(uints, uint16(str))
	}

	return uints
}

func ErrorsToStrings(errs ...error) []string {
	strings := make([]string, len(errs))

	for i := 0; i < len(errs); i++ {
		strings[i] = errs[i].Error()
	}

	return strings
}

func DataToMap(data interface{}) (map[string]interface{}, error) {
	dataMap := make(map[string]interface{})

	if bytes, err := json.Marshal(data); err != nil {
		return dataMap, err
	} else {
		if err := json.Unmarshal(bytes, &dataMap); err != nil {
			return dataMap, err
		} else {
			return dataMap, nil
		}
	}
}

func MapCopy(src *map[string]interface{}, dst *map[string]interface{}) {
	for key, srcValue := range *src {
		srcChildMap, srcIsMap := srcValue.(map[string]interface{})
		dstValue, ok := (*dst)[key]

		var dstChildMap map[string]interface{}
		var dstIsMap bool = false

		if ok {
			dstChildMap, dstIsMap = dstValue.(map[string]interface{})
			if srcIsMap && dstIsMap {
				MapCopy(&srcChildMap, &dstChildMap)
			}
		}

		if !ok || (!srcIsMap && !dstIsMap) {
			if srcValue != nil {
				valueCopy := srcValue
				(*dst)[key] = valueCopy
			}
		}

	}
}

func MapDecode(source io.Reader, target interface{}) error {
	tempBytes, err := json.Marshal(target)
	if err != nil {
		return err
	}

	tempMap := make(map[string]interface{})
	if err := json.Unmarshal(tempBytes, &tempMap); err != nil {
		return err
	}

	sourceMap := make(map[string]interface{})
	if err := json.NewDecoder(source).Decode(&sourceMap); err != nil {
		return err
	}

	MapCopy(&sourceMap, &tempMap)

	tempBytes, _ = json.Marshal(tempMap)
	if err := json.Unmarshal(tempBytes, &target); err != nil {
		return err
	}

	return nil
}

func RemoveStringsFromSlice(toRemove []string, from []string) (results, removed, notRemoved []string) {
	results = []string{}
	removed = []string{}
	notRemoved = []string{}
	removedSet := make(map[string]bool)

outer:
	for _, ofFrom := range from {
		for _, ofToRemove := range toRemove {
			if ofFrom == ofToRemove {
				removedSet[ofToRemove] = true
				continue outer
			}
		}
		results = append(results, ofFrom)
	}

	for _, remove := range toRemove {
		if removedSet[remove] {
			removed = append(removed, remove)
		} else {
			notRemoved = append(notRemoved, remove)
		}
	}

	return
}
