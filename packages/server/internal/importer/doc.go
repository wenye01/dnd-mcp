// Package importer provides map import functionality from various VTT formats.
//
// The importer module supports:
//   - Universal VTT (.uvtt) format
//   - Foundry VTT Scene (.json) format
//   - Foundry VTT Module (LevelDB compendium) format
//   - Auto-detection of format from data
//
// Architecture:
//   - Format detection via FormatDetector
//   - Format-specific Parsers for raw data parsing
//   - Converters to transform parsed data into Map models
//   - Validators to ensure converted maps meet requirements
//   - ImportService as the main orchestrator
//
// Extensibility:
//   New formats can be added by implementing:
//   - Parser interface for parsing raw data
//   - Converter interface for converting to Map model
//   - Registering with ImportService
//
// 规则参考: N/A - This is an import utility module
package importer
