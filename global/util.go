package global

import (
	"Message-Generator/stats"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"regexp"
	"strings"
)

// FastRemove removes an index from a slice of strings without maintaining order
func FastRemove(s []string, i int) []string {
	s[i] = s[len(s)-1] // Copy last element to index i
	s = s[:len(s)-1]   // Truncate slice
	return s
}

// PrettyPrint prints out a map in a pretty format.
func PrettyPrint(v interface{}) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err == nil {
		fmt.Println(string(b))
	}
}

func RandomNumber(min, max int) int {
	var result int
	switch {
	case min > max:
		// Fail with error
		return result
	case max == min:
		result = max
	case max > min:
		maxRand := max - min
		b, err := rand.Int(rand.Reader, big.NewInt(int64(maxRand)))
		if err != nil {
			return result
		}
		result = min + int(b.Int64())
	}
	return result
}

func PickRandomFromSlice(slice []string) string {
	return slice[RandomNumber(0, len(slice))]
}

func RecoverFullName(functionName string) {
	if r := recover(); r != nil {
		stats.Log("recovered '" + functionName)
	}
}

func LoadChannels() {
	jsonFile, _ := os.Open("./global/channels.json")
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal(byteValue, &Directives)
}

func FastRemoveDirective(s []Directive, i int) []Directive {
	s[i] = s[len(s)-1] // Copy last element to index i
	s = s[:len(s)-1]   // Truncate slice
	return s
}

func UpdateChannels(mode string, channel Directive) error {
	if mode == "update" || mode == "remove" {
		for i, directive := range Directives {
			if directive.ChannelName == channel.ChannelName {
				Directives = FastRemoveDirective(Directives, i)
				break
			}
		}
	}

	if mode == "update" || mode == "add" {
		Directives = append(Directives, channel)
	}

	return SaveChannels()
}

func SaveChannels() error {
	file, err := json.MarshalIndent(Directives, "", " ")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile("./global/channels.json", file, 0644)
	if err != nil {
		return err
	}
	return err
}

func LoadRegex() {
	jsonFile, err := os.Open("./global/regex.json")
	if err != nil {
		panic(err)
	}
	defer jsonFile.Close()
	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		panic(err)
	}
	json.Unmarshal(byteValue, &RegexList)

	r := strings.Join(RegexList, "|")
	Regex = regexp.MustCompile(r)
}

func UpdateRegex() error {
	r := strings.Join(RegexList, "|")
	Regex = regexp.MustCompile(r)

	return SaveRegex()
}

func SaveRegex() error {
	file, err := json.MarshalIndent(RegexList, "", " ")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile("./global/regex.json", file, 0644)
	if err != nil {
		return err
	}
	return err
}

func LoadBannedUsers() {
	jsonFile, err := os.Open("./global/banned-users.json")
	if err != nil {
		panic(err)
	}
	defer jsonFile.Close()
	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		panic(err)
	}
	json.Unmarshal(byteValue, &BannedUsers)
}

func SaveBannedUsers() error {
	file, err := json.MarshalIndent(BannedUsers, "", " ")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile("./global/banned-users.json", file, 0644)
	if err != nil {
		return err
	}
	return err
}
