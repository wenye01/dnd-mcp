// Package import_test provides integration tests for the import module
package import_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dnd-mcp/server/internal/importer/format"
	"github.com/dnd-mcp/server/internal/importer/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNDJSONParser_RealModule tests parsing a real FVTT module
// This test requires the actual baileywiki-maps module to be present
func TestNDJSONParser_RealModule(t *testing.T) {
	modulePath := "../../../../../docs/data/baileywiki-maps"

	// Skip if module is not available
	if _, err := os.Stat(modulePath); os.IsNotExist(err) {
		t.Skip("baileywiki-maps module not available, skipping real module test")
	}

	p := parser.NewNDJSONParser()
	require.NoError(t, p.Open(modulePath))
	defer p.Close()

	// Test module info
	info, err := p.GetModuleInfo()
	require.NoError(t, err)

	assert.Equal(t, "baileywiki-maps", info.Name)
	assert.Equal(t, "Baileywiki Free Maps Pack", info.Title)
	assert.Contains(t, info.Description, "free maps")
	assert.Equal(t, "Baileywiki", info.Author)

	t.Logf("Module info: %+v", info)

	// Test list scenes
	scenes, err := p.ListScenes()
	require.NoError(t, err)

	t.Logf("Found %d scenes", len(scenes))
	assert.Greater(t, len(scenes), 0, "Should have at least some scenes")

	// Print first few scene names
	for i, scene := range scenes {
		if i >= 5 {
			break
		}
		t.Logf("Scene %d: %s", i+1, scene)
	}

	// Test get first scene
	if len(scenes) > 0 {
		firstScene := scenes[0]
		result, err := p.GetScene(firstScene)
		require.NoError(t, err)

		assert.NotNil(t, result)
		assert.Equal(t, format.FormatFVTTScene, result.Format)

		data, ok := result.Data.(map[string]interface{})
		require.True(t, ok)

		// Verify scene has required fields
		assert.Contains(t, data, "_id")
		assert.Contains(t, data, "name")
		assert.Contains(t, data, "width")
		assert.Contains(t, data, "height")
		assert.Contains(t, data, "grid")

		t.Logf("First scene data: name=%s, width=%v, height=%v, grid=%v",
			data["name"], data["width"], data["height"], data["grid"])
	}
}

// TestNDJSONParser_SampleModule tests the sample module in testdata
func TestNDJSONParser_SampleModule(t *testing.T) {
	// Use the sample module created by test-data preparation
	modulePath := "../../testdata/modules/sample_ndjson_module"

	// Skip if sample module doesn't exist yet
	if _, err := os.Stat(modulePath); os.IsNotExist(err) {
		t.Skip("sample NDJSON module not available, run test data preparation first")
	}

	p := parser.NewNDJSONParser()
	require.NoError(t, p.Open(modulePath))
	defer p.Close()

	// Test module info
	info, err := p.GetModuleInfo()
	require.NoError(t, err)

	assert.Equal(t, "sample-ndjson-module", info.Name)
	assert.Equal(t, "Sample NDJSON Module", info.Title)

	// Test list scenes
	scenes, err := p.ListScenes()
	require.NoError(t, err)

	t.Logf("Sample module has %d scenes", len(scenes))

	// Verify we can get each scene
	for _, sceneName := range scenes {
		result, err := p.GetScene(sceneName)
		require.NoErrorf(t, err, "Failed to get scene: %s", sceneName)

		assert.NotNil(t, result)
		assert.Equal(t, format.FormatFVTTScene, result.Format)

		data, ok := result.Data.(map[string]interface{})
		require.Truef(t, ok, "Scene data is not a map: %s", sceneName)

		t.Logf("Scene: %s, ID: %v", sceneName, data["_id"])
	}
}

// TestNDJSONParser_AllBaileywikiScenes tests parsing all scenes from baileywiki-maps
func TestNDJSONParser_AllBaileywikiScenes(t *testing.T) {
	modulePath := "../../../../../docs/data/baileywiki-maps"

	// Skip if module is not available
	if _, err := os.Stat(modulePath); os.IsNotExist(err) {
		t.Skip("baileywiki-maps module not available, skipping")
	}

	p := parser.NewNDJSONParser()
	require.NoError(t, p.Open(modulePath))
	defer p.Close()

	scenes, err := p.GetAllScenes()
	require.NoError(t, err)

	t.Logf("Total scenes: %d", len(scenes))

	// Verify all scenes have required data
	for i, scene := range scenes {
		require.NotNil(t, scene.Data)
		assert.Equal(t, format.FormatFVTTScene, scene.Format)

		data, ok := scene.Data.(map[string]interface{})
		require.True(t, ok, "Scene %d: data is not a map", i)

		// Check required fields
		assert.Contains(t, data, "_id", "Scene %d: missing _id", i)
		assert.Contains(t, data, "name", "Scene %d: missing name", i)
		assert.Contains(t, data, "width", "Scene %d: missing width", i)
		assert.Contains(t, data, "height", "Scene %d: missing height", i)

		// Log some details
		if i < 3 {
			t.Logf("Scene %d: %s (%s)", i+1, data["name"], data["_id"])
		}
	}
}

// TestNDJSONParser_MultipleOpensCloses tests multiple open/close cycles
func TestNDJSONParser_MultipleOpensCloses(t *testing.T) {
	modulePath := "../../../../../docs/data/baileywiki-maps"

	// Skip if module is not available
	if _, err := os.Stat(modulePath); os.IsNotExist(err) {
		t.Skip("baileywiki-maps module not available, skipping")
	}

	p := parser.NewNDJSONParser()

	// First open
	require.NoError(t, p.Open(modulePath))
	scenes1, err := p.ListScenes()
	require.NoError(t, err)
	count1 := len(scenes1)

	// Close
	require.NoError(t, p.Close())

	// Second open
	require.NoError(t, p.Open(modulePath))
	scenes2, err := p.ListScenes()
	require.NoError(t, err)
	count2 := len(scenes2)

	// Should get same results
	assert.Equal(t, count1, count2)

	require.NoError(t, p.Close())
}

// TestNDJSONParser_ConcurrentAccess tests concurrent access to the parser
func TestNDJSONParser_ConcurrentAccess(t *testing.T) {
	modulePath := "../../../../../docs/data/baileywiki-maps"

	// Skip if module is not available
	if _, err := os.Stat(modulePath); os.IsNotExist(err) {
		t.Skip("baileywiki-maps module not available, skipping")
	}

	p := parser.NewNDJSONParser()
	require.NoError(t, p.Open(modulePath))
	defer p.Close()

	// Get scenes list first
	scenes, err := p.ListScenes()
	require.NoError(t, err)

	if len(scenes) < 2 {
		t.Skip("Need at least 2 scenes for concurrent test")
	}

	// Test concurrent scene access
	done := make(chan bool, 2)

	go func() {
		_, err := p.GetScene(scenes[0])
		assert.NoError(t, err)
		done <- true
	}()

	go func() {
		_, err := p.GetScene(scenes[1])
		assert.NoError(t, err)
		done <- true
	}()

	// Wait for both goroutines
	<-done
	<-done
}

// TestNDJSONParser_ModuleStructure verifies the module structure expectations
func TestNDJSONParser_ModuleStructure(t *testing.T) {
	modulePath := "../../../../../docs/data/baileywiki-maps"

	// Skip if module is not available
	if _, err := os.Stat(modulePath); os.IsNotExist(err) {
		t.Skip("baileywiki-maps module not available, skipping")
	}

	// Verify module.json exists
	moduleJSONPath := filepath.Join(modulePath, "module.json")
	info, err := os.Stat(moduleJSONPath)
	require.NoError(t, err, "module.json should exist")
	assert.False(t, info.IsDir(), "module.json should be a file")

	// Verify packs directory exists
	packsDir := filepath.Join(modulePath, "packs")
	info, err = os.Stat(packsDir)
	require.NoError(t, err, "packs directory should exist")
	assert.True(t, info.IsDir(), "packs should be a directory")

	// List files in packs
	entries, err := os.ReadDir(packsDir)
	require.NoError(t, err)

	var dbFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".db" {
			dbFiles = append(dbFiles, entry.Name())
		}
	}

	t.Logf("Found .db files in packs: %v", dbFiles)
	assert.NotEmpty(t, dbFiles, "Should have at least one .db file")

	// Verify the scenes.db file exists
	scenesDBPath := filepath.Join(packsDir, "baileywiki-maps.db")
	info, err = os.Stat(scenesDBPath)
	require.NoError(t, err, "baileywiki-maps.db should exist")
	assert.Greater(t, info.Size(), int64(0), "baileywiki-maps.db should not be empty")
}
