package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// Deterministic seeded random
func TestSeededRandomDeterminism(t *testing.T) {
	// mulberry32 PRNG
	mulberry32 := func(seed uint32) func() float64 {
		return func() float64 {
			seed |= 0
			seed = seed + 0x6D2B79F5
			t := uint64(seed) * uint64(seed)
			seed = uint32(t) ^ uint32(t>>32)
			return float64(seed) / 4294967296.0
		}
	}

	rng1 := mulberry32(42)
	rng2 := mulberry32(42)

	for i := 0; i < 100; i++ {
		a := rng1()
		b := rng2()
		if a != b {
			t.Fatalf("seeded random diverged at iteration %d: %f != %f", i, a, b)
		}
	}

	// Different seeds should produce different sequences
	rng3 := mulberry32(99)
	different := false
	rngA := mulberry32(42)
	for i := 0; i < 20; i++ {
		if rngA() != rng3() {
			different = true
			break
		}
	}
	if !different {
		t.Fatal("different seeds produced identical sequences")
	}
}

// Test sketch schema validation
func TestSketchSchemaValidation(t *testing.T) {
	validSketch := LiraSketch{
		Schema:          "lira.sketch.v1",
		ID:              "test-sketch",
		Name:            "test",
		DurationSeconds: 3.0,
		Mood:            "happy",
		Material:        "glass",
		Shape:           "rise",
		Density:         5,
		Brightness:      7,
		Softness:        6,
		Movement:        4,
		Seed:            "abc123",
	}

	if validSketch.Schema != "lira.sketch.v1" {
		t.Error("schema must be lira.sketch.v1")
	}
	if validSketch.ID == "" {
		t.Error("id must not be empty")
	}
	if validSketch.Name == "" {
		t.Error("name must not be empty")
	}
	if validSketch.DurationSeconds <= 0 || validSketch.DurationSeconds > 10 {
		t.Error("duration must be between 0 and 10")
	}
	if validSketch.Density < 1 || validSketch.Density > 10 {
		t.Error("density must be 1-10")
	}
	if validSketch.Brightness < 1 || validSketch.Brightness > 10 {
		t.Error("brightness must be 1-10")
	}
	if validSketch.Softness < 1 || validSketch.Softness > 10 {
		t.Error("softness must be 1-10")
	}
	if validSketch.Movement < 1 || validSketch.Movement > 10 {
		t.Error("movement must be 1-10")
	}

	validMoods := map[string]bool{
		"calm": true, "happy": true, "curious": true, "tense": true,
		"magic": true, "mechanical": true, "aquatic": true, "warm": true,
	}
	if !validMoods[validSketch.Mood] {
		t.Errorf("invalid mood: %s", validSketch.Mood)
	}

	validMaterials := map[string]bool{
		"glass": true, "wood": true, "bell": true, "bubble": true,
		"pluck": true, "breath": true, "metal": true, "soft-noise": true, "toy": true,
	}
	if !validMaterials[validSketch.Material] {
		t.Errorf("invalid material: %s", validSketch.Material)
	}

	validShapes := map[string]bool{
		"rise": true, "fall": true, "bounce": true, "pulse": true,
		"wave": true, "sparkle": true, "swell": true, "pop": true, "orbit": true,
	}
	if !validShapes[validSketch.Shape] {
		t.Errorf("invalid shape: %s", validSketch.Shape)
	}
}

// Test variant generation metadata
func TestVariantMetadata(t *testing.T) {
	variant := LiraVariant{
		Schema:    "lira.variant.v1",
		ID:        "var-001",
		SketchID:  "sketch-001",
		Seed:      "seed-42",
		Index:     3,
		ExportIDs: []string{},
	}

	if variant.Schema != "lira.variant.v1" {
		t.Error("variant schema must be lira.variant.v1")
	}
	if variant.ID == "" {
		t.Error("variant id must not be empty")
	}
	if variant.SketchID == "" {
		t.Error("variant must reference a sketch")
	}
	if variant.Index < 0 {
		t.Error("variant index must be non-negative")
	}
}

// Test export metadata
func TestExportMetadata(t *testing.T) {
	exp := LiraExport{
		Schema:          "lira.export.v1",
		ID:              "exp-001",
		SketchID:        "sketch-001",
		VariantID:       "var-001",
		Filename:        "test.wav",
		Format:          "wav",
		DurationSeconds: 2.5,
		SampleRate:      44100,
	}

	if exp.Schema != "lira.export.v1" {
		t.Error("export schema must be lira.export.v1")
	}
	if exp.ID == "" {
		t.Error("export id must not be empty")
	}
	if exp.Filename == "" {
		t.Error("export must have a filename")
	}

	validFormats := map[string]bool{"wav": true, "json": true, "zip": true}
	if !validFormats[exp.Format] {
		t.Errorf("invalid export format: %s", exp.Format)
	}
}

// Test project file read/write
func TestProjectFileReadWrite(t *testing.T) {
	dir := t.TempDir()
	h := newHub(dir)

	state := newProjectState("testproj")
	state.Project.Name = "Test Project"
	state.Project.Author = "Test Author"

	if err := h.saveProject("testproj", state); err != nil {
		t.Fatalf("saveProject failed: %v", err)
	}

	// Verify file exists
	projFile := filepath.Join(dir, "testproj", "project.json")
	if _, err := os.Stat(projFile); os.IsNotExist(err) {
		t.Fatal("project.json was not created")
	}

	// Load and verify
	loaded, err := h.loadProject("testproj")
	if err != nil {
		t.Fatalf("loadProject failed: %v", err)
	}
	if loaded.Project.Name != "Test Project" {
		t.Errorf("expected 'Test Project', got '%s'", loaded.Project.Name)
	}
	if loaded.Project.Author != "Test Author" {
		t.Errorf("expected 'Test Author', got '%s'", loaded.Project.Author)
	}
}

// Test sketch CRUD through file persistence
func TestSketchFilePersistence(t *testing.T) {
	dir := t.TempDir()
	h := newHub(dir)

	state := newProjectState("testproj")
	sketch := LiraSketch{
		Schema:          "lira.sketch.v1",
		ID:              "sk-1",
		Name:            "tiny victory",
		DurationSeconds: 2.2,
		Mood:            "happy",
		Material:        "bell",
		Shape:           "rise",
		Density:         5,
		Brightness:      8,
		Softness:        5,
		Movement:        6,
		Seed:            "default",
	}
	state.Sketches[sketch.ID] = sketch

	if err := h.saveProject("testproj", state); err != nil {
		t.Fatalf("saveProject failed: %v", err)
	}

	loaded, err := h.loadProject("testproj")
	if err != nil {
		t.Fatalf("loadProject failed: %v", err)
	}

	loadedSketch, ok := loaded.Sketches["sk-1"]
	if !ok {
		t.Fatal("sketch not found after reload")
	}
	if loadedSketch.Name != "tiny victory" {
		t.Errorf("expected 'tiny victory', got '%s'", loadedSketch.Name)
	}
	if loadedSketch.Mood != "happy" {
		t.Errorf("expected mood 'happy', got '%s'", loadedSketch.Mood)
	}
}

// Test API handlers
func TestVersionHandler(t *testing.T) {
	srv := &server{
		hub:         newHub(t.TempDir()),
		allowOrigin: "*",
		appVersion:  "test",
		appSHA:      "abc123",
	}

	req := httptest.NewRequest(http.MethodGet, "/api/version", nil)
	w := httptest.NewRecorder()
	srv.handleVersion(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp["version"] != "test" {
		t.Errorf("expected version 'test', got '%s'", resp["version"])
	}
}

func TestHealthHandler(t *testing.T) {
	srv := &server{
		hub:         newHub(t.TempDir()),
		allowOrigin: "*",
		appVersion:  "test",
		appSHA:      "abc123",
	}

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	srv.handleHealth(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestProjectAPI(t *testing.T) {
	srv := &server{
		hub:         newHub(t.TempDir()),
		allowOrigin: "*",
	}

	// GET project (auto-creates)
	req := httptest.NewRequest(http.MethodGet, "/api/project?project=testproj", nil)
	w := httptest.NewRecorder()
	srv.handleProject(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var proj LiraProject
	if err := json.Unmarshal(w.Body.Bytes(), &proj); err != nil {
		t.Fatalf("failed to parse project: %v", err)
	}
	if proj.Schema != "lira.project.v1" {
		t.Errorf("expected schema 'lira.project.v1', got '%s'", proj.Schema)
	}

	// PUT project
	updateBody := `{"name":"Updated Project","author":"Test","sampleRate":48000}`
	req = httptest.NewRequest(http.MethodPut, "/api/project?project=testproj", bytes.NewReader([]byte(updateBody)))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	srv.handleProject(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	if err := json.Unmarshal(w.Body.Bytes(), &proj); err != nil {
		t.Fatalf("failed to parse project: %v", err)
	}
	if proj.Name != "Updated Project" {
		t.Errorf("expected 'Updated Project', got '%s'", proj.Name)
	}
}

func TestSketchesAPI(t *testing.T) {
	srv := &server{
		hub:         newHub(t.TempDir()),
		allowOrigin: "*",
	}

	// POST sketch
	body := `{
		"id": "sk-1",
		"name": "tiny victory",
		"durationSeconds": 2.2,
		"mood": "happy",
		"material": "bell",
		"shape": "rise",
		"density": 5,
		"brightness": 8,
		"softness": 5,
		"movement": 6,
		"seed": "default"
	}`
	req := httptest.NewRequest(http.MethodPost, "/api/sketches?project=testproj", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.handleSketches(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	// GET sketches list
	req = httptest.NewRequest(http.MethodGet, "/api/sketches?project=testproj", nil)
	w = httptest.NewRecorder()
	srv.handleSketches(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var sketches []LiraSketch
	if err := json.Unmarshal(w.Body.Bytes(), &sketches); err != nil {
		t.Fatalf("failed to parse sketches: %v", err)
	}
	if len(sketches) != 1 {
		t.Fatalf("expected 1 sketch, got %d", len(sketches))
	}
	if sketches[0].Name != "tiny victory" {
		t.Errorf("expected 'tiny victory', got '%s'", sketches[0].Name)
	}

	// GET single sketch
	req = httptest.NewRequest(http.MethodGet, "/api/sketches/sk-1?project=testproj", nil)
	w = httptest.NewRecorder()
	srv.routes().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// DELETE sketch
	req = httptest.NewRequest(http.MethodDelete, "/api/sketches/sk-1?project=testproj", nil)
	w = httptest.NewRecorder()
	srv.routes().ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", w.Code, w.Body.String())
	}
}

func TestVariantsAPI(t *testing.T) {
	srv := &server{
		hub:         newHub(t.TempDir()),
		allowOrigin: "*",
	}

	// First create a sketch
	sketchBody := `{"id":"sk-1","name":"test","durationSeconds":3,"mood":"calm","material":"glass","shape":"rise","density":5,"brightness":5,"softness":5,"movement":5,"seed":"x"}`
	req := httptest.NewRequest(http.MethodPost, "/api/sketches?project=testproj", bytes.NewReader([]byte(sketchBody)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.handleSketches(w, req)

	// POST variant
	varBody := `{"id":"var-1","seed":"seed-1","index":0}`
	req = httptest.NewRequest(http.MethodPost, "/api/sketches/sk-1/variants?project=testproj", bytes.NewReader([]byte(varBody)))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	srv.routes().ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var variant LiraVariant
	if err := json.Unmarshal(w.Body.Bytes(), &variant); err != nil {
		t.Fatalf("failed to parse variant: %v", err)
	}
	if variant.SketchID != "sk-1" {
		t.Errorf("expected sketchId 'sk-1', got '%s'", variant.SketchID)
	}

	// GET variants
	req = httptest.NewRequest(http.MethodGet, "/api/sketches/sk-1/variants?project=testproj", nil)
	w = httptest.NewRecorder()
	srv.routes().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var variants []LiraVariant
	if err := json.Unmarshal(w.Body.Bytes(), &variants); err != nil {
		t.Fatalf("failed to parse variants: %v", err)
	}
	if len(variants) != 1 {
		t.Fatalf("expected 1 variant, got %d", len(variants))
	}
}

func TestExportsAPI(t *testing.T) {
	srv := &server{
		hub:         newHub(t.TempDir()),
		allowOrigin: "*",
	}

	// POST export
	body := `{
		"sketchId": "sk-1",
		"variantId": "var-1",
		"filename": "test.wav",
		"format": "wav",
		"durationSeconds": 2.5,
		"sampleRate": 44100
	}`
	req := httptest.NewRequest(http.MethodPost, "/api/exports?project=testproj", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.handleExports(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var exp LiraExport
	if err := json.Unmarshal(w.Body.Bytes(), &exp); err != nil {
		t.Fatalf("failed to parse export: %v", err)
	}
	if exp.Format != "wav" {
		t.Errorf("expected format 'wav', got '%s'", exp.Format)
	}

	// GET exports
	req = httptest.NewRequest(http.MethodGet, "/api/exports?project=testproj", nil)
	w = httptest.NewRecorder()
	srv.handleExports(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var exports []LiraExport
	if err := json.Unmarshal(w.Body.Bytes(), &exports); err != nil {
		t.Fatalf("failed to parse exports: %v", err)
	}
	if len(exports) != 1 {
		t.Fatalf("expected 1 export, got %d", len(exports))
	}
}

func TestSanitizeProjectID(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"my-project", "my-project"},
		{"default", "default"},
		{"../etc/passwd", "default"},
		{"hello world", "default"},
		{"", "default"},
		{"valid_name-123", "valid_name-123"},
	}
	for _, tt := range tests {
		result := sanitizeProjectID(tt.input)
		if result != tt.expected {
			t.Errorf("sanitizeProjectID(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestNewProjectState(t *testing.T) {
	state := newProjectState("myproj")
	if state.Project.Name != "myproj" {
		t.Errorf("expected project name 'myproj', got '%s'", state.Project.Name)
	}
	if state.Project.Schema != "lira.project.v1" {
		t.Errorf("expected schema 'lira.project.v1', got '%s'", state.Project.Schema)
	}
	if state.Project.SampleRate != 44100 {
		t.Errorf("expected sample rate 44100, got %d", state.Project.SampleRate)
	}
	if len(state.Sketches) != 0 {
		t.Errorf("expected 0 sketches, got %d", len(state.Sketches))
	}
}

func TestCORSHeaders(t *testing.T) {
	srv := &server{
		hub:         newHub(t.TempDir()),
		allowOrigin: "*",
	}

	req := httptest.NewRequest(http.MethodOptions, "/api/version", nil)
	w := httptest.NewRecorder()
	srv.handleVersion(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204 for OPTIONS, got %d", w.Code)
	}
	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("expected CORS header")
	}
}
