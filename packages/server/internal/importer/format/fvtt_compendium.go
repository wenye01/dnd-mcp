// Package format defines types for FVTT Compendium format
package format

// FVTTFolder represents a folder structure
type FVTTFolder struct {
	ID       string `json:"_id"`
	Name     string `json:"name"`
	Parent   string `json:"parent"`
	Type     string `json:"type"`
	Sorting  string `json:"sorting"`
	Color    string `json:"color"`
}

// FVTTSceneFlags contains commonly used flags from FVTT scenes
type FVTTSceneFlags struct {
	// CF (Compendium Folders) related flags
	CF *FVTTFlagCF `json:"cf,omitempty"`
}

// FVTTFlagCF contains Compendium Folders specific flags
type FVTTFlagCF struct {
	ID         string      `json:"id"`
	Path       string      `json:"path"`
	Color      string      `json:"color"`
	Name       string      `json:"name"`
	Children   []interface{} `json:"children"`
	FolderPath []string    `json:"folderPath"`
	FontColor  string      `json:"fontColor"`
	Icon       string      `json:"icon"`
	Sorting    string      `json:"sorting"`
	Contents   []string    `json:"contents"`
	Version    string      `json:"version"`
}

// FVTTModuleManifest represents the module.json file
type FVTTModuleManifest struct {
	ID           string             `json:"id"`
	Title        string             `json:"title"`
	Description  string             `json:"description"`
	Version      string             `json:"version"`
	Authors      []FVTTModuleAuthor `json:"authors"`
	Systems      []string           `json:"systems"`
	Packs        []FVTTPack         `json:"packs"`
	Dependencies []FVTTDependency   `json:"dependencies"`
	ESModules    []string           `json:"esmodules"`
	Styles       []string           `json:"styles"`
	Flags        map[string]interface{} `json:"flags"`
}

// FVTTModuleAuthor represents a module author
type FVTTModuleAuthor struct {
	Name    string `json:"name"`
	Email   string `json:"email,omitempty"`
	URL     string `json:"url,omitempty"`
	Discord string `json:"discord,omitempty"`
}

// FVTTPack represents a compendium pack
type FVTTPack struct {
	Name    string `json:"name"`
	Label   string `json:"label"`
	Path    string `json:"path"`
	Entity  string `json:"entity"` // "Scene", "Actor", etc.
	Type    string `json:"type"`   // "Actor", "Item", "JournalEntry", etc.
	System  string `json:"system,omitempty"`
	Private bool   `json:"private,omitempty"`
}

// FVTTDependency represents a module dependency
type FVTTDependency struct {
	Name    string `json:"name"`
	Type    string `json:"type,omitempty"`
	Manifest string `json:"manifest,omitempty"`
}
