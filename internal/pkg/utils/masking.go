// Package utils provides common functional helpers such as data sanitization
// and sensitive information masking before it reaches the logging or telemetry layers.
package utils

import (
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"strings"
)

const (
	// MaxFieldSize defines the maximum allowed size (in bytes) for a string field in logs.
	// If exceeded, the value is replaced with a warning message to prevent log bloat.
	MaxFieldSize = 2048
	// maxDepth limits recursion to prevent stack overflow on deeply nested or circular objects.
	maxDepth = 3
)

// sensitiveKeys defines a list of keywords identified as confidential.
// Any field containing these keywords will have its value redacted.
var sensitiveKeys = []string{"password", "token", "secret", "otp", "credential", "authorization"}

// MaskSensitive processes any data type (struct, map, slice, string) to:
// 1. Redact sensitive values based on predefined keys.
// 2. Enforce size limits on large string fields.
// 3. Recursively mask nested JSON strings found within fields.
//
// Example:
//
//	maskedBody := utils.MaskSensitive(req.Body)
func MaskSensitive(data any) any {
	return maskRecursive(data, 0)
}

// MaskHttpHeaders filters HTTP headers based on a whitelist of allowed keys
// and redacts sensitive information like Authorization tokens.
// It returns a flattened map[string]string for cleaner log representation.
func MaskHttpHeaders(headers map[string][]string) map[string]string {
	// Whitelist approach to ensure only non-sensitive infra headers are logged
	allowed := map[string]bool{
		"authorization": true,
		"x-request-id":  true,
		"x-user-id":     true,
		"user-agent":    true,
		"content-type":  true,
		"accept":        true,
	}

	out := make(map[string]string)
	for k, v := range headers {
		key := strings.ToLower(k)
		if !allowed[key] {
			continue
		}

		val := strings.Join(v, ", ")
		if IsSensitiveKey(key) {
			out[k] = "******** [REDACTED]"
		} else {
			out[k] = val
		}
	}
	return out
}

// IsSensitiveKey checks if a given key name contains any sensitive keywords.
// It is case-insensitive and matches substrings (e.g., "access_token" matches "token").
func IsSensitiveKey(key string) bool {
	lowerKey := strings.ToLower(key)
	return slices.ContainsFunc(sensitiveKeys, func(s string) bool {
		return strings.Contains(lowerKey, s)
	})
}

// ContainsSensitiveToken provides a quick check for sensitive tokens within a raw string.
func ContainsSensitiveToken(msg string) bool {
	lower := strings.ToLower(msg)
	for _, word := range sensitiveKeys {
		if strings.Contains(lower, word) {
			return true
		}
	}
	return false
}

func maskRecursive(data any, depth int) any {
	if data == nil || depth > maxDepth {
		return data
	}

	val := reflect.ValueOf(data)

	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return nil
		}
		val = val.Elem()
	}

	switch val.Kind() {
	case reflect.String:
		return maskString(val.String(), depth)

	case reflect.Slice, reflect.Array:
		return maskSlice(val, depth)

	case reflect.Map:
		return maskMap(val, depth)

	case reflect.Struct:
		b, _ := json.Marshal(data)
		var m any
		if err := json.Unmarshal(b, &m); err == nil {
			return maskRecursive(m, depth)
		}
		return data

	default:
		return data
	}
}

func maskSlice(val reflect.Value, depth int) []any {
	limit := min(val.Len(), 10)
	newSlice := make([]any, val.Len())
	for i := 0; i < val.Len(); i++ {
		if i < limit {
			newSlice[i] = maskRecursive(val.Index(i).Interface(), depth+1)
		} else {
			newSlice[i] = val.Index(i).Interface()
		}
	}
	return newSlice
}

func maskString(v string, depth int) any {
	trimmed := strings.TrimSpace(v)
	if len(trimmed) == 0 {
		return v
	}

	if len(trimmed) > MaxFieldSize {
		return fmt.Sprintf("[field size %d bytes, too large to log]", len(trimmed))
	}
	if (strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[")) && depth < maxDepth {
		var nested any
		if err := json.Unmarshal([]byte(trimmed), &nested); err == nil {
			masked := maskRecursive(nested, depth+1)
			if b, err := json.Marshal(masked); err == nil {
				return string(b)
			}
		}
	}

	lower := strings.ToLower(trimmed)
	for _, word := range sensitiveKeys {
		if strings.Contains(lower, word) {
			return "******** [REDACTED]"
		}
	}

	return v
}

func maskMap(val reflect.Value, depth int) map[string]any {
	newMap := make(map[string]any, val.Len())
	iter := val.MapRange()
	for iter.Next() {
		k := iter.Key().String()
		v := iter.Value().Interface()

		if IsSensitiveKey(k) {
			newMap[k] = "******** [REDACTED]"
			continue
		}
		newMap[k] = maskRecursive(v, depth+1)
	}
	return newMap
}
