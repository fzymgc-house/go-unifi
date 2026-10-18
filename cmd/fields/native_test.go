package main

import (
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	nativeUserJSON    = "{\n    \"mac\": \"^([0-9A-Fa-f]{2}:){5}([0-9A-Fa-f]{2})$\",\n    \"name\": \"\",\n    \"note\": \"\"\n}"
	nativeSettingJSON = "{\n  \"mgmt\": {\n    \"x_ssh_enabled\": \"true|false\"\n  },\n  \"lte\": {\n    \"enabled\": \"true|false\"\n  }\n}"
	nativeStrayJSON   = "{\n    \"EVT_AD_BackupCreated\": \"\",\n    \"EVT_AD_BackupFailed\": \"\"\n}"
)

// buildNativeImage mimics the Network 11 binary: the field definitions sit between
// unrelated bytes, as pretty-printed JSON, with their names stored elsewhere.
func buildNativeImage(t *testing.T, blobs ...string) string {
	t.Helper()

	noise := make([]byte, 4096)
	_, _ = rand.New(rand.NewSource(1)).Read(noise)

	var image []byte
	image = append(image, noise...)
	image = append(image, "api/fields/User.json\x00api/fields/Setting.json\x00"...)
	for _, blob := range blobs {
		image = append(image, noise[:256]...)
		image = append(image, blob...)
	}
	image = append(image, noise...)

	path := filepath.Join(t.TempDir(), "unifi")
	require.NoError(t, os.WriteFile(path, image, 0o755))
	return path
}

// referenceFieldsDir mimics the fields directory of the version before, which
// names each carved blob through its top-level keys.
func referenceFieldsDir(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "User.json"), []byte(testUserJSON), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "Setting.json"), []byte(testSettingJSON), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "SettingMgmt.json"), []byte(`{"x_ssh_enabled": "true|false"}`), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "WlanGroup.json"), []byte(`{"name": ".{1,128}"}`), 0o644))
	return dir
}

func TestExtractFromNativeImageNamesBlobsByReferenceKeys(t *testing.T) {
	t.Parallel()

	image := buildNativeImage(t, nativeUserJSON, nativeStrayJSON, nativeSettingJSON, "{\n    \"name\": \"\"\n}")
	fieldsDir := filepath.Join(t.TempDir(), "v11.0.89")

	require.NoError(t, ExtractFromNativeImage(image, referenceFieldsDir(t), fieldsDir))

	userJSON, err := os.ReadFile(filepath.Join(fieldsDir, "User.json"))
	require.NoError(t, err)
	assert.JSONEq(t, nativeUserJSON, string(userJSON), "the carved definition replaces the reference one")

	lteJSON, err := os.ReadFile(filepath.Join(fieldsDir, "SettingLte.json"))
	require.NoError(t, err, "a new section of Setting.json is split out like the others")
	assert.JSONEq(t, `{"enabled": "true|false"}`, string(lteJSON))

	_, err = os.Stat(filepath.Join(fieldsDir, "FirewallPolicy.json"))
	require.NoError(t, err, "the custom overrides are copied in")

	groupJSON, err := os.ReadFile(filepath.Join(fieldsDir, "WlanGroup.json"))
	require.NoError(t, err, "a reference file that no blob matches is carried forward")
	assert.JSONEq(t, `{"name": ".{1,128}"}`, string(groupJSON), "a one-key blob never takes a reference file's name")

	entries, err := os.ReadDir(fieldsDir)
	require.NoError(t, err)
	for _, entry := range entries {
		assert.NotContains(t, entry.Name(), "EVT_AD", "a blob that matches no reference file is left out")
	}
	_, err = os.Stat(fieldsDir + ".tmp")
	assert.True(t, os.IsNotExist(err), "the temporary directory is removed")
}

func TestExtractFromNativeImageWithoutSettingErrors(t *testing.T) {
	t.Parallel()

	image := buildNativeImage(t, nativeUserJSON)
	fieldsDir := filepath.Join(t.TempDir(), "v11.0.89")

	err := ExtractFromNativeImage(image, referenceFieldsDir(t), fieldsDir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "settings")
	_, statErr := os.Stat(fieldsDir)
	assert.True(t, os.IsNotExist(statErr), "a failed run leaves no fields directory")
}

func TestExtractFromNativeImageSkipsWhenCached(t *testing.T) {
	t.Parallel()

	fieldsDir := filepath.Join(t.TempDir(), "v11.0.89")
	require.NoError(t, os.MkdirAll(fieldsDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(fieldsDir, "Setting.json"), []byte(testSettingJSON), 0o644))

	require.NoError(t, ExtractFromNativeImage(filepath.Join(t.TempDir(), "missing"), t.TempDir(), fieldsDir))
}
