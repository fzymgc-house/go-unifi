package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// fieldsBlobStart marks a pretty-printed JSON object at the indentation the field
// definitions use. The native image stores each api/fields file as one such blob.
var fieldsBlobStart = regexp.MustCompile(`\{\n {2,4}"`)

// fieldsBlobMaxLen bounds one definition. The largest, Setting.json, is under 100 KiB.
const fieldsBlobMaxLen = 1 << 20

// fieldsNameMinScore is the Jaccard similarity of the top-level keys below which a
// blob is not given the reference file's name.
const fieldsNameMinScore = 0.4

// fieldsNameMinKeys is the number of top-level keys below which a blob is not named:
// one key does not tell two definitions apart.
const fieldsNameMinKeys = 2

// ExtractFromNativeImage populates fieldsDir with the api/fields JSON definitions
// carved out of a Network native-image binary, plus the custom overrides and the
// Setting.json split. Network 11 on UniFi OS ships no .deb and no jar: the
// definitions sit inside /usr/lib/unifi/lib/unifi as pretty-printed JSON, with
// their file names stored elsewhere in the image. Each blob is named after the
// file in referenceDir, a fields directory of an earlier version, whose top-level
// keys it shares the most. A blob that matches no reference file is reported and
// left out.
func ExtractFromNativeImage(binaryPath, referenceDir, fieldsDir string) error {
	cached, err := isValidFieldsCache(fieldsDir)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if cached {
		return nil
	}

	data, err := os.ReadFile(binaryPath)
	if err != nil {
		return fmt.Errorf("unable to read native image: %w", err)
	}

	reference, err := referenceKeySets(referenceDir)
	if err != nil {
		return err
	}

	tmpDir := fieldsDir + ".tmp"
	if err := os.RemoveAll(tmpDir); err != nil {
		return err
	}
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	named, unnamed := nameFieldsBlobs(carveFieldsBlobs(data), reference)
	for name, blob := range named {
		if err := os.WriteFile(filepath.Join(tmpDir, name+".json"), blob.text, 0o644); err != nil {
			return fmt.Errorf("unable to write %s.json: %w", name, err)
		}
	}
	for _, blob := range unnamed {
		fmt.Printf("no reference file matches the definition at offset %d with keys %v\n", blob.offset, blob.keys)
	}
	if len(named) == 0 {
		return fmt.Errorf("no api/fields definition in %s matches a file of %s", binaryPath, referenceDir)
	}

	// A reference file that no blob matches is carried forward, so the generated code
	// keeps the type. The Setting split and the custom overrides below rewrite their own
	// files, so those are left out.
	custom, err := customNames()
	if err != nil {
		return err
	}
	for name := range reference {
		if _, ok := named[name]; ok || strings.HasPrefix(name, "Setting") || custom[name] {
			continue
		}
		fmt.Printf("%s.json: no definition in the native image, carried forward from %s\n", name, referenceDir)
		content, err := os.ReadFile(filepath.Join(referenceDir, name+".json"))
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(tmpDir, name+".json"), content, 0o644); err != nil {
			return err
		}
	}

	if err := splitSettings(tmpDir); err != nil {
		return err
	}
	if err := copyCustom(tmpDir); err != nil {
		return err
	}

	if err := os.RemoveAll(fieldsDir); err != nil {
		return err
	}
	return os.Rename(tmpDir, fieldsDir)
}

// fieldsBlob is one JSON object carved out of the native image.
type fieldsBlob struct {
	offset int
	text   []byte
	keys   []string
}

// carveFieldsBlobs returns every distinct pretty-printed JSON object in data, in
// offset order.
func carveFieldsBlobs(data []byte) []fieldsBlob {
	var blobs []fieldsBlob
	seen := map[string]bool{}
	for _, loc := range fieldsBlobStart.FindAllIndex(data, -1) {
		start := loc[0]
		end := min(start+fieldsBlobMaxLen, len(data))
		dec := json.NewDecoder(bytes.NewReader(data[start:end]))
		var object map[string]json.RawMessage
		if err := dec.Decode(&object); err != nil || len(object) == 0 {
			continue
		}
		text := data[start : start+int(dec.InputOffset())]
		if seen[string(text)] {
			continue
		}
		seen[string(text)] = true
		keys := make([]string, 0, len(object))
		for k := range object {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		blobs = append(blobs, fieldsBlob{offset: start, text: text, keys: keys})
	}
	return blobs
}

// referenceKeySets reads the top-level keys of every JSON object file in dir, by
// file name without the extension.
func referenceKeySets(dir string) (map[string][]string, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		return nil, err
	}
	reference := map[string][]string{}
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}
		var object map[string]json.RawMessage
		if err := json.Unmarshal(content, &object); err != nil || len(object) == 0 {
			continue
		}
		keys := make([]string, 0, len(object))
		for k := range object {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		name := filepath.Base(file)
		reference[name[:len(name)-len(".json")]] = keys
	}
	if len(reference) == 0 {
		return nil, fmt.Errorf("no JSON object file in reference directory %s", dir)
	}
	return reference, nil
}

// nameFieldsBlobs gives each blob the name of the reference file whose keys it
// shares the most, keeping the best blob per name, and returns the rest. A blob
// with fewer than fieldsNameMinKeys keys is dropped without a report.
func nameFieldsBlobs(blobs []fieldsBlob, reference map[string][]string) (map[string]fieldsBlob, []fieldsBlob) {
	named := map[string]fieldsBlob{}
	scores := map[string]float64{}
	var unnamed []fieldsBlob
	for _, blob := range blobs {
		if len(blob.keys) < fieldsNameMinKeys {
			continue
		}
		bestName, bestScore := "", 0.0
		for name, keys := range reference {
			if score := jaccard(blob.keys, keys); score > bestScore {
				bestName, bestScore = name, score
			}
		}
		if bestScore < fieldsNameMinScore {
			unnamed = append(unnamed, blob)
			continue
		}
		if prev, ok := scores[bestName]; !ok || bestScore > prev {
			named[bestName] = blob
			scores[bestName] = bestScore
		}
	}
	return named, unnamed
}

// customNames returns the names, without the extension, of the custom overrides.
func customNames() (map[string]bool, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	files, err := filepath.Glob(filepath.Join(findModuleRoot(wd), "cmd", "fields", "custom", "*.json"))
	if err != nil {
		return nil, err
	}
	names := map[string]bool{}
	for _, file := range files {
		name := filepath.Base(file)
		names[name[:len(name)-len(".json")]] = true
	}
	return names, nil
}

// jaccard is the size of the intersection over the size of the union of two sorted
// key lists.
func jaccard(a, b []string) float64 {
	inA := make(map[string]bool, len(a))
	for _, k := range a {
		inA[k] = true
	}
	both := 0
	for _, k := range b {
		if inA[k] {
			both++
		}
	}
	union := len(a) + len(b) - both
	if union == 0 {
		return 0
	}
	return float64(both) / float64(union)
}
