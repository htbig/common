package cfg

import (
	"testing"
)

const (
	cfg_file     = "test.conf"
	cfg_file_bad = "/usr/bin/cat"

	Key   = "abcdefg"
	value = "1234"
	line  = "abcdefg 1234"
)

func TestConfig(t *testing.T) {

	t.Log("[case] Test parsing config file")
	cfg, err := LoadConfig(cfg_file)
	if err == nil {
		for idx, kv_pair := range cfg.kvPairs {
			t.Log("[info] client config:", idx, kv_pair.Key, kv_pair.Values)
		}
	} else {
		t.Error("[err] Read parsing config:", err)
		return
	}

	t.Log("[case] Test add KV-pair")
	kv_pair := KVPair{Key, []string{value}}
	err = cfg.AddKVPair(kv_pair)
	if err == nil {
		t.Log("[info] Add KV-pair successful")
	} else {
		t.Error("[err] Add KV-pair:", err)
	}

	t.Log("[case] Test add KV-pair from string")
	err = cfg.AddStrings(Key, value)
	if err == nil {
		t.Log("[info] Add string successful")
	} else {
		t.Error("[err] Add string:", err)
	}

	t.Log("[case] Test add KV-pair")
	err = cfg.AddLine(line)
	if err == nil {
		t.Log("[info] Add line successful")
	} else {
		t.Error("[err] Add line:", err)
	}

	t.Log("[case] Test find KV-pair")
	kv_pairs, err := cfg.GetKVPair(Key)
	if err == nil {
		t.Log("[info] Add line successful")
		for idx, kv_pair := range kv_pairs {
			t.Log("[info] client config:", idx, kv_pair.Key, kv_pair.Values)
		}
	} else {
		t.Error("[err] Add line:", err)
	}

	t.Log("[case] Test delete KV-pair")
	err = cfg.Delete(kv_pair)
	if err == nil {
		t.Log("[info] Add line successful")
	} else {
		t.Error("[err] Add line:", err)
	}

	for idx, kv_pair := range cfg.kvPairs {
		t.Log("[info] client config:", idx, kv_pair.Key, kv_pair.Values)
	}

	cfg.Close()

	return
}
