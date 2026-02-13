// errors.go — Sentinel errors for Stricture.
//
// These are the expected failure modes that callers check with errors.Is().
// Every sentinel error is a specific, documented condition — not a catch-all.
package model

import "errors"

// Config errors.
var (
	// ErrConfigNotFound is returned when no .stricture.yml exists in the project.
	ErrConfigNotFound = errors.New("config not found")

	// ErrConfigInvalid is returned when .stricture.yml has invalid YAML or schema.
	ErrConfigInvalid = errors.New("config invalid")

	// ErrUnknownRule is returned when config references a rule ID that doesn't exist.
	ErrUnknownRule = errors.New("unknown rule")
)

// Parsing errors.
var (
	// ErrUnsupportedLanguage is returned when a file's extension has no registered adapter.
	ErrUnsupportedLanguage = errors.New("unsupported language")

	// ErrParseFailure is returned when a source file cannot be parsed (syntax error).
	ErrParseFailure = errors.New("parse failure")
)

// Manifest errors.
var (
	// ErrManifestNotFound is returned when no .stricture-manifest.yml exists.
	ErrManifestNotFound = errors.New("manifest not found")

	// ErrManifestInvalid is returned when the manifest fails schema validation.
	ErrManifestInvalid = errors.New("manifest invalid")

	// ErrContractNotFound is returned when a referenced contract ID doesn't exist.
	ErrContractNotFound = errors.New("contract not found")
)

// Cache errors.
var (
	// ErrCacheCorrupt is returned when a cache entry is unreadable or has a version mismatch.
	ErrCacheCorrupt = errors.New("cache corrupt")
)

// Runtime errors.
var (
	// ErrRulePanic is returned when a rule's Check() method panics.
	// The engine recovers from the panic and wraps it in this error.
	ErrRulePanic = errors.New("rule panic")
)
