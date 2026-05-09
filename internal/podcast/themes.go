// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 The helmdeck contributors

package podcast

import "fmt"

// Theme is the closed set of podcast styles the script generator
// supports. Each theme maps to a fragment in the LLM system prompt
// that bakes in podcast best practices (hooks, pacing, calls to
// action, structural patterns).
//
// Adding a new theme: append a constant here, add a case to
// ThemeFragment, and add a row to the docs page. Keep the set small
// — closed-set themes keep prompt-mode output predictable across
// models.
type Theme string

const (
	ThemeInterview    Theme = "interview"
	ThemeDebate       Theme = "debate"
	ThemeNewsRoundup  Theme = "news-roundup"
	ThemeDeepDive     Theme = "deep-dive"
	ThemeSoloEssay    Theme = "solo-essay"

	DefaultTheme = ThemeDeepDive
)

// ValidTheme reports whether a given theme is in the closed set.
func ValidTheme(t string) bool {
	switch Theme(t) {
	case ThemeInterview, ThemeDebate, ThemeNewsRoundup, ThemeDeepDive, ThemeSoloEssay:
		return true
	}
	return false
}

// ThemeFragment returns the system-prompt fragment baked into the
// LLM call when generating a script in mode B (prompt) or mode C
// (source_url/source_text). Each fragment is ~150 words — small
// enough to keep prompt-mode token cost low.
func ThemeFragment(t Theme) string {
	switch t {
	case ThemeInterview:
		return `STYLE: Interview podcast.
- Two speakers: a HOST and a GUEST. The host's name should be the first key in the speakers map; the guest is the second.
- Open with a one-sentence hook from the host that previews the conversation's payoff.
- The host asks open-ended questions. The guest does ~70% of the talking.
- Avoid yes/no questions; favor "tell me about", "walk me through", "what was surprising".
- Close with one actionable takeaway the listener can apply this week.
- Pacing: hook → topic intro → 3-5 deepening question rounds → takeaway → outro.
- Avoid filler ("um", "like", "you know"). Avoid host monologues longer than 2 sentences.
- Total length should land near the user's duration_target_min.`

	case ThemeDebate:
		return `STYLE: Debate podcast.
- Two speakers staking opposing positions on a single thesis.
- Speaker A presents their case in 60-90 words.
- Speaker B responds with their case in 60-90 words.
- Three rounds of point/counterpoint follow, each 30-60 words per turn.
- STEEL-MAN the opposing view: each speaker must restate the other's strongest argument before disagreeing. No strawmen.
- Close with a moderator-style summary by Speaker A acknowledging both views and naming what changed (or didn't) for them.
- Tone: civil, intellectually serious, but allow personality.
- Total length should land near the user's duration_target_min.`

	case ThemeNewsRoundup:
		return `STYLE: News roundup.
- One or two speakers, fast-paced.
- 3-5 short stories, each 60-120 seconds spoken.
- Each story has a one-line hook, the news, the so-what, and a transition.
- Halfway through, insert a sponsor-break placeholder turn from the host: "We'll be right back after a word from our sponsor." (No actual sponsor copy.)
- Close with a "what we're watching this week" segment teasing one upcoming thing.
- Tone: brisk, slightly playful, never breathless.
- Total length should land near the user's duration_target_min.`

	case ThemeDeepDive:
		return `STYLE: Deep-dive podcast.
- One or two speakers, single topic.
- Narrative arc: PROBLEM (why this matters) → EXPLORATION (the substance) → RESOLUTION (what we now know) → IMPLICATION (what to do with it).
- Use concrete examples and named numbers. Avoid vague claims.
- Pacing: write for the ear — short sentences, one idea per breath, occasional rhetorical pauses.
- If two speakers, alternate naturally: one introduces a sub-topic, the other reflects or challenges, then they continue.
- Close with one open question the listener can sit with.
- Total length should land near the user's duration_target_min.`

	case ThemeSoloEssay:
		return `STYLE: Solo essay.
- ONE speaker only. The first key in the speakers map IS the speaker.
- Monologue style — written for the ear, not the page.
- Short sentences. One idea per breath. Strategic repetition for emphasis.
- Use rhetorical questions to invite the listener in: "What if I told you...?"
- Structure: hook → personal anecdote or example → thesis → 2-3 supporting threads → resolution.
- 8-12 minute sweet spot for retention; aim for the user's duration_target_min within that window.
- Tone: thoughtful, conversational, occasionally vulnerable.`

	default:
		return ""
	}
}

// SystemPromptForScript returns the full system prompt for script
// generation, theme + general podcast best-practices baked in. The
// LLM is asked to emit JSON: an array of {speaker, text} objects.
//
// The speaker names that the LLM is allowed to use are listed
// inline so the model knows which voice slots are available.
func SystemPromptForScript(theme Theme, speakerNames []string, durationMin int) string {
	if !ValidTheme(string(theme)) {
		theme = DefaultTheme
	}
	speakerList := "(none)"
	if len(speakerNames) > 0 {
		speakerList = ""
		for i, n := range speakerNames {
			if i > 0 {
				speakerList += ", "
			}
			speakerList += `"` + n + `"`
		}
	}
	wordTarget := durationMin * 150 // ElevenLabs ~150 wpm

	return fmt.Sprintf(`You are a podcast script writer. Output ONLY a JSON array of {"speaker": "...", "text": "..."} objects. No surrounding code fences, no preamble, no explanation. Each object's "speaker" field MUST be one of these exact names: %s. Each "text" is the spoken line for that speaker.

%s

Target total length: ~%d words across all speakers (≈%d minutes at typical TTS pace). Don't pad to hit the target; quality over length. Avoid stage directions, sound effects, or non-spoken content — text-to-speech engines will read everything you write out loud.`,
		speakerList, ThemeFragment(theme), wordTarget, durationMin)
}
