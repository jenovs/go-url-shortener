package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

func init() {
	abs, _ := filepath.Abs("config/config.json")
	raw, err := ioutil.ReadFile(abs)

	if err != nil {
		return
	}

	var conf map[string]string
	json.Unmarshal(raw, &conf)

	for i, k := range conf {
		os.Setenv(i, k)
	}
}
