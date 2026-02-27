package envconf

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
)

func loadDotEnv(opts loaderOptions) (map[string]string, error) {
	values := make(map[string]string)
	if !opts.dotEnvEnabled {
		return values, nil
	}

	f, err := os.Open(opts.dotEnvPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return values, nil
		}
		return nil, fmt.Errorf("envconf: failed to open %s: %w", opts.dotEnvPath, err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		eq := strings.IndexByte(line, '=')
		if eq <= 0 {
			return nil, fmt.Errorf("envconf: malformed .env line %d", lineNo)
		}

		key := strings.TrimSpace(line[:eq])
		if key == "" {
			return nil, fmt.Errorf("envconf: malformed .env line %d", lineNo)
		}
		value := strings.TrimSpace(line[eq+1:])
		if len(value) >= 2 {
			if (value[0] == '\'' && value[len(value)-1] == '\'') || (value[0] == '"' && value[len(value)-1] == '"') {
				value = value[1 : len(value)-1]
			}
		}
		values[key] = value
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("envconf: failed reading %s: %w", opts.dotEnvPath, err)
	}

	return values, nil
}
