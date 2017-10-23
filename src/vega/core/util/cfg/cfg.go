// Copyright (c) 2016, Virtual Gateway Labs. All rights reserved.

// Package cfg provide APIs for parsing the unix style config file
package cfg

import (
	"errors"
	"io/ioutil"
	"os"
	"strings"
	"syscall"
)

type (
	Config struct {
		kvPairs []KVPair
		file    *os.File
	}

	KVPair struct {
		Key    string
		Values []string
	}
)

// Load config file and parse the content to the KV pairs. The config file will
// keep opened until Close() is called. All the changes made will sync to the
// disk only when Sync() or Close() is called
func LoadConfig(file_path string) (cfg *Config, err error) {
	file, err := os.OpenFile(file_path, syscall.O_RDWR, 0644)
	if err != nil {
		return
	}

	file_data, err := ioutil.ReadAll(file)
	if err != nil {
		file.Close()
		return
	}

	cfg = new(Config)
	cfg.file = file
	lines := strings.Split(string(file_data), "\n")

	for _, line := range lines {
		line = strings.TrimSuffix(line, "\r")
		cfg.AddLine(line)
	}

	return
}

// Replace the KV pair in the list to the new one
func (cfg *Config) Replace(kvPairOld KVPair, kvPairNew KVPair) (err error) {

	found, index := cfg.findMatch(kvPairOld)
	if found {
		cfg.kvPairs[index] = kvPairNew
	} else {
		err = errors.New("KV Pair not found")
	}

	return
}

// Add a KV pair to the list, if duplicates, the action will be ignored without
// returning errors
func (cfg *Config) AddKVPair(kvPair KVPair) (err error) {

	if len(kvPair.Key) == 0 {
		err = errors.New("The key in the input KV pair is empty")
		return
	}

	if strings.HasPrefix(kvPair.Key, "#") {
		err = errors.New("The key cannot start with '#'")
		return
	}

	found, _ := cfg.findMatch(kvPair)
	if !found {
		cfg.kvPairs = append(cfg.kvPairs, kvPair)
	}

	return
}

// Add a Key and value string to the list, if duplicates, the action will be
// ignored without returning errors
func (cfg *Config) AddStrings(key string, values ...string) (err error) {

	kv_pair := KVPair{key, values}
	cfg.AddKVPair(kv_pair)
	return
}

// Add a line to the list, if duplicates, the action will be ignored without
// returning errors
func (cfg *Config) AddLine(line string) (err error) {
	if len(line) != 0 && !strings.HasPrefix(line, "#") {
		fields := strings.Fields(line)
		kv_pair := KVPair{}

		for idx, field := range fields {
			if idx == 0 {
				kv_pair.Key = field
			} else {
				kv_pair.Values = append(kv_pair.Values, field)
			}
		}

		cfg.AddKVPair(kv_pair)
	} else {
		err = errors.New("Empty line or the line is commented")
	}

	return
}

// Get KV pair by the Key, could return multiple results. If none is found, will
// return a empty slice
func (cfg Config) GetKVPair(key string) (kvPairs []KVPair, err error) {

	if key == "" {
		err = errors.New("Key string is empty")
	}

	for _, kv_pair := range cfg.kvPairs {
		if kv_pair.Key == key {
			kvPairs = append(kvPairs, kv_pair)
		}
	}

	return
}

// Get option line by the key, could return multiple results. If none is found, will
// return a empty slice
func (cfg Config) GetLine(key string) (lines []string, err error) {

	if key == "" {
		err = errors.New("Key string is empty")
	}

	for _, kv_pair := range cfg.kvPairs {
		if kv_pair.Key == key {
			line := strings.Join(kv_pair.Values, " ")
			line = strings.Join([]string{kv_pair.Key, line}, " ")
			lines = append(lines, line)
		}
	}

	return
}

// Get value by the key, could return multiple results. If none is found, will
// return a empty slice
func (cfg Config) GetValues(key string) (values []string, err error) {
	values = []string{}

	if key == "" {
		err = errors.New("Key string is empty")
	}

	for _, kv_pair := range cfg.kvPairs {
		if kv_pair.Key == key {
			value := strings.Join(kv_pair.Values, " ")
			values = append(values, value)
		}
	}

	return
}

// Return how many
func (cfg *Config) Find(key string) (num int) {
	for _, kv_pair := range cfg.kvPairs {
		if kv_pair.Key == key {
			num++
		}
	}

	return
}

// Return index of the KV pair that found in the list
func (cfg *Config) findMatch(kvPair KVPair) (found bool, index int) {
	for idx, kv_pair := range cfg.kvPairs {
		if kv_pair.Key == kvPair.Key &&
			strings.Join(kv_pair.Values, " ") == strings.Join(kvPair.Values, " ") {
			found = true
			index = idx
			break
		}
	}

	return
}

// Get KV pair by the Key, could return multiple results
func (cfg Config) GetAll() (kvPairs []KVPair, err error) {
	if cfg.file == nil {
		err = errors.New("No file loaded")
	}

	kvPairs = cfg.kvPairs

	return
}

// Empty The list
func (cfg *Config) DeleteAll() (err error) {
	cfg.kvPairs = []KVPair{}
	return
}

// Delete KV pair by the given key. All the matched pair will be deleted
func (cfg *Config) DeleteByKey(key string) (err error) {

	err = cfg.Delete(KVPair{key, []string{}})
	return
}

// Delete KV pair by the given KV pair. if the value is empty, will delete all
// KV pairs which has matching key
func (cfg *Config) Delete(kvPair KVPair) (err error) {

	var match bool
	var del_idx []int

	if kvPair.Key == "" {
		err = errors.New("Key in the KV pair is empty")
	}

	for idx, kv_pair := range cfg.kvPairs {
		if kv_pair.Key == kvPair.Key {
			if len(kvPair.Values) == 0 {
				match = true
			} else if len(kv_pair.Values) == len(kvPair.Values) {
				for i, value := range kvPair.Values {
					if value != kv_pair.Values[i] {
						match = false
						break
					} else {
						match = true
					}
				}
			}

			if match {
				del_idx = append(del_idx, idx)
			}
		}
	}

	dels_rev := []int{}
	for i := len(del_idx) - 1; i >= 0; i-- {
		dels_rev = append(dels_rev, del_idx[i])
	}

	// delete the element from the slice
	for _, del := range dels_rev {
		cfg.kvPairs = append(cfg.kvPairs[:del], cfg.kvPairs[del+1:]...)
	}

	return
}

// Save config back to file
func (cfg *Config) Sync() (err error) {

	var data string

	for _, kv_pair := range cfg.kvPairs {
		value := strings.Join(kv_pair.Values, " ")
		line := kv_pair.Key + " " + value
		data += line + "\n"
	}

	err = cfg.file.Truncate(0)
	if err != nil {
		return
	}

	_, err = cfg.file.Seek(0, 0)
	if err != nil {
		return
	}

	_, err = cfg.file.WriteString(data)
	if err != nil {
		return
	}

	err = cfg.file.Sync()
	if err != nil {
		return
	}

	return
}

// Close the config file
func (cfg *Config) Close() (err error) {
	err = cfg.Sync()
	if err != nil {
		return
	}

	err = cfg.file.Close()
	if err != nil {
		return
	}

	cfg.file = nil

	return
}
