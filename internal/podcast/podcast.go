// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 The helmdeck contributors

// Package podcast is the shared core for helmdeck's podcast.generate
// pack. It defines the Engine interface that TTS backends implement
// (ElevenLabs day 1; PlayHT, Hume.ai, Resemble.ai, etc. as future
// PRs slot in by adding a new file in this package), plus the
// ffmpeg-concat helper that stitches per-turn MP3 segments into the
// final podcast audio.
//
// Why a separate package: the pack handler (internal/packs/builtin/
// podcast_generate.go) shouldn't grow N engine implementations
// inline. Each engine is a separate file under internal/podcast/
// implementing Engine; the pack handler just picks one by name and
// iterates turns.
package podcast

import (
	"context"
	"fmt"
)

// Turn is one speaker line in a podcast script. Multi-speaker
// dialogue is just an ordered slice of Turns; solo monologue is a
// slice with one speaker name throughout.
type Turn struct {
	Speaker string `json:"speaker"`
	Text    string `json:"text"`
}

// SynthesizeOptions carries the per-call knobs an engine needs. Not
// every engine honors every field — VoiceID and ModelID are required
// for ElevenLabs; Stability/SimilarityBoost are ElevenLabs-specific
// but harmless for engines that ignore them.
type SynthesizeOptions struct {
	VoiceID         string
	ModelID         string
	Stability       float64
	SimilarityBoost float64
}

// Engine synthesizes one Turn into MP3 bytes. The pack handler
// iterates turns, calls Synthesize per turn, and concats the
// returned bytes into the final podcast audio via Concat.
//
// Implementations live alongside this file:
//   - ElevenLabsEngine    (elevenlabs.go) — ships day 1
//   - future: PlayHTEngine, HumeEngine, ResembleEngine ...
type Engine interface {
	// Name returns the canonical engine identifier as it appears in
	// the pack input (e.g. "elevenlabs", "playht"). Used by
	// PickEngine to route a name string to an implementation.
	Name() string

	// Synthesize converts one Turn to MP3 bytes. The voice/model are
	// in opts so the same engine can be invoked with different voices
	// for each turn (multi-speaker dialogue).
	Synthesize(ctx context.Context, turn Turn, opts SynthesizeOptions) ([]byte, error)
}

// ErrEngineNotFound is returned by PickEngine when the requested
// engine name doesn't match any registered implementation.
var ErrEngineNotFound = fmt.Errorf("podcast: engine not found")

// PickEngine routes an engine name to an implementation. The
// elevenlabs engine is constructed with the supplied apiKey (empty
// string is allowed — silent-fallback behavior is the engine's
// responsibility, not the picker's).
//
// Future engines extend the switch as they're added. We deliberately
// don't use a global registry pattern: the call sites (pack handler,
// tests) all know which credentials they have, and a switch keeps
// the day-1 surface small.
func PickEngine(name, apiKey string) (Engine, error) {
	switch name {
	case "", "elevenlabs":
		return &ElevenLabsEngine{APIKey: apiKey}, nil
	default:
		return nil, fmt.Errorf("%w: %q (day 1 only ships %q)", ErrEngineNotFound, name, "elevenlabs")
	}
}
