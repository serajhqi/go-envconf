package envconf

import (
	"fmt"
	"os"
	"reflect"
	"strings"
)

func validateTarget(target any) (reflect.Type, error) {
	rv := reflect.ValueOf(target)
	if rv.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("envconf: target must be pointer to struct, got %T", target)
	}

	elem := rv.Elem()
	if elem.Kind() != reflect.Struct {
		return nil, fmt.Errorf("envconf: target must point to struct, got %s", elem.Kind())
	}

	return rv.Type(), nil
}

func populateStruct(target any, dotEnvValues map[string]string) error {
	rv := reflect.ValueOf(target).Elem()
	rt := rv.Type()

	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		envKey := strings.TrimSpace(field.Tag.Get("env"))
		if envKey == "" {
			continue
		}

		fv := rv.Field(i)
		if !fv.CanSet() {
			return fmt.Errorf("envconf: field %s is not settable", field.Name)
		}
		if field.Type.Kind() != reflect.String {
			return fmt.Errorf("envconf: field %s has unsupported kind %s", field.Name, field.Type.Kind())
		}

		if value, ok := os.LookupEnv(envKey); ok {
			fv.SetString(value)
			continue
		}

		if value, ok := dotEnvValues[envKey]; ok {
			fv.SetString(value)
			continue
		}

		if def, ok := field.Tag.Lookup("default"); ok {
			fv.SetString(def)
			continue
		}

		return fmt.Errorf("envconf: missing required value for field %s (env %s)", field.Name, envKey)
	}

	return nil
}
