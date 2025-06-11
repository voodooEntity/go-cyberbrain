package util

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/voodooEntity/gits"
	"github.com/voodooEntity/gits/src/query"
	"github.com/voodooEntity/gits/src/transport"
	"reflect"
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

// ### keep for no for debug reasons
func structHasMethod(v interface{}, methodName string) bool {
	// Get the reflect.Value of the instance.
	// We need to ensure we're working with the addressable value if the methods
	// are defined on pointer receivers (which is common for structs that hold state).
	val := reflect.ValueOf(v)

	// If the value is a pointer, get the element it points to.
	// This handles cases where 'v' is already a pointer (e.g., *MyAction).
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// If after dereferencing, it's still not a struct or valid value, return false.
	// This catches cases where 'v' might be nil or a basic type.
	if !val.IsValid() {
		return false
	}

	// If the methods are defined on a pointer receiver (e.g., func (m *MyAction) SetData(...)),
	// and 'val' is currently the struct value, we need to get its address.
	// MethodByName will only find methods on the exact type it's reflecting upon.
	// If the original 'v' was a value type and its methods are on a pointer receiver,
	// we need to get the address of 'val' to find those methods.
	if val.Kind() == reflect.Struct && !val.CanAddr() {
		// If it's a struct and not addressable, we can't get a pointer to it for method lookup.
		// This scenario means 'v' was passed as a value, and its methods are on a pointer receiver.
		// To fix this, you would typically pass the *address* of the struct (e.g., &myInstance)
		// to `HasMethod` initially. For the purpose of this function, if it's not addressable,
		// we cannot proceed to find pointer receiver methods.
		fmt.Printf("Warning: Passed value of type %T is not addressable; cannot find methods on pointer receivers.\n", v)
		return false
	}

	// Get the method by its name.
	// If the concrete methods are defined on pointer receivers, val needs to be addressable.
	// For instance, if SetData is func (m *MyAction) SetData(...), then `reflect.ValueOf(myActionInstance).MethodByName("SetData")`
	// won't find it if `myActionInstance` is a value, but `reflect.ValueOf(&myActionInstance).MethodByName("SetData")` will.
	// By using `val.Addr()` if `val` is a struct and `val.CanAddr()`, we ensure we're looking on the pointer type if applicable.
	var method reflect.Value
	if val.Kind() == reflect.Struct && val.CanAddr() {
		method = val.Addr().MethodByName(methodName)
	} else {
		method = val.MethodByName(methodName)
	}

	// Check if the method was found and is valid.
	return method.IsValid()
}
