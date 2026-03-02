// Package parser_test provides unit tests for parsers
package parser_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dnd-mcp/server/internal/importer/format"
	"github.com/dnd-mcp/server/internal/importer/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNDJSONParser_Open(t *testing.T) {
	t.Run("valid module with module.json", func(t *testing.T) {
		tempDir := t.TempDir()
		createTestModule(t, tempDir)

		p := parser.NewNDJSONParser()
		err := p.Open(tempDir)

		assert.NoError(t, err)
		assert.NoError(t, p.Close())
	})

	t.Run("non-existent module path", func(t *testing.T) {
		p := parser.NewNDJSONParser()
		err := p.Open("../../testdata/modules/does_not_exist")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "does not exist")
	})

	t.Run("module without module.json", func(t *testing.T) {
		tempDir := t.TempDir()
		// Create directory but no module.json
		require.NoError(t, os.MkdirAll(filepath.Join(tempDir, "packs"), 0755))

		p := parser.NewNDJSONParser()
		err := p.Open(tempDir)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "module.json")
	})

	t.Run("module without scene pack", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create module.json without Scene pack
		moduleJSON := `{
			"id": "test-module",
			"title": "Test Module",
			"description": "Test Description",
			"version": "1.0.0",
			"authors": [{"name": "Test Author"}],
			"compatibility": {"minimum": "10", "verified": "12"},
			"packs": [
				{
					"name": "actors",
					"label": "Test Actors",
					"path": "packs/actors.db",
					"type": "Actor"
				}
			]
		}`

		require.NoError(t, os.WriteFile(filepath.Join(tempDir, "module.json"), []byte(moduleJSON), 0644))

		p := parser.NewNDJSONParser()
		err := p.Open(tempDir)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no Scene pack")
	})
}

func TestNDJSONParser_ListScenes(t *testing.T) {
	// Create a temporary test module
	tempDir := t.TempDir()
	createTestModule(t, tempDir)

	p := parser.NewNDJSONParser()
	require.NoError(t, p.Open(tempDir))
	defer p.Close()

	scenes, err := p.ListScenes()
	require.NoError(t, err)

	// We created 2 scenes in the test data (1 is temp entity)
	assert.Len(t, scenes, 2)

	// Check that scene names are present
	assert.Contains(t, scenes, "Test Scene 1")
	assert.Contains(t, scenes, "Test Scene 2")
}

func TestNDJSONParser_GetScene(t *testing.T) {
	tempDir := t.TempDir()
	createTestModule(t, tempDir)

	p := parser.NewNDJSONParser()
	require.NoError(t, p.Open(tempDir))
	defer p.Close()

	t.Run("get scene by name", func(t *testing.T) {
		result, err := p.GetScene("Test Scene 1")
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, format.FormatFVTTScene, result.Format)

		data, ok := result.Data.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "Test Scene 1", data["name"])
		assert.Equal(t, "scene1-id", data["_id"])
	})

	t.Run("get scene by ID", func(t *testing.T) {
		result, err := p.GetScene("scene2-id")
		require.NoError(t, err)
		assert.NotNil(t, result)

		data, ok := result.Data.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "Test Scene 2", data["name"])
	})

	t.Run("scene not found", func(t *testing.T) {
		result, err := p.GetScene("Non-existent Scene")
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("get temp entity (should be skipped in ListScenes but accessible by ID)", func(t *testing.T) {
		result, err := p.GetScene("temp-id")
		require.NoError(t, err)
		assert.NotNil(t, result)
	})
}

func TestNDJSONParser_GetModuleInfo(t *testing.T) {
	tempDir := t.TempDir()
	createTestModule(t, tempDir)

	p := parser.NewNDJSONParser()
	require.NoError(t, p.Open(tempDir))
	defer p.Close()

	info, err := p.GetModuleInfo()
	require.NoError(t, err)

	assert.Equal(t, "test-module", info.Name)
	assert.Equal(t, "Test Module", info.Title)
	assert.Equal(t, "Test Description", info.Description)
	assert.Equal(t, "1.0.0", info.Version)
	assert.Equal(t, "Test Author", info.Author)
	assert.Equal(t, "dnd5e", info.System)
	assert.Equal(t, 2, info.SceneCount) // 2 non-temp scenes
}

func TestNDJSONParser_GetAllScenes(t *testing.T) {
	tempDir := t.TempDir()
	createTestModule(t, tempDir)

	p := parser.NewNDJSONParser()
	require.NoError(t, p.Open(tempDir))
	defer p.Close()

	scenes, err := p.GetAllScenes()
	require.NoError(t, err)

	// Should return 2 scenes (temp entity is filtered out)
	assert.Len(t, scenes, 2)

	// Verify scene data
	for _, scene := range scenes {
		assert.Equal(t, format.FormatFVTTScene, scene.Format)
		data, ok := scene.Data.(map[string]interface{})
		require.True(t, ok)
		name := data["name"].(string)
		assert.NotContains(t, name, "#[CF_")
	}
}

func TestNDJSONParser_Close(t *testing.T) {
	tempDir := t.TempDir()
	createTestModule(t, tempDir)

	p := parser.NewNDJSONParser()
	require.NoError(t, p.Open(tempDir))

	// Close should release resources
	assert.NoError(t, p.Close())

	// After close, operations should fail
	_, err := p.ListScenes()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not opened")
}

func TestNDJSONParser_OpenWithoutClose(t *testing.T) {
	tempDir := t.TempDir()
	createTestModule(t, tempDir)

	p := parser.NewNDJSONParser()
	require.NoError(t, p.Open(tempDir))

	// Operations should work
	scenes, err := p.ListScenes()
	require.NoError(t, err)
	assert.NotEmpty(t, scenes)
}

// createTestModule creates a minimal test module structure
func createTestModule(t *testing.T, dir string) {
	t.Helper()

	// Create module.json
	moduleJSON := `{
		"id": "test-module",
		"title": "Test Module",
		"description": "Test Description",
		"version": "1.0.0",
		"authors": [{"name": "Test Author"}],
		"compatibility": {"minimum": "10", "verified": "12"},
		"packs": [
			{
				"name": "scenes",
				"label": "Test Scenes",
				"path": "packs/scenes.db",
				"type": "Scene"
			}
		]
	}`

	// Create packs directory
	packsDir := filepath.Join(dir, "packs")
	require.NoError(t, os.MkdirAll(packsDir, 0755))

	// Create module.json
	modulePath := filepath.Join(dir, "module.json")
	require.NoError(t, os.WriteFile(modulePath, []byte(moduleJSON), 0644))

	// Create scenes.db with test data
	scenesDB := `{"_id":"scene1-id","name":"Test Scene 1","active":false,"navigation":true,"width":2000,"height":1500,"grid":100,"gridDistance":5,"gridUnits":"ft","tokenVision":true,"fogExploration":true,"globalLight":false,"darkness":0,"drawings":[],"tokens":[],"lights":[],"notes":[],"sounds":[],"templates":[],"tiles":[],"walls":[]}
{"_id":"scene2-id","name":"Test Scene 2","active":false,"navigation":true,"width":3000,"height":2000,"grid":100,"gridDistance":5,"gridUnits":"ft","tokenVision":true,"fogExploration":true,"globalLight":false,"darkness":0,"drawings":[],"tokens":[],"lights":[],"notes":[],"sounds":[],"templates":[],"tiles":[],"walls":[]}
{"_id":"temp-id","name":"#[CF_tempEntity]","active":false,"navigation":false,"width":1000,"height":1000,"grid":100,"gridDistance":5,"gridUnits":"ft","tokenVision":true,"fogExploration":true,"globalLight":false,"darkness":0,"drawings":[],"tokens":[],"lights":[],"notes":[],"sounds":[],"templates":[],"tiles":[],"walls":[]}
`

	scenesPath := filepath.Join(packsDir, "scenes.db")
	require.NoError(t, os.WriteFile(scenesPath, []byte(scenesDB), 0644))
}

// TestNDJSONParser_MalformedJSON tests handling of malformed JSON
func TestNDJSONParser_MalformedJSON(t *testing.T) {
	tempDir := t.TempDir()

	// Create module.json
	packsDir := filepath.Join(tempDir, "packs")
	require.NoError(t, os.MkdirAll(packsDir, 0755))

	moduleJSON := `{
		"id": "test-module",
		"title": "Test Module",
		"description": "Test Description",
		"version": "1.0.0",
		"authors": [{"name": "Test Author"}],
		"compatibility": {"minimum": "10", "verified": "12"},
		"packs": [
			{
				"name": "scenes",
				"label": "Test Scenes",
				"path": "packs/scenes.db",
				"type": "Scene"
			}
		]
	}`

	modulePath := filepath.Join(tempDir, "module.json")
	require.NoError(t, os.WriteFile(modulePath, []byte(moduleJSON), 0644))

	// Create scenes.db with some valid and some invalid JSON
	scenesDB := `{"_id":"scene1-id","name":"Valid Scene","active":false,"navigation":true,"width":2000,"height":1500}
this is not valid json
{"_id":"scene2-id","name":"Another Valid Scene","active":false,"navigation":true,"width":2000,"height":1500}
`

	scenesPath := filepath.Join(packsDir, "scenes.db")
	require.NoError(t, os.WriteFile(scenesPath, []byte(scenesDB), 0644))

	p := parser.NewNDJSONParser()
	require.NoError(t, p.Open(tempDir))
	defer p.Close()

	// Should still be able to list valid scenes
	scenes, err := p.ListScenes()
	require.NoError(t, err)
	assert.Len(t, scenes, 2)
}

// TestNDJSONParser_EmptyScenesDB tests handling of empty scenes.db
func TestNDJSONParser_EmptyScenesDB(t *testing.T) {
	tempDir := t.TempDir()

	// Create module.json
	packsDir := filepath.Join(tempDir, "packs")
	require.NoError(t, os.MkdirAll(packsDir, 0744))

	moduleJSON := `{
		"id": "test-module",
		"title": "Test Module",
		"description": "Test Description",
		"version": "1.0.0",
		"authors": [{"name": "Test Author"}],
		"compatibility": {"minimum": "10", "verified": "12"},
		"packs": [
			{
				"name": "scenes",
				"label": "Test Scenes",
				"path": "packs/scenes.db",
				"type": "Scene"
			}
		]
	}`

	modulePath := filepath.Join(tempDir, "module.json")
	require.NoError(t, os.WriteFile(modulePath, []byte(moduleJSON), 0644))

	// Create empty scenes.db
	scenesPath := filepath.Join(packsDir, "scenes.db")
	require.NoError(t, os.WriteFile(scenesPath, []byte(""), 0644))

	p := parser.NewNDJSONParser()
	require.NoError(t, p.Open(tempDir))
	defer p.Close()

	scenes, err := p.ListScenes()
	require.NoError(t, err)
	assert.Empty(t, scenes)

	info, err := p.GetModuleInfo()
	require.NoError(t, err)
	assert.Equal(t, 0, info.SceneCount)
}
