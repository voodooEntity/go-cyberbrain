package util

import (
	"crypto/rand"
	"encoding/base64"
	"github.com/voodooEntity/gits"
	"github.com/voodooEntity/gits/src/query"
	"github.com/voodooEntity/gits/src/transport"
)

func IsAlive(gitsInstance *gits.Gits) bool {
	qry := query.New().Read("AI").Match(
		"Value",
		"==",
		"Bezel",
	).Match(
		"Properties.State",
		"==",
		"Alive",
	)
	ret := gitsInstance.Query().Execute(qry)
	if 0 < ret.Amount {
		return true
	}
	return false
}

func Terminate(gitsInstance *gits.Gits) bool {
	qry := query.New().Update("AI").Match(
		"Value",
		"==",
		"Bezel",
	).Set(
		"Properties.State",
		"Dead",
	)
	ret := gitsInstance.Query().Execute(qry)
	if 0 < ret.Amount {
		return true
	}
	return false
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

func ResolveEntityField(entity transport.TransportEntity, field string) string {
	switch field {
	case "Value":
		return entity.Value
	case "Context":
		return entity.Context
	default:
		if len(field) > 11 && field[:11] == "Properties." {
			if val, ok := entity.Properties[field[11:]]; ok {
				return val
			}
		}
	}
	return ""
}
