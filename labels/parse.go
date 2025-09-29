package labels

import (
	"alati_projekat/model"
	"errors"
	"strings"
)

func Parse(s string) (map[string]string, error) {
	result := map[string]string{}
	if strings.TrimSpace(s) == "" {
		return result, nil
	}
	parts := strings.Split(s, ";")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		kv := strings.SplitN(part, ":", 2)
		if len(kv) != 2 {
			return nil, errors.New("invalid label format, expected k:v;k2:v2")
		}
		k := strings.TrimSpace(kv[0])
		v := strings.TrimSpace(kv[1])
		if k == "" || v == "" {
			return nil, errors.New("label key/value must be non-empty")
		}
		result[k] = v
	}
	return result, nil
}
func HasAll(cfg model.Configuration, want map[string]string) bool {
	if len(want) == 0 {
		return true
	}
	m := cfg.LabelsMap()
	for k, v := range want {
		if mv, ok := m[k]; !ok || mv != v {
			return false
		}
	}
	return true
}
