package translator

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"os"

	"gopkg.in/yaml.v3"
)

// EmbedFSLocalesYAML is an embedded file system that contains the locale YAML files.
//
//go:embed locales/*/*.yaml
var EmbedFSLocalesYAML embed.FS

// fromMap recursively flattens a map of any type into a map of strings.
// It prefixes each key with the given prefix.
func fromMap(source map[string]any, dest map[string]string, prefix string) {
	if len(prefix) > 0 {
		prefix += ":"
	}
	for k, v := range source {
		if recMap, ok := v.(map[string]any); ok {
			fromMap(recMap, dest, prefix+k)
		} else if recMap, ok := v.(map[string]string); ok {
			dest[prefix+k] = recMap[prefix+k]
		} else if strVal, ok := v.(string); ok {
			dest[prefix+k] = strVal
		}
	}
}

// LocalesFromFS reads the locale data from a file system and returns a map of locale data.
func LocalesFromFS(fsys fs.FS) (map[string]map[string]string, error) {
	rootDir, err := fs.ReadDir(fsys, "locales")
	if err != nil {
		return nil, fmt.Errorf("%w:%s", err, "can't read locales")
	}

	langMap := make(map[string]map[string]string)
	for _, locale := range rootDir {
		file, err := fsys.Open(fmt.Sprintf("locales/%s/data.yaml", locale.Name()))
		if err != nil && os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("%w:%s", err, fmt.Sprintf("can't read locales/%s/data.yaml", locale.Name()))
		}
		fileBytes, err := io.ReadAll(file)
		localeData := make(map[string]any)
		err = yaml.Unmarshal(fileBytes, &localeData)
		if err != nil {
			return nil, fmt.Errorf("%w:%s", err, fmt.Sprintf("can't parse locales/%s/data.yaml", locale.Name()))
		}
		dest := make(map[string]string)
		fromMap(localeData, dest, "")
		langMap[locale.Name()] = dest
	}
	return langMap, nil
}
