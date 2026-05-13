package main

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	pathpkg "path"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"sort"
	"strings"
	"sync"
	"time"
)

var version = "dev"

//go:embed all:web/dist
var embeddedAssets embed.FS

var embeddedUI fs.FS

func init() {
	sub, err := fs.Sub(embeddedAssets, "web/dist")
	if err == nil {
		embeddedUI = sub
	}
}

// --- Lira data types ---

type LiraMood string

type LiraMaterial string

type LiraShape string

type LiraExportFormat string

const (
	ExportFormatWAV  LiraExportFormat = "wav"
	ExportFormatJSON LiraExportFormat = "json"
	ExportFormatZIP  LiraExportFormat = "zip"
)

type LiraProject struct {
	Schema          string  `json:"schema"`
	Name            string  `json:"name"`
	Author          string  `json:"author"`
	Description     string  `json:"description"`
	SampleRate      int     `json:"sampleRate"`
	ExportFormat    string  `json:"exportFormat"`
	DefaultDuration float64 `json:"defaultDuration"`
	OutputDir       string  `json:"outputDir"`
	NamingPattern   string  `json:"namingPattern"`
	Version         int64   `json:"version"`
	UpdatedAt       string  `json:"updatedAt"`
}

type LiraSketch struct {
	Schema          string  `json:"schema"`
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	DurationSeconds float64 `json:"durationSeconds"`
	Mood            string  `json:"mood"`
	Material        string  `json:"material"`
	Shape           string  `json:"shape"`
	Density         int     `json:"density"`
	Brightness      int     `json:"brightness"`
	Softness        int     `json:"softness"`
	Movement        int     `json:"movement"`
	Seed            string  `json:"seed"`
	CreatedAt       string  `json:"createdAt"`
	UpdatedAt       string  `json:"updatedAt"`
}

type LiraVariant struct {
	Schema    string   `json:"schema"`
	ID        string   `json:"id"`
	SketchID  string   `json:"sketchId"`
	Seed      string   `json:"seed"`
	Index     int      `json:"index"`
	CreatedAt string   `json:"createdAt"`
	ExportIDs []string `json:"exportIds"`
}

type LiraExport struct {
	Schema          string  `json:"schema"`
	ID              string  `json:"id"`
	SketchID        string  `json:"sketchId"`
	VariantID       string  `json:"variantId,omitempty"`
	Filename        string  `json:"filename"`
	Format          string  `json:"format"`
	CreatedAt       string  `json:"createdAt"`
	DurationSeconds float64 `json:"durationSeconds"`
	SampleRate      int     `json:"sampleRate"`
}

type projectState struct {
	Project  LiraProject
	Sketches map[string]LiraSketch
	Variants map[string]map[string]LiraVariant // sketchID -> variantID -> variant
	Exports  map[string]LiraExport
}

func newProjectState(projectID string) *projectState {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	return &projectState{
		Project: LiraProject{
			Schema:          "lira.project.v1",
			Name:            projectID,
			SampleRate:      44100,
			ExportFormat:    "wav",
			DefaultDuration: 3,
			OutputDir:       "exports",
			NamingPattern:   "{project}-{sketch}-{variant}-{seed}",
			Version:         0,
			UpdatedAt:       now,
		},
		Sketches: make(map[string]LiraSketch),
		Variants: make(map[string]map[string]LiraVariant),
		Exports:  make(map[string]LiraExport),
	}
}

// --- Hub ---

type hub struct {
	mu       sync.RWMutex
	projects map[string]*projectState
	dataDir  string
}

var projectIDPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

func sanitizeProjectID(raw string) string {
	if projectIDPattern.MatchString(raw) {
		return raw
	}
	return "default"
}

func newHub(dataDir string) *hub {
	return &hub{
		projects: make(map[string]*projectState),
		dataDir:  dataDir,
	}
}

func (h *hub) projectDir(projectID string) string {
	return filepath.Join(h.dataDir, projectID)
}

func (h *hub) projectFile(projectID string) string {
	return filepath.Join(h.projectDir(projectID), "project.json")
}

func (h *hub) sketchesDir(projectID string) string {
	return filepath.Join(h.projectDir(projectID), "sketches")
}

func (h *hub) sketchFile(projectID, sketchID string) string {
	return filepath.Join(h.sketchesDir(projectID), sketchID+".json")
}

func (h *hub) variantsDir(projectID, sketchID string) string {
	return filepath.Join(h.projectDir(projectID), "variants", sketchID)
}

func (h *hub) variantFile(projectID, sketchID, variantID string) string {
	return filepath.Join(h.variantsDir(projectID, sketchID), variantID+".json")
}

func (h *hub) exportsDir(projectID, sketchID string) string {
	return filepath.Join(h.projectDir(projectID), "exports", sketchID)
}

func (h *hub) exportFile(projectID, sketchID, exportID string) string {
	return filepath.Join(h.exportsDir(projectID, sketchID), exportID+".json")
}

func writeJSONFile(path string, v any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	bytes, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, bytes, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func readJSONFile(path string, v any) error {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, v)
}

func (h *hub) loadProject(projectID string) (*projectState, error) {
	state := newProjectState(projectID)

	// Load project metadata
	if err := readJSONFile(h.projectFile(projectID), &state.Project); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	if state.Project.Schema == "" {
		state.Project.Schema = "lira.project.v1"
	}

	// Load sketches
	entries, err := os.ReadDir(h.sketchesDir(projectID))
	if err == nil {
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
				continue
			}
			var sketch LiraSketch
			if err := readJSONFile(filepath.Join(h.sketchesDir(projectID), entry.Name()), &sketch); err != nil {
				continue
			}
			if sketch.ID != "" {
				state.Sketches[sketch.ID] = sketch
			}
		}
	}

	// Load variants
	for sketchID := range state.Sketches {
		varEntries, err := os.ReadDir(h.variantsDir(projectID, sketchID))
		if err != nil {
			continue
		}
		state.Variants[sketchID] = make(map[string]LiraVariant)
		for _, entry := range varEntries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
				continue
			}
			var variant LiraVariant
			if err := readJSONFile(filepath.Join(h.variantsDir(projectID, sketchID), entry.Name()), &variant); err != nil {
				continue
			}
			if variant.ID != "" {
				state.Variants[sketchID][variant.ID] = variant
			}
		}
	}

	// Load exports
	for sketchID := range state.Sketches {
		expEntries, err := os.ReadDir(h.exportsDir(projectID, sketchID))
		if err != nil {
			continue
		}
		for _, entry := range expEntries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
				continue
			}
			var exp LiraExport
			if err := readJSONFile(filepath.Join(h.exportsDir(projectID, sketchID), entry.Name()), &exp); err != nil {
				continue
			}
			if exp.ID != "" {
				state.Exports[exp.ID] = exp
			}
		}
	}

	return state, nil
}

func (h *hub) saveProject(projectID string, state *projectState) error {
	// Save project metadata
	if err := writeJSONFile(h.projectFile(projectID), state.Project); err != nil {
		return err
	}

	// Save sketches
	for _, sketch := range state.Sketches {
		if err := writeJSONFile(h.sketchFile(projectID, sketch.ID), sketch); err != nil {
			return err
		}
	}

	// Save variants
	for sketchID, variants := range state.Variants {
		for _, variant := range variants {
			if err := writeJSONFile(h.variantFile(projectID, sketchID, variant.ID), variant); err != nil {
				return err
			}
		}
	}

	// Save exports
	for _, exp := range state.Exports {
		sketchID := exp.SketchID
		if sketchID == "" {
			continue
		}
		if err := writeJSONFile(h.exportFile(projectID, sketchID, exp.ID), exp); err != nil {
			return err
		}
	}

	return nil
}

func (h *hub) getOrCreateState(projectID string) (*projectState, error) {
	h.mu.RLock()
	if state, ok := h.projects[projectID]; ok {
		h.mu.RUnlock()
		return state, nil
	}
	h.mu.RUnlock()

	h.mu.Lock()
	defer h.mu.Unlock()

	if state, ok := h.projects[projectID]; ok {
		return state, nil
	}

	state, err := h.loadProject(projectID)
	if err != nil {
		state = newProjectState(projectID)
	}
	h.projects[projectID] = state
	return state, nil
}

// --- Server ---

type server struct {
	hub         *hub
	allowOrigin string
	uiFS        fs.FS
	appVersion  string
	appSHA      string
}

func resolveGitSHA() string {
	if sha := resolveGitSHAFromBuildInfo(); sha != "" {
		return sha
	}
	if sha := resolveGitSHAFromGit(); sha != "" {
		return sha
	}
	return "unknown"
}

func resolveGitSHAFromBuildInfo() string {
	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}
	sha := ""
	dirty := false
	for _, setting := range buildInfo.Settings {
		switch setting.Key {
		case "vcs.revision":
			sha = strings.TrimSpace(setting.Value)
		case "vcs.modified":
			dirty = setting.Value == "true"
		}
	}
	if sha == "" {
		return ""
	}
	if len(sha) > 12 {
		sha = sha[:12]
	}
	if dirty {
		sha += "-dirty"
	}
	return sha
}

func resolveGitSHAFromGit() string {
	shaBytes, err := exec.Command("git", "rev-parse", "--short=12", "HEAD").Output()
	if err != nil {
		return ""
	}
	sha := strings.TrimSpace(string(shaBytes))
	if sha == "" {
		return ""
	}
	statusBytes, err := exec.Command("git", "status", "--porcelain").Output()
	if err == nil && strings.TrimSpace(string(statusBytes)) != "" {
		sha += "-dirty"
	}
	return sha
}

func (s *server) writeCORS(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	if s.allowOrigin == "*" {
		w.Header().Set("Access-Control-Allow-Origin", "*")
	} else if origin == s.allowOrigin {
		w.Header().Set("Access-Control-Allow-Origin", origin)
	}
	w.Header().Set("Vary", "Origin")
	w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Cache-Control", "no-store")
}

func (s *server) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.writeCORS(w, r)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"version": s.appVersion,
		"sha":     s.appSHA,
	})
}

func (s *server) handleVersion(w http.ResponseWriter, r *http.Request) {
	s.writeCORS(w, r)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"version": s.appVersion,
		"sha":     s.appSHA,
	})
}

func getProjectParam(r *http.Request) string {
	id := r.URL.Query().Get("project")
	if id == "" {
		return "default"
	}
	return sanitizeProjectID(id)
}

func extractSketchID(path string) string {
	// path is like /api/sketches/{id} or /api/sketches/{id}/variants or /api/sketches/{id}/variants/{vid}
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) >= 3 && parts[0] == "api" && parts[1] == "sketches" {
		return parts[2]
	}
	return ""
}

func extractVariantID(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) >= 5 && parts[0] == "api" && parts[1] == "sketches" && parts[3] == "variants" {
		return parts[4]
	}
	return ""
}

// --- Project handlers ---

func (s *server) handleProject(w http.ResponseWriter, r *http.Request) {
	s.writeCORS(w, r)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	projectID := getProjectParam(r)

	switch r.Method {
	case http.MethodGet:
		state, err := s.hub.getOrCreateState(projectID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(state.Project)

	case http.MethodPut:
		var proj LiraProject
		if err := json.NewDecoder(r.Body).Decode(&proj); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		state, err := s.hub.getOrCreateState(projectID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		s.hub.mu.Lock()
		proj.Schema = "lira.project.v1"
		proj.Version = state.Project.Version + 1
		proj.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
		state.Project = proj
		s.hub.mu.Unlock()
		if err := s.hub.saveProject(projectID, state); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(state.Project)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// --- Sketches handlers ---

func (s *server) handleSketches(w http.ResponseWriter, r *http.Request) {
	s.writeCORS(w, r)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	projectID := getProjectParam(r)

	switch r.Method {
	case http.MethodGet:
		state, err := s.hub.getOrCreateState(projectID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		sketches := make([]LiraSketch, 0, len(state.Sketches))
		for _, s := range state.Sketches {
			sketches = append(sketches, s)
		}
		sort.Slice(sketches, func(i, j int) bool {
			return sketches[i].CreatedAt < sketches[j].CreatedAt
		})
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(sketches)

	case http.MethodPost:
		var sketch LiraSketch
		if err := json.NewDecoder(r.Body).Decode(&sketch); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		if sketch.ID == "" {
			sketch.ID = fmt.Sprintf("sketch-%d", time.Now().UnixNano())
		}
		sketch.Schema = "lira.sketch.v1"
		now := time.Now().UTC().Format(time.RFC3339Nano)
		if sketch.CreatedAt == "" {
			sketch.CreatedAt = now
		}
		sketch.UpdatedAt = now

		state, err := s.hub.getOrCreateState(projectID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		s.hub.mu.Lock()
		state.Sketches[sketch.ID] = sketch
		s.hub.mu.Unlock()
		if err := s.hub.saveProject(projectID, state); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(sketch)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *server) handleSketch(w http.ResponseWriter, r *http.Request) {
	s.writeCORS(w, r)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	projectID := getProjectParam(r)
	sketchID := extractSketchID(r.URL.Path)

	state, err := s.hub.getOrCreateState(projectID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	switch r.Method {
	case http.MethodGet:
		sketch, ok := state.Sketches[sketchID]
		if !ok {
			http.Error(w, "sketch not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(sketch)

	case http.MethodPut:
		var sketch LiraSketch
		if err := json.NewDecoder(r.Body).Decode(&sketch); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		sketch.ID = sketchID
		sketch.Schema = "lira.sketch.v1"
		sketch.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
		if sketch.CreatedAt == "" {
			if existing, ok := state.Sketches[sketchID]; ok {
				sketch.CreatedAt = existing.CreatedAt
			} else {
				sketch.CreatedAt = sketch.UpdatedAt
			}
		}
		s.hub.mu.Lock()
		state.Sketches[sketchID] = sketch
		s.hub.mu.Unlock()
		if err := s.hub.saveProject(projectID, state); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(sketch)

	case http.MethodDelete:
		if _, ok := state.Sketches[sketchID]; !ok {
			http.Error(w, "sketch not found", http.StatusNotFound)
			return
		}
		s.hub.mu.Lock()
		delete(state.Sketches, sketchID)
		delete(state.Variants, sketchID)
		// Clean up export files
		for id, exp := range state.Exports {
			if exp.SketchID == sketchID {
				delete(state.Exports, id)
			}
		}
		s.hub.mu.Unlock()
		if err := s.hub.saveProject(projectID, state); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// --- Variants handlers ---

func (s *server) handleVariants(w http.ResponseWriter, r *http.Request) {
	s.writeCORS(w, r)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	projectID := getProjectParam(r)
	sketchID := extractSketchID(r.URL.Path)

	state, err := s.hub.getOrCreateState(projectID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if _, ok := state.Sketches[sketchID]; !ok {
		http.Error(w, "sketch not found", http.StatusNotFound)
		return
	}

	switch r.Method {
	case http.MethodGet:
		variants := make([]LiraVariant, 0)
		if sketchVariants, ok := state.Variants[sketchID]; ok {
			for _, v := range sketchVariants {
				variants = append(variants, v)
			}
		}
		sort.Slice(variants, func(i, j int) bool {
			return variants[i].Index < variants[j].Index
		})
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(variants)

	case http.MethodPost:
		var variant LiraVariant
		if err := json.NewDecoder(r.Body).Decode(&variant); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		if variant.ID == "" {
			variant.ID = fmt.Sprintf("var-%d", time.Now().UnixNano())
		}
		variant.Schema = "lira.variant.v1"
		variant.SketchID = sketchID
		if variant.CreatedAt == "" {
			variant.CreatedAt = time.Now().UTC().Format(time.RFC3339Nano)
		}
		if variant.ExportIDs == nil {
			variant.ExportIDs = []string{}
		}

		s.hub.mu.Lock()
		if state.Variants[sketchID] == nil {
			state.Variants[sketchID] = make(map[string]LiraVariant)
		}
		state.Variants[sketchID][variant.ID] = variant
		s.hub.mu.Unlock()
		if err := s.hub.saveProject(projectID, state); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(variant)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *server) handleVariant(w http.ResponseWriter, r *http.Request) {
	s.writeCORS(w, r)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	projectID := getProjectParam(r)
	sketchID := extractSketchID(r.URL.Path)
	variantID := extractVariantID(r.URL.Path)

	state, err := s.hub.getOrCreateState(projectID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var variant LiraVariant
	if err := json.NewDecoder(r.Body).Decode(&variant); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	variant.ID = variantID
	variant.Schema = "lira.variant.v1"
	variant.SketchID = sketchID

	s.hub.mu.Lock()
	if state.Variants[sketchID] == nil {
		state.Variants[sketchID] = make(map[string]LiraVariant)
	}
	state.Variants[sketchID][variantID] = variant
	s.hub.mu.Unlock()
	if err := s.hub.saveProject(projectID, state); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(variant)
}

// --- Exports handlers ---

func (s *server) handleExports(w http.ResponseWriter, r *http.Request) {
	s.writeCORS(w, r)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	projectID := getProjectParam(r)
	state, err := s.hub.getOrCreateState(projectID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	switch r.Method {
	case http.MethodGet:
		exports := make([]LiraExport, 0, len(state.Exports))
		for _, e := range state.Exports {
			exports = append(exports, e)
		}
		sort.Slice(exports, func(i, j int) bool {
			return exports[i].CreatedAt > exports[j].CreatedAt
		})
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(exports)

	case http.MethodPost:
		var exp LiraExport
		if err := json.NewDecoder(r.Body).Decode(&exp); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		if exp.ID == "" {
			exp.ID = fmt.Sprintf("exp-%d", time.Now().UnixNano())
		}
		exp.Schema = "lira.export.v1"
		if exp.CreatedAt == "" {
			exp.CreatedAt = time.Now().UTC().Format(time.RFC3339Nano)
		}

		s.hub.mu.Lock()
		state.Exports[exp.ID] = exp
		s.hub.mu.Unlock()
		if err := s.hub.saveProject(projectID, state); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(exp)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *server) handleExport(w http.ResponseWriter, r *http.Request) {
	s.writeCORS(w, r)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	projectID := getProjectParam(r)
	exportID := extractSketchID(r.URL.Path) // reusing extraction from path

	state, err := s.hub.getOrCreateState(projectID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	exp, ok := state.Exports[exportID]
	if !ok {
		http.Error(w, "export not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(exp)
}

// --- UI serving ---

func fileExistsFS(fsys fs.FS, path string) bool {
	if fsys == nil {
		return false
	}
	info, err := fs.Stat(fsys, path)
	return err == nil && !info.IsDir()
}

func embeddedUIAvailable() bool {
	return fileExistsFS(embeddedUI, "index.html")
}

func resolveUIFS(uiDir string) fs.FS {
	uiDir = strings.TrimSpace(uiDir)
	switch {
	case uiDir != "":
		return os.DirFS(uiDir)
	case embeddedUIAvailable():
		return embeddedUI
	default:
		return nil
	}
}

func (s *server) serveHTML(w http.ResponseWriter, r *http.Request, path string) {
	html, err := fs.ReadFile(s.uiFS, path)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if r.Method == http.MethodHead {
		return
	}
	w.Write(html)
}

func (s *server) handleUI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if s.uiFS == nil {
		http.NotFound(w, r)
		return
	}

	cleanPath := pathpkg.Clean("/" + r.URL.Path)
	relativePath := strings.TrimPrefix(cleanPath, "/")
	if relativePath == "" || relativePath == "." {
		relativePath = "index.html"
	}

	serveIfExists := func(path string) bool {
		if !fileExistsFS(s.uiFS, path) {
			return false
		}
		if pathpkg.Ext(path) == ".html" {
			s.serveHTML(w, r, path)
			return true
		}
		http.ServeFileFS(w, r, s.uiFS, path)
		return true
	}

	if pathpkg.Ext(relativePath) != "" {
		if serveIfExists(relativePath) {
			return
		}
		http.NotFound(w, r)
		return
	}

	if serveIfExists(relativePath + ".html") {
		return
	}
	if serveIfExists(pathpkg.Join(relativePath, "index.html")) {
		return
	}
	if serveIfExists("index.html") {
		return
	}
	http.NotFound(w, r)
}

// --- Routing ---

func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}

func (s *server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.handleHealth)
	mux.HandleFunc("/api/version", s.handleVersion)
	mux.HandleFunc("/api/project", s.handleProject)

	// Sketches collection
	mux.HandleFunc("/api/sketches", s.handleSketches)
	// Sketches with / at end (for collection operations like list/create with trailing slash)
	mux.HandleFunc("/api/sketches/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		// Check if it's a variant sub-route
		if strings.Contains(path, "/variants/") {
			s.handleVariant(w, r)
			return
		}
		if strings.HasSuffix(path, "/variants") || strings.Contains(path, "/variants?") {
			s.handleVariants(w, r)
			return
		}
		s.handleSketch(w, r)
	})

	mux.HandleFunc("/api/exports", s.handleExports)
	mux.HandleFunc("/api/exports/", s.handleExport)

	if s.uiFS != nil {
		mux.HandleFunc("/", s.handleUI)
	} else {
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			s.writeCORS(w, r)
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			if r.URL.Path != "/" {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			fmt.Fprintf(w, "lira %s (%s)\n", s.appVersion, s.appSHA)
		})
	}

	return requestLogger(mux)
}

// --- Run ---

func run(ctx context.Context, addr, dataDir, allowOrigin, uiDir string) error {
	srv := &server{
		hub:         newHub(dataDir),
		allowOrigin: allowOrigin,
		uiFS:        resolveUIFS(uiDir),
		appVersion:  version,
		appSHA:      resolveGitSHA(),
	}

	httpServer := &http.Server{
		Addr:    addr,
		Handler: srv.routes(),
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		httpServer.Shutdown(shutdownCtx)
	}()

	log.Printf(asciiLogo)
	log.Printf("lira %s (%s)", srv.appVersion, srv.appSHA)
	if srv.uiFS != nil {
		log.Printf("listening on %s (data dir: %s, ui: embedded)", addr, dataDir)
	} else {
		log.Printf("listening on %s (data dir: %s, ui: unavailable)", addr, dataDir)
	}
	return httpServer.ListenAndServe()
}

const asciiLogo = `
  ▐▛███▜▌  ▐▛███▜▌  ▐▛███▜▌  ▐▛███▜▌
   ▐▌      ▐▌       ▐▌  ▐▌  ▐▌  ▐▌
  ▐▛███▜▌  ▐▌       ▐▛███▜▌  ▐▛███▜▌
     ▐▌    ▐▌       ▐▌ ▐▌   ▐▌  ▐▌
  ▐▛███▜▌  ▐▛███▜▌  ▐▌  ▐▌  ▐▌  ▐▌
                lira - sound sketchbook
`

// --- CLI ---

func serveCommand(args []string) error {
	flags := flag.NewFlagSet("lira", flag.ContinueOnError)
	flags.Usage = printUsage

	addr := flags.String("addr", ":8090", "address to listen on")
	dataDir := flags.String("data-dir", "./data", "directory where project data is stored")
	allowOrigin := flags.String("allow-origin", "*", "CORS allowed origin")
	uiDir := flags.String("ui-dir", "", "optional directory to serve static UI files from")

	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected arguments: %s", strings.Join(flags.Args(), " "))
	}
	return run(context.Background(), *addr, *dataDir, *allowOrigin, *uiDir)
}

func exportCommand(args []string) error {
	flags := flag.NewFlagSet("lira export", flag.ContinueOnError)
	project := flags.String("project", "default", "project name")
	sketch := flags.String("sketch", "", "sketch name")
	format := flags.String("format", "json", "export format: wav, json, zip")

	if err := flags.Parse(args); err != nil {
		return err
	}

	if *sketch == "" {
		return fmt.Errorf("--sketch is required for export")
	}

	fmt.Printf("Export stub: project=%s sketch=%s format=%s\n", *project, *sketch, *format)
	fmt.Println("Note: Audio export is implemented in the browser UI for MVP.")
	fmt.Println("Use the web interface at http://localhost:8090 to export sounds.")
	return nil
}

func printUsage() {
	fmt.Print(`lira is a fast sound sketchbook.

Usage:
  lira
  lira serve [flags]
  lira version
  lira export --sketch <name> [flags]

Flags:
  --addr string        address to listen on (default ":8090")
  --data-dir string    directory where project data is stored (default "./data")
  --allow-origin string CORS allowed origin (default "*")
  --ui-dir string      optional directory to serve static UI files from

Export flags:
  --project string     project name (default "default")
  --sketch string      sketch name (required)
  --format string      export format: json (default "json")
`)
}

func runCLI(args []string) error {
	if len(args) == 0 {
		return serveCommand(nil)
	}

	switch args[0] {
	case "help", "-h", "--help":
		printUsage()
		return nil
	case "version", "--version":
		fmt.Println(version)
		return nil
	case "serve":
		return serveCommand(args[1:])
	case "export":
		return exportCommand(args[1:])
	default:
		return serveCommand(args)
	}
}

func main() {
	if err := runCLI(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
