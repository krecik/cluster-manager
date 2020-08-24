package main

import "log"

func fallbackBoolWithDefault(defaultValue bool, values ...*bool) bool {
	for _, v := range values {
		if v != nil {
			return *v
		}
	}
	return defaultValue
}

func fallbackStringWithDefault(defaultValue string, values ...*string) string {
	for _, v := range values {
		if v != nil {
			return *v
		}
	}
	return defaultValue
}

func fallbackString(values ...*string) string {
	for _, v := range values {
		if v != nil && *v != "" {
			return *v
		}
	}
	fatal("you must provide a value")
	return ""
}

// based on https://github.com/helm/helm/blob/cd50d0c3621ad91b3848f14b7ef3a8d6aa29d2c9/pkg/chartutil/coalesce.go#L37

func isTable(v interface{}) bool {
	_, ok := v.(map[interface{}]interface{})
	return ok
}

func mergeStructs(dst, src map[interface{}]interface{}) map[interface{}]interface{} {
	if src == nil {
		return dst
	}
	if dst == nil {
		return src
	}
	for key, val := range src {
		if dv, ok := dst[key]; ok && dv == nil {
			delete(dst, key)
		} else if !ok {
			dst[key] = val
		} else if isTable(val) {
			if isTable(dv) {
				mergeStructs(dv.(map[interface{}]interface{}), val.(map[interface{}]interface{}))
			} else {
				log.Printf("warning: cannot overwrite table with non table for %s (%v)", key, val)
			}
		} else if isTable(dv) {
			log.Printf("warning: destination for %s is a table. Ignoring non-table value %v", key, val)
		} else {
			dst[key] = dv
		}
	}
	return dst
}

func mergeDicts(dicts ...map[string]string) map[string]string {
	output := map[string]string{}
	for _, dict := range dicts {
		for k, v := range dict {
			output[k] = v
		}
	}
	return output
}
