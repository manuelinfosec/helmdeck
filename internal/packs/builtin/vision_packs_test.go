package builtin

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/tosin2013/helmdeck/internal/gateway"
	"github.com/tosin2013/helmdeck/internal/packs"
	"github.com/tosin2013/helmdeck/internal/session"
)

// scriptedDispatcher hands out a queue of canned model replies and
// records what was sent. The vision packs iterate the action loop,
// so we need to return different responses on successive calls.
type scriptedDispatcher struct {
	calls   int
	replies []string
	captured []gateway.ChatRequest
}

func (s *scriptedDispatcher) Dispatch(_ context.Context, req gateway.ChatRequest) (gateway.ChatResponse, error) {
	s.captured = append(s.captured, req)
	idx := s.calls
	if idx >= len(s.replies) {
		idx = len(s.replies) - 1
	}
	s.calls++
	return gateway.ChatResponse{
		Choices: []gateway.Choice{{
			Index:   0,
			Message: gateway.Message{Role: "assistant", Content: gateway.TextContent(s.replies[idx])},
		}},
	}, nil
}

func newVisionPackEngine(t *testing.T, ex session.Executor) *packs.Engine {
	t.Helper()
	return packs.New(
		packs.WithRuntime(fakeRuntime{}),
		packs.WithSessionExecutor(ex),
	)
}

func TestVisionClickAnywhere_TwoStepDone(t *testing.T) {
	// Step 1: model says click(50, 60). Step 2: model says done.
	disp := &scriptedDispatcher{replies: []string{
		`{"action":"click","x":50,"y":60,"reason":"submit button"}`,
		`{"action":"done","reason":"clicked successfully"}`,
	}}
	// Executor needs to return PNG bytes for screenshot calls and
	// success for xdotool. We use the recordingExecutor pattern from
	// desktop_run_app_test.go but with a per-call script.
	ex := &recordingExecutor{
		replies: []session.ExecResult{
			{Stdout: []byte("\x89PNG-step1")}, // screenshot 1
			{},                                 // xdotool click
			{Stdout: []byte("\x89PNG-step2")}, // screenshot 2
		},
	}
	eng := newVisionPackEngine(t, ex)
	pack := VisionClickAnywhere(disp)

	res, err := eng.Execute(context.Background(), pack,
		json.RawMessage(`{"goal":"click submit","model":"ollama/llama3.2-vision","max_steps":4}`))
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	var out struct {
		Completed   bool                   `json:"completed"`
		Steps       int                    `json:"steps"`
		FinalAction map[string]interface{} `json:"final_action"`
	}
	if err := json.Unmarshal(res.Output, &out); err != nil {
		t.Fatal(err)
	}
	if !out.Completed {
		t.Errorf("expected completed=true, got %+v", out)
	}
	if out.Steps != 2 {
		t.Errorf("expected 2 steps, got %d", out.Steps)
	}
	if out.FinalAction["action"] != "done" {
		t.Errorf("final action wrong: %+v", out.FinalAction)
	}
	if disp.calls != 2 {
		t.Errorf("dispatcher should have been called twice, got %d", disp.calls)
	}
	// Multimodal payload assertion: every call should have one image_url part.
	for i, req := range disp.captured {
		if !req.Messages[1].Content.IsMultipart() {
			t.Errorf("call %d: user message not multipart", i)
		}
	}
}

func TestVisionClickAnywhere_RespectsMaxSteps(t *testing.T) {
	// Model never returns done — pack should bail at max_steps.
	disp := &scriptedDispatcher{replies: []string{
		`{"action":"click","x":1,"y":1,"reason":"keep trying"}`,
	}}
	ex := &recordingExecutor{
		replies: []session.ExecResult{
			{Stdout: []byte("png")}, {}, // step 1: screenshot + click
			{Stdout: []byte("png")}, {}, // step 2
			{Stdout: []byte("png")}, {}, // step 3
		},
	}
	eng := newVisionPackEngine(t, ex)
	pack := VisionClickAnywhere(disp)
	res, err := eng.Execute(context.Background(), pack,
		json.RawMessage(`{"goal":"x","model":"ollama/llama3.2-vision","max_steps":3}`))
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	var out struct {
		Completed bool `json:"completed"`
		Steps     int  `json:"steps"`
	}
	_ = json.Unmarshal(res.Output, &out)
	if out.Completed {
		t.Errorf("should not be completed when model never says done")
	}
	if out.Steps != 3 {
		t.Errorf("expected exactly max_steps=3 steps, got %d", out.Steps)
	}
}

func TestVisionExtractVisibleText_LiftsReason(t *testing.T) {
	disp := &scriptedDispatcher{replies: []string{
		`{"action":"done","reason":"Username:\nPassword:\nSubmit"}`,
	}}
	ex := &recordingExecutor{replies: []session.ExecResult{
		{Stdout: []byte("\x89PNG-screen")},
	}}
	eng := newVisionPackEngine(t, ex)
	res, err := eng.Execute(context.Background(), VisionExtractVisibleText(disp),
		json.RawMessage(`{"model":"ollama/llama3.2-vision"}`))
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	var out struct {
		Text  string `json:"text"`
		Model string `json:"model"`
	}
	_ = json.Unmarshal(res.Output, &out)
	if !strings.Contains(out.Text, "Username:") || !strings.Contains(out.Text, "Submit") {
		t.Errorf("text not extracted: %q", out.Text)
	}
	if out.Model != "ollama/llama3.2-vision" {
		t.Errorf("model not echoed: %q", out.Model)
	}
}

func TestVisionFillFormByLabel_FillsAllFields(t *testing.T) {
	// Two fields, sorted alphabetically: password then username.
	// Each field's loop returns done on the first iteration.
	disp := &scriptedDispatcher{replies: []string{
		`{"action":"done","reason":"password filled"}`,
		`{"action":"done","reason":"username filled"}`,
	}}
	ex := &recordingExecutor{
		replies: []session.ExecResult{
			{Stdout: []byte("png1")}, // screenshot for field 1
			{Stdout: []byte("png2")}, // screenshot for field 2
		},
	}
	eng := newVisionPackEngine(t, ex)
	res, err := eng.Execute(context.Background(), VisionFillFormByLabel(disp),
		json.RawMessage(`{"model":"ollama/llama3.2-vision","fields":{"username":"alice","password":"hunter2"},"max_steps":4}`))
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	var out struct {
		Completed    bool     `json:"completed"`
		FieldsFilled []string `json:"fields_filled"`
		Steps        int      `json:"steps"`
	}
	_ = json.Unmarshal(res.Output, &out)
	if !out.Completed {
		t.Errorf("expected all fields filled, got %+v", out)
	}
	if len(out.FieldsFilled) != 2 {
		t.Errorf("expected 2 fields filled, got %v", out.FieldsFilled)
	}
	// Sorted order: password before username.
	if out.FieldsFilled[0] != "password" || out.FieldsFilled[1] != "username" {
		t.Errorf("fields_filled order wrong: %v", out.FieldsFilled)
	}
}

func TestVisionFillFormByLabel_RequiresFields(t *testing.T) {
	disp := &scriptedDispatcher{replies: []string{`{"action":"done"}`}}
	eng := newVisionPackEngine(t, &recordingExecutor{})
	_, err := eng.Execute(context.Background(), VisionFillFormByLabel(disp),
		json.RawMessage(`{"model":"ollama/llama3.2-vision","fields":{}}`))
	if err == nil {
		t.Fatal("expected error for empty fields map")
	}
}

// --- Issue #102: action history + ctx cancellation + fill_form artifacts ---

// TestVisionClickAnywhere_AccumulatesHistory drives 3 iterations and
// asserts that the 3rd dispatcher call carries 2 prior attempts in
// its user-message text. This is the load-bearing change for #102:
// without history threading, weak models re-emit the same failed
// click on every turn.
func TestVisionClickAnywhere_AccumulatesHistory(t *testing.T) {
	disp := &scriptedDispatcher{replies: []string{
		`{"action":"click","x":376,"y":69,"reason":"click URL bar"}`,
		`{"action":"click","x":380,"y":75,"reason":"try slightly right"}`,
		`{"action":"done","reason":"focused"}`,
	}}
	ex := &recordingExecutor{
		replies: []session.ExecResult{
			{Stdout: []byte("\x89PNG-iter1")}, // screenshot 1
			{},                                 // xdotool click 1
			{Stdout: []byte("\x89PNG-iter2")}, // screenshot 2
			{},                                 // xdotool click 2
			{Stdout: []byte("\x89PNG-iter3")}, // screenshot 3
			// (no click on iter 3 — model emits done)
		},
	}
	eng := newVisionPackEngine(t, ex)
	res, err := eng.Execute(context.Background(), VisionClickAnywhere(disp),
		json.RawMessage(`{"goal":"focus the URL bar","model":"ollama/llama3.2-vision","max_steps":5}`))
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	var out struct {
		Completed bool `json:"completed"`
		Steps     int  `json:"steps"`
	}
	_ = json.Unmarshal(res.Output, &out)
	if !out.Completed || out.Steps != 3 {
		t.Errorf("expected completed=true steps=3, got %+v", out)
	}
	if disp.calls != 3 {
		t.Fatalf("expected 3 dispatcher calls, got %d", disp.calls)
	}

	// First iteration: no history.
	textOf := func(req gateway.ChatRequest) string {
		var s string
		for _, p := range req.Messages[1].Content.Parts() {
			if p.Type == gateway.ContentPartText {
				s += p.Text
			}
		}
		return s
	}
	if t1 := textOf(disp.captured[0]); strings.Contains(t1, "Prior attempts") {
		t.Errorf("first iteration should NOT have history, got: %q", t1)
	}
	// Second iteration: 1 prior attempt.
	t2 := textOf(disp.captured[1])
	if !strings.Contains(t2, "Prior attempts") || !strings.Contains(t2, "click(376, 69)") {
		t.Errorf("second iteration missing 1st-attempt history, got: %q", t2)
	}
	// Third iteration: 2 prior attempts.
	t3 := textOf(disp.captured[2])
	if !strings.Contains(t3, "click(376, 69)") || !strings.Contains(t3, "click(380, 75)") {
		t.Errorf("third iteration missing accumulated history, got: %q", t3)
	}
}

// TestVisionClickAnywhere_HonorsContextCancellation verifies that a
// cancelled context exits the pack-level loop instead of running to
// max_steps. Without the ctx.Err() check, the loop keeps spinning.
func TestVisionClickAnywhere_HonorsContextCancellation(t *testing.T) {
	disp := &scriptedDispatcher{replies: []string{
		`{"action":"click","x":50,"y":60,"reason":"first"}`,
		`{"action":"click","x":50,"y":60,"reason":"second"}`,
		`{"action":"click","x":50,"y":60,"reason":"third"}`,
	}}
	ex := &recordingExecutor{
		replies: []session.ExecResult{
			{Stdout: []byte("png1")}, {}, // screenshot + click
			{Stdout: []byte("png2")}, {}, // screenshot + click
		},
	}
	eng := newVisionPackEngine(t, ex)

	ctx, cancel := context.WithCancel(context.Background())
	// Cancel immediately so the loop's ctx.Err() check fires on
	// iteration 1. We don't need timing — we just need the check to
	// short-circuit before max_steps is reached.
	cancel()

	pack := VisionClickAnywhere(disp)
	_, err := eng.Execute(ctx, pack,
		json.RawMessage(`{"goal":"click","model":"ollama/llama3.2-vision","max_steps":50}`))
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
	if !strings.Contains(err.Error(), "canceled") && !strings.Contains(err.Error(), "context") {
		t.Errorf("expected context-cancellation error, got: %v", err)
	}
}

// TestVisionFillFormByLabel_RecordsArtifacts verifies the parity fix:
// fill_form now records per-step screenshots like click_anywhere does.
// Without this, form-fill workflows leave no audit trail.
func TestVisionFillFormByLabel_RecordsArtifacts(t *testing.T) {
	disp := &scriptedDispatcher{replies: []string{
		`{"action":"done","reason":"name filled"}`,
	}}
	ex := &recordingExecutor{
		replies: []session.ExecResult{
			{Stdout: []byte("\x89PNG-form-shot")},
		},
	}
	eng := newVisionPackEngine(t, ex)
	res, err := eng.Execute(context.Background(), VisionFillFormByLabel(disp),
		json.RawMessage(`{"model":"ollama/llama3.2-vision","fields":{"name":"alice"},"max_steps":2}`))
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if len(res.Artifacts) == 0 {
		t.Fatal("expected at least one artifact recorded — fill_form parity with click_anywhere")
	}
	first := res.Artifacts[0]
	if first.Pack != "vision.fill_form_by_label" {
		t.Errorf("artifact pack = %q, want vision.fill_form_by_label", first.Pack)
	}
	if !strings.HasPrefix(first.Key, "vision.fill_form_by_label/") || !strings.Contains(first.Key, "step-") || !strings.HasSuffix(first.Key, ".png") {
		t.Errorf("unexpected artifact key: %q", first.Key)
	}
}

func TestVisionPacksRequireDesktopMode(t *testing.T) {
	for _, tc := range []struct {
		name string
		pack *packs.Pack
	}{
		{"click_anywhere", VisionClickAnywhere(nil)},
		{"extract_visible_text", VisionExtractVisibleText(nil)},
		{"fill_form_by_label", VisionFillFormByLabel(nil)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.pack.SessionSpec.Env["HELMDECK_MODE"] != "desktop" {
				t.Errorf("pack should request desktop mode, got %v", tc.pack.SessionSpec.Env)
			}
			if !tc.pack.NeedsSession {
				t.Errorf("pack should require a session")
			}
		})
	}
}
