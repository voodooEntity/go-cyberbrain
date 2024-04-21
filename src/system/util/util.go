package util

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"github.com/voodooEntity/archivist"
	"github.com/voodooEntity/gits/src/query"
	"github.com/voodooEntity/go-cyberbrain-plugin-interface/src/interfaces"
	"os"
	"path/filepath"
	"plugin"
	"strings"
)

func GetAvailablePlugins(pluginsDir string) []string {
	dirs, err := DirectoryWalkMatch(pluginsDir, "*.so")
	if err != nil {
		archivist.Fatal("Could not read given plugin directory ", pluginsDir)
		os.Exit(0)
	}
	var ret []string
	for _, f := range dirs {
		ret = append(ret, strings.TrimPrefix(f, pluginsDir))
	}
	return ret
}

func LoadPlugin(pluginDirectory string, plug string) (interfaces.PluginInterface, error) {
	plugInst, err := plugin.Open(pluginDirectory + plug)
	if err != nil {
		return nil, errors.New("Could not load plugin '" + plug + "' with error: " + err.Error())
	} else {
		sym, err := plugInst.Lookup("Export")
		if err != nil {
			return nil, errors.New("Plugin '" + plug + "' doesnt export neccesary Export var with error: " + err.Error())
		} else {
			typedSymbol, ok := sym.(interfaces.PluginInterface)
			if !ok {
				return nil, errors.New("Plugin '" + plug + "' does not match the Plugin interfaces of bezel ")
			} else {
				return typedSymbol.New(), nil
			}
		}
	}
}

func IsActive() bool {
	qry := query.New().Read("AI").Match(
		"Value",
		"==",
		"Bezel",
	).Match(
		"Properties.State",
		"==",
		"Alive",
	)
	ret := query.Execute(qry)
	if 0 < ret.Amount {
		return true
	}
	return false
}

func Shutdown() bool {
	qry := query.New().Update("AI").Match(
		"Value",
		"==",
		"Bezel",
	).Set(
		"Properties.State",
		"Dead",
	)
	ret := query.Execute(qry)
	if 0 < ret.Amount {
		return true
	}
	return false
}

func DirectoryWalkMatch(root, pattern string) ([]string, error) {
	var matches []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if matched, err := filepath.Match(pattern, filepath.Base(path)); err != nil {
			return err
		} else if matched || "*" == pattern {
			matches = append(matches, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return matches, nil
}

func StringInArray(haystack []string, needle string) bool {
	for _, val := range haystack {
		if needle == val {
			return true
		}
	}
	return false
}

func UniqueID() string {
	// Generate 16 random bytes
	bytes := make([]byte, 16)
	_, err := rand.Read(bytes)
	if err != nil {
		panic(err) // Handle the error according to your needs
	}

	// Encode the random bytes to base64
	encoded := base64.RawURLEncoding.EncodeToString(bytes)

	// Return the unique ID
	return encoded
}

func CopyStringStringMap(data map[string]string) map[string]string {
	ret := make(map[string]string)
	for k, v := range data {
		ret[k] = v
	}
	return ret
}
