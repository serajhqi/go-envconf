package envconf

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
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

		raw, ok := os.LookupEnv(envKey)
		if !ok {
			raw, ok = dotEnvValues[envKey]
		}
		if !ok {
			raw, ok = field.Tag.Lookup("default")
		}
		if !ok {
			return fmt.Errorf("envconf: missing required value for field %s (env %s)", field.Name, envKey)
		}

		if err := setFieldFromString(field, fv, envKey, raw); err != nil {
			return err
		}
	}

	return nil
}

func setFieldFromString(field reflect.StructField, fv reflect.Value, envKey, raw string) error {
	switch fv.Kind() {
	case reflect.String:
		fv.SetString(raw)
		return nil
	case reflect.Bool:
		v, err := strconv.ParseBool(raw)
		if err != nil {
			return parseError(field.Name, envKey, raw, field.Type, err)
		}
		fv.SetBool(v)
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		durationType := reflect.TypeOf(time.Duration(0))
		if field.Type == durationType {
			d, err := time.ParseDuration(raw)
			if err != nil {
				return parseError(field.Name, envKey, raw, field.Type, err)
			}
			fv.SetInt(int64(d))
			return nil
		}
		v, err := strconv.ParseInt(raw, 10, fv.Type().Bits())
		if err != nil {
			return parseError(field.Name, envKey, raw, field.Type, err)
		}
		fv.SetInt(v)
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		v, err := strconv.ParseUint(raw, 10, fv.Type().Bits())
		if err != nil {
			return parseError(field.Name, envKey, raw, field.Type, err)
		}
		fv.SetUint(v)
		return nil
	case reflect.Float32, reflect.Float64:
		v, err := strconv.ParseFloat(raw, fv.Type().Bits())
		if err != nil {
			return parseError(field.Name, envKey, raw, field.Type, err)
		}
		fv.SetFloat(v)
		return nil
	default:
		return fmt.Errorf("envconf: field %s has unsupported kind %s", field.Name, field.Type.Kind())
	}
}

func parseError(fieldName, envKey, raw string, targetType reflect.Type, err error) error {
	return fmt.Errorf("envconf: failed parsing field %s (env %s, value %q, target %s): %w", fieldName, envKey, raw, targetType, err)
}
