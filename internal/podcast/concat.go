// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 The helmdeck contributors

package podcast

import (
	"context"
	"fmt"
	"strings"

	"github.com/tosin2013/helmdeck/internal/session"
)

const (
	defaultSilenceMs    = 600
	concatBytesCap      = 256 << 20 // 256 MiB final-podcast cap (~3.5 hours of 128 kbps mp3)
	concatTempDir       = "/tmp/helmdeck-podcast"
)

// ConcatOptions controls how Concat stitches per-turn MP3 segments.
type ConcatOptions struct {
	// SilenceBetweenTurnsMs is the gap inserted between turns to
	// give the listener a beat. Default is defaultSilenceMs (600ms).
	SilenceBetweenTurnsMs int
}

// Concat takes a slice of per-turn MP3 byte buffers and produces a
// single MP3 with silence padding between turns. The work runs
// inside a session sidecar (ffmpeg + ffprobe must be installed
// there — they are in the helmdeck-sidecar image).
//
// Returns the final MP3 bytes plus the total duration in seconds
// (computed from per-turn ffprobe + the silence count).
//
// The session-side temp dir is /tmp/helmdeck-podcast — cleaned up at
// the end of the call (best-effort).
func Concat(ctx context.Context, ex session.Executor, sessionID string, turns [][]byte, opts ConcatOptions) ([]byte, float64, error) {
	if len(turns) == 0 {
		return nil, 0, fmt.Errorf("podcast/concat: no turns to stitch")
	}
	silenceMs := opts.SilenceBetweenTurnsMs
	if silenceMs <= 0 {
		silenceMs = defaultSilenceMs
	}

	// 1. Make the temp dir.
	if _, err := ex.Exec(ctx, sessionID, session.ExecRequest{
		Cmd: []string{"sh", "-c", "rm -rf " + concatTempDir + " && mkdir -p " + concatTempDir},
	}); err != nil {
		return nil, 0, fmt.Errorf("mkdir tempdir: %w", err)
	}

	// 2. Write each turn's MP3 bytes via base64+decode (Exec.Stdin
	// would let us stream, but `tee` from stdin is awkward across
	// the session-executor abstraction; base64 inline is small
	// enough for our 32 MiB-per-turn cap and avoids the multi-step
	// dance).
	for i, mp3 := range turns {
		if err := writeTurnFile(ctx, ex, sessionID, i, mp3); err != nil {
			return nil, 0, fmt.Errorf("write turn %d: %w", i, err)
		}
	}

	// 3. Make a silence segment used between every pair of turns.
	// Once, reused.
	silenceSec := float64(silenceMs) / 1000.0
	silencePath := concatTempDir + "/silence.mp3"
	if _, err := ex.Exec(ctx, sessionID, session.ExecRequest{
		Cmd: []string{"sh", "-c", fmt.Sprintf(
			"ffmpeg -y -f lavfi -i anullsrc=r=44100:cl=mono -t %.3f -acodec libmp3lame %s",
			silenceSec, silencePath,
		)},
	}); err != nil {
		return nil, 0, fmt.Errorf("silence segment: %w", err)
	}

	// 4. Build a concat-demuxer list: turn-0, silence, turn-1, silence, ..., turn-N
	var listB strings.Builder
	for i := range turns {
		fmt.Fprintf(&listB, "file '%s/turn-%03d.mp3'\n", concatTempDir, i)
		if i < len(turns)-1 {
			fmt.Fprintf(&listB, "file '%s'\n", silencePath)
		}
	}
	listPath := concatTempDir + "/concat.txt"
	if _, err := ex.Exec(ctx, sessionID, session.ExecRequest{
		Cmd:   []string{"sh", "-c", "cat > " + listPath},
		Stdin: []byte(listB.String()),
	}); err != nil {
		return nil, 0, fmt.Errorf("write concat list: %w", err)
	}

	// 5. Concat. Re-encode (-acodec libmp3lame) rather than -c copy
	// because the silence segment may have a slightly different
	// frame-size from ElevenLabs' MP3s, and concat-demuxer with
	// -c copy is finicky about that.
	finalPath := concatTempDir + "/final.mp3"
	if _, err := ex.Exec(ctx, sessionID, session.ExecRequest{
		Cmd: []string{"sh", "-c", fmt.Sprintf(
			"ffmpeg -y -f concat -safe 0 -i %s -acodec libmp3lame -b:a 128k %s",
			listPath, finalPath,
		)},
	}); err != nil {
		return nil, 0, fmt.Errorf("ffmpeg concat: %w", err)
	}

	// 6. ffprobe the duration of the result.
	probeRes, err := ex.Exec(ctx, sessionID, session.ExecRequest{
		Cmd: []string{"sh", "-c",
			"ffprobe -v error -show_entries format=duration -of csv=p=0 " + finalPath},
	})
	if err != nil {
		return nil, 0, fmt.Errorf("ffprobe duration: %w", err)
	}
	durStr := strings.TrimSpace(string(probeRes.Stdout))
	var duration float64
	fmt.Sscanf(durStr, "%f", &duration)

	// 7. Read the final MP3 bytes back. Cap at 256 MiB.
	readRes, err := ex.Exec(ctx, sessionID, session.ExecRequest{
		Cmd: []string{"sh", "-c",
			fmt.Sprintf("dd if=%s bs=1M count=256 2>/dev/null", finalPath)},
	})
	if err != nil {
		return nil, 0, fmt.Errorf("read final mp3: %w", err)
	}
	if len(readRes.Stdout) == 0 {
		return nil, 0, fmt.Errorf("final mp3 is empty")
	}
	if len(readRes.Stdout) > concatBytesCap {
		return nil, 0, fmt.Errorf("final mp3 %d bytes exceeds cap", len(readRes.Stdout))
	}

	// 8. Cleanup, best-effort.
	_, _ = ex.Exec(ctx, sessionID, session.ExecRequest{
		Cmd: []string{"sh", "-c", "rm -rf " + concatTempDir},
	})

	return readRes.Stdout, duration, nil
}

// writeTurnFile writes the per-turn MP3 bytes into the session temp
// dir at /tmp/helmdeck-podcast/turn-NNN.mp3. Uses Stdin to stream
// without base64-encoding overhead.
func writeTurnFile(ctx context.Context, ex session.Executor, sessionID string, idx int, mp3 []byte) error {
	path := fmt.Sprintf("%s/turn-%03d.mp3", concatTempDir, idx)
	res, err := ex.Exec(ctx, sessionID, session.ExecRequest{
		Cmd:   []string{"sh", "-c", "cat > " + path},
		Stdin: mp3,
	})
	if err != nil {
		return err
	}
	if res.ExitCode != 0 {
		return fmt.Errorf("write turn %d exit %d: %s", idx, res.ExitCode, strings.TrimSpace(string(res.Stderr)))
	}
	return nil
}

// SilenceTurn produces a silent MP3 segment of the given duration —
// used as the fallback when no API key is configured. Same command
// shape as Concat's between-turn silence; lets the pack handler
// reuse the same Concat path with all-silence "synthesis".
func SilenceTurn(ctx context.Context, ex session.Executor, sessionID string, seconds float64) ([]byte, error) {
	silencePath := concatTempDir + "/silent-turn.mp3"
	if _, err := ex.Exec(ctx, sessionID, session.ExecRequest{
		Cmd: []string{"sh", "-c", "mkdir -p " + concatTempDir},
	}); err != nil {
		return nil, fmt.Errorf("mkdir: %w", err)
	}
	if _, err := ex.Exec(ctx, sessionID, session.ExecRequest{
		Cmd: []string{"sh", "-c", fmt.Sprintf(
			"ffmpeg -y -f lavfi -i anullsrc=r=44100:cl=mono -t %.3f -acodec libmp3lame %s",
			seconds, silencePath,
		)},
	}); err != nil {
		return nil, fmt.Errorf("silence ffmpeg: %w", err)
	}
	res, err := ex.Exec(ctx, sessionID, session.ExecRequest{
		Cmd: []string{"sh", "-c", "cat " + silencePath},
	})
	if err != nil {
		return nil, fmt.Errorf("read silent turn: %w", err)
	}
	return res.Stdout, nil
}
