package config

import (
	"github.com/voodooEntity/archivist"
	"strconv"
)

var data = map[string]string{
	"STORAGE_DIR":   "storage/",
	"PLUGIN_DIR":    "plugins/",
	"KNOWLEDGE_DIR": "knowledge/",
}

func Set(key string, val string) bool {
	if _, ok := data[key]; ok {
		if 0 < len(val) {
			data[key] = val
			return true
		}
	}
	return false
}

func Get(key string) string {
	if val, ok := data[key]; ok {
		return val
	}
	return ""
}

func GetInt(key string) (int, bool) {
	if val, ok := data[key]; ok {
		intVal, err := strconv.Atoi(val)
		if nil != err {
			archivist.Error("Trying to cast config value of key :'" + key + "' as int resulting in error: " + err.Error())
			return 0, false
		}
		return intVal, true
	}
	return 0, false
}
