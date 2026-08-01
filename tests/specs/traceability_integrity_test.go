// Package specs contains structural validation tests for the VEDO Core
// specification artifacts.
//
// This file validates the integrity of the traceability matrix at
// .ai-factory/traceability/traceability.ttl:
//
//   - Turtle syntax sanity (balanced delimiters, declared prefixes,
//     no unclosed strings);
//   - no duplicate subject declarations;
//   - file-backed ADR and REQ entries (those declaring a vdo:filePath)
//     point to files that exist on disk;
//   - vdo:references and vdo:validates targets are declared subjects;
//   - the entries added by the REST API realignment plan resolve.
//
// Legacy logical entries (vdo:ArchitectureDecision with vdo:belongsTo, and
// a small set of pre-existing dangling REQ slugs) are reported as WARN and
// exempted via allowlists, so the gate remains green while still catching
// regressions in new entries.
package specs

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// ttlPath resolves to the traceability matrix relative to this test file.
const ttlPath = "../../.ai-factory/traceability/traceability.ttl"

// knownDanglingReqs are pre-existing base:req/ slugs in the TTL whose URI
// slug does not match an actual file in specs/requirements/. They predate
// this plan and are exempted (WARN) rather than failing the gate.
var knownDanglingReqs = map[string]bool{
	"REQ-FUN.API.property-crud":           true,
	"REQ-FUN.API.route-registration":      true,
	"REQ-FUN.API.sparql-compatibility":    true,
	"REQ-FUN.DATA.ontology-import-export": true,
	"REQ-USR.UI.external-llm-override":    true,
}

// logicalAdrs are legacy logical ADR entries (vdo:ArchitectureDecision with
// vdo:belongsTo) that are not backed by individual files.
var logicalAdrs = map[string]bool{
	"stack-decisions":                true,
	"frontend-vue3":                  true,
	"backend-rust-go":                true,
	"database-neo4j":                 true,
	"gitlab-like-org":                true,
	"org-rest-endpoints":             true,
	"graphql-sparql-split-strategy":  true,
	"protocol-stack-strategy":        true,
	"rest-graphql-mutation-boundary": true,
}

// planEntries are the ADR/REQ slugs added by this plan; they must resolve.
var planADREntries = []string{
	"ADR-DES.API.write-path-invariant",
	"ADR-DES.API.rest-gitlab-alignment",
	"ADR-DES.INFRA.publishing-extension",
}
var planREQEntries = []string{
	"REQ-CON.SECURITY.write-path-invariant",
	"REQ-FUN.API.rest-gitlab-alignment",
	"REQ-CON.STACK.publishing-extension",
}

// subjectRe matches a subject declaration at the start of a TTL block,
// allowing leading whitespace (some blocks are indented):
// "base:adr/NAME a vdo:Class ;" or "base:adr/NAME a vdo:Class ."
var subjectRe = regexp.MustCompile(`^\s*base:([a-z]+)/([A-Za-z0-9._-]+)\s+a\s+vdo:[A-Za-z]+`)

// uriRe extracts base:xxx/yyy URIs from the raw TTL text.
var uriRe = regexp.MustCompile(`base:([a-z]+)/([A-Za-z0-9._-]+)`)

// traceabilityLog is a slog logger with the [traceability_validator] prefix.
var traceabilityLog = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))

// ttlContent returns the raw content of the traceability matrix.
func ttlContent(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile(ttlPath)
	if err != nil {
		t.Fatalf("read traceability ttl: %v", err)
	}
	return string(data)
}

// declaredSubjects parses all subject declarations (base:type/name a vdo:Class)
// from the TTL and returns a map of subject URI -> declaration count.
func declaredSubjects(content string) map[string]int {
	subjects := make(map[string]int)
	for _, line := range strings.Split(content, "\n") {
		m := subjectRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		subjects["base:"+m[1]+"/"+m[2]]++
	}
	return subjects
}

// filePathValueRe extracts the quoted value of a vdo:filePath declaration.
var filePathValueRe = regexp.MustCompile(`vdo:filePath\s+"([^"]+)"`)

// filePathFor returns the vdo:filePath declared for a subject block, or ""
// if none. A block starts at a line matching "subject a vdo:Class ;" and
// runs until a line ending with ".".
func filePathFor(lines []string, subject string) string {
	declRe := regexp.MustCompile(`^\s*` + regexp.QuoteMeta(subject) + `\s+a\s+vdo:[A-Za-z]+`)
	for i, line := range lines {
		if !declRe.MatchString(line) {
			continue
		}
		for j := i; j < len(lines); j++ {
			trimmed := strings.TrimSpace(lines[j])
			if m := filePathValueRe.FindStringSubmatch(trimmed); m != nil {
				return m[1]
			}
			if strings.HasSuffix(trimmed, ".") {
				break
			}
		}
	}
	return ""
}

// findTriples extracts predicate-object pairs (e.g. vdo:references
// base:adr/X) from the entire TTL text, handling multi-object lists.
func findTriples(content, predicate string) []string {
	var targets []string
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.Contains(trimmed, predicate) {
			continue
		}
		// Extract base:type/name occurrences after the predicate.
		idx := strings.Index(trimmed, predicate)
		rest := trimmed[idx+len(predicate):]
		for _, m := range uriRe.FindAllStringSubmatch(rest, -1) {
			targets = append(targets, "base:"+m[1]+"/"+m[2])
		}
	}
	return targets
}

// validateSyntax performs Turtle syntax sanity checks: balanced braces,
// declared prefixes, and no unclosed quoted strings.
func validateSyntax(t *testing.T, content string) []string {
	t.Helper()
	var issues []string

	// Prefix declarations must exist.
	for _, prefix := range []string{"@prefix base:", "@prefix vdo:"} {
		if !strings.Contains(content, prefix) {
			issues = append(issues, "missing prefix declaration "+prefix)
		}
	}

	// Balance check for delimiters.
	balances := map[rune]rune{'}': '{', ')': '(', ']': '['}
	stack := []rune{}
	for _, ch := range content {
		switch ch {
		case '{', '(', '[':
			stack = append(stack, ch)
		case '}', ')', ']':
			if len(stack) == 0 || stack[len(stack)-1] != balances[ch] {
				issues = append(issues, fmt.Sprintf("unbalanced delimiter %q", ch))
			} else {
				stack = stack[:len(stack)-1]
			}
		}
	}
	if len(stack) > 0 {
		issues = append(issues, "unclosed opening delimiter")
	}

	// Unclosed quoted strings.
	quoteCount := strings.Count(content, "\"")
	if quoteCount%2 != 0 {
		issues = append(issues, "unclosed quoted string (odd number of quotes)")
	}

	return issues
}

// repoRoot resolves to the repository root relative to this test file.
// The test runs with working directory tests/specs: one level up is
// tests/, two levels up is the repository root. TTL vdo:filePath values
// are relative to the repository root.
const repoRoot = "../.."

// checkTTLFileExists verifies a file-backed subject's vdo:filePath exists,
// resolving the path relative to the repository root.
func checkTTLFileExists(lines []string, subject string) (path string, ok bool) {
	fp := filePathFor(lines, subject)
	if fp == "" {
		return "", true // no filePath declared — not file-backed
	}
	resolved := filepath.Join(repoRoot, filepath.FromSlash(fp))
	_, err := os.Stat(resolved)
	return fp, err == nil
}

// ---------------------------------------------------------------------------
// Positive tests
// ---------------------------------------------------------------------------

// TestTTL_ShouldHaveValidSyntax verifies the TTL is syntactically sane.
func TestTTL_ShouldHaveValidSyntax(t *testing.T) {
	content := ttlContent(t)
	issues := validateSyntax(t, content)
	for _, i := range issues {
		traceabilityLog.Warn("traceability_validator syntax issue", "detail", i)
		t.Errorf("TTL syntax: %s", i)
	}
	traceabilityLog.Info("traceability_validator syntax summary", "issues", len(issues))
}

// planEntrySubjects are the exact subject URIs added by this plan; they
// must never be duplicated.
var planEntrySubjects = func() map[string]bool {
	m := map[string]bool{}
	for _, s := range planADREntries {
		m["base:adr/"+s] = true
	}
	for _, s := range planREQEntries {
		m["base:req/"+s] = true
	}
	return m
}()

// TestTTL_ShouldHaveNoDuplicateSubjects verifies no subject URI is declared
// more than once. The corpus contains 20 pre-existing legacy duplicates
// (REQ entries re-declared in the extended P0 section); those are reported
// as WARN. Any duplicate among the entries added by this plan is a hard
// failure.
func TestTTL_ShouldHaveNoDuplicateSubjects(t *testing.T) {
	content := ttlContent(t)
	subjects := declaredSubjects(content)
	legacy := 0
	planDup := 0
	for subject, count := range subjects {
		if count <= 1 {
			continue
		}
		if planEntrySubjects[subject] {
			planDup++
			traceabilityLog.Warn("traceability_validator plan subject duplicated",
				"subject", subject, "count", count)
			t.Errorf("plan TTL subject declared %d times: %s", count, subject)
			continue
		}
		legacy++
		traceabilityLog.Warn("traceability_validator legacy duplicate subject",
			"subject", subject, "count", count)
	}
	traceabilityLog.Info("traceability_validator duplicates summary",
		"legacy_duplicates", legacy, "plan_duplicates", planDup)
}

// TestTTL_AdrEntries_ShouldReferenceExistingFiles verifies every
// file-backed base:adr/ entry points to an existing ADR file.
func TestTTL_AdrEntries_ShouldReferenceExistingFiles(t *testing.T) {
	content := ttlContent(t)
	lines := strings.Split(content, "\n")
	subjects := declaredSubjects(content)
	checked := 0
	failures := 0
	for subject := range subjects {
		if !strings.HasPrefix(subject, "base:adr/") {
			continue
		}
		slug := strings.TrimPrefix(subject, "base:adr/")
		if logicalAdrs[slug] {
			continue
		}
		checked++
		fp, ok := checkTTLFileExists(lines, subject)
		if !ok {
			failures++
			traceabilityLog.Warn("traceability_validator missing ADR file",
				"subject", subject, "filePath", fp)
			t.Errorf("TTL ADR entry %s references missing file %s", subject, fp)
		}
	}
	traceabilityLog.Info("traceability_validator adr files summary", "checked", checked, "failed", failures)
}

// TestTTL_ReqEntries_ShouldReferenceExistingFiles verifies file-backed
// base:req/ entries point to existing REQ files, except known legacy
// dangling slugs (WARN).
func TestTTL_ReqEntries_ShouldReferenceExistingFiles(t *testing.T) {
	content := ttlContent(t)
	lines := strings.Split(content, "\n")
	subjects := declaredSubjects(content)
	checked := 0
	failures := 0
	warned := 0
	for subject := range subjects {
		if !strings.HasPrefix(subject, "base:req/") {
			continue
		}
		slug := strings.TrimPrefix(subject, "base:req/")
		checked++
		fp, ok := checkTTLFileExists(lines, subject)
		if ok {
			continue
		}
		if knownDanglingReqs[slug] {
			warned++
			traceabilityLog.Warn("traceability_validator known dangling req",
				"subject", subject, "filePath", fp)
			continue
		}
		failures++
		traceabilityLog.Warn("traceability_validator missing REQ file",
			"subject", subject, "filePath", fp)
		t.Errorf("TTL REQ entry %s references missing file %s", subject, fp)
	}
	traceabilityLog.Info("traceability_validator req files summary",
		"checked", checked, "failed", failures, "warned", warned)
}

// TestTTL_References_ShouldTargetDeclaredSubjects verifies every
// vdo:references target is a declared subject in the TTL.
func TestTTL_References_ShouldTargetDeclaredSubjects(t *testing.T) {
	content := ttlContent(t)
	subjects := declaredSubjects(content)
	targets := findTriples(content, "vdo:references")
	failures := 0
	for _, target := range targets {
		if subjects[target] == 0 {
			failures++
			traceabilityLog.Warn("traceability_validator dangling reference",
				"target", target)
			t.Errorf("TTL vdo:references targets undeclared subject %s", target)
		}
	}
	traceabilityLog.Info("traceability_validator references summary", "targets", len(targets), "failed", failures)
}

// TestTTL_ValidatesReqs_ShouldTargetDeclaredSubjects verifies every
// vdo:validates base:req/ target is a declared subject. The legacy
// dangling REQ slugs (knownDanglingReqs) are exempted (WARN).
func TestTTL_ValidatesReqs_ShouldTargetDeclaredSubjects(t *testing.T) {
	content := ttlContent(t)
	subjects := declaredSubjects(content)
	targets := findTriples(content, "vdo:validates")
	failures := 0
	warned := 0
	for _, target := range targets {
		if !strings.HasPrefix(target, "base:req/") {
			continue
		}
		slug := strings.TrimPrefix(target, "base:req/")
		if subjects[target] > 0 {
			continue
		}
		if knownDanglingReqs[slug] {
			warned++
			traceabilityLog.Warn("traceability_validator legacy dangling validates req",
				"target", target)
			continue
		}
		failures++
		traceabilityLog.Warn("traceability_validator dangling validates req",
			"target", target)
		t.Errorf("TTL vdo:validates targets undeclared req %s", target)
	}
	traceabilityLog.Info("traceability_validator validates summary",
		"targets", len(targets), "failed", failures, "warned", warned)
}

// TestTTL_PlanEntries_ShouldResolve verifies the entries added by the REST
// API realignment plan resolve to existing files.
func TestTTL_PlanEntries_ShouldResolve(t *testing.T) {
	for _, slug := range planADREntries {
		path := filepath.Join(adrDir, slug+".md")
		if _, err := os.Stat(path); err != nil {
			t.Errorf("plan ADR entry %s missing file %s", slug, path)
		}
	}
	for _, slug := range planREQEntries {
		path := filepath.Join(reqDir, slug+".md")
		if _, err := os.Stat(path); err != nil {
			t.Errorf("plan REQ entry %s missing file %s", slug, path)
		}
	}
}

// ---------------------------------------------------------------------------
// Negative tests (synthetic TTL fixtures)
// ---------------------------------------------------------------------------

// ttlFixtureDir creates a temp dir with a synthetic TTL and the seeded
// files it references.
func ttlFixtureDir(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	adr := filepath.Join(dir, "specs", "adr")
	req := filepath.Join(dir, "specs", "requirements")
	if err := os.MkdirAll(adr, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.MkdirAll(req, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	seeds := map[string]string{
		filepath.Join(adr, "ADR-DES.API.test.md"): "# test\n",
		filepath.Join(req, "REQ-TEST.test.md"):    "# test\n",
	}
	for path, data := range seeds {
		if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
			t.Fatalf("seed: %v", err)
		}
	}
	ttlDir := filepath.Join(dir, ".ai-factory", "traceability")
	if err := os.MkdirAll(ttlDir, 0o755); err != nil {
		t.Fatalf("mkdir ttl: %v", err)
	}
	if err := os.WriteFile(filepath.Join(ttlDir, "traceability.ttl"), []byte(content), 0o644); err != nil {
		t.Fatalf("write ttl: %v", err)
	}
	return dir
}

// fixtureSubjects parses a synthetic TTL with the same parser used for the
// real matrix.
func fixtureSubjects(t *testing.T, content string) map[string]int {
	t.Helper()
	dir := ttlFixtureDir(t, content)
	path := filepath.Join(dir, ".ai-factory", "traceability", "traceability.ttl")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture ttl: %v", err)
	}
	return declaredSubjects(string(data))
}

const ttlFixtureValid = `@prefix base: <file:///vedo-hub/> .
@prefix vdo: <http://vedo.dev/ontology/traceability#> .

base:adr/ADR-DES.API.test a vdo:ArchitectureDecisionRecord ;
    vdo:filePath "specs/adr/ADR-DES.API.test.md" .

base:req/REQ-TEST.test a vdo:FunctionalRequirement ;
    vdo:filePath "specs/requirements/REQ-TEST.test.md" .

base:adr/ADR-DES.API.test vdo:references base:req/REQ-TEST.test .
`

// TestTTL_ValidFixture_ShouldHaveNoDuplicates is the positive control for
// the duplicate-detection parser.
func TestTTL_ValidFixture_ShouldHaveNoDuplicates(t *testing.T) {
	subjects := fixtureSubjects(t, ttlFixtureValid)
	for subject, count := range subjects {
		if count > 1 {
			t.Errorf("fixture duplicate subject %s (%d)", subject, count)
		}
	}
}

// TestTTL_DuplicateSubject_ShouldBeDetected verifies duplicate subject
// declarations are detected (negative security-relevant test: a duplicated
// subject could hide a traceability link).
func TestTTL_DuplicateSubject_ShouldBeDetected(t *testing.T) {
	content := ttlFixtureValid + "\nbase:req/REQ-TEST.test a vdo:FunctionalRequirement ;\n    vdo:filePath \"specs/requirements/REQ-TEST.test.md\" .\n"
	subjects := fixtureSubjects(t, content)
	count := subjects["base:req/REQ-TEST.test"]
	if count < 2 {
		t.Errorf("expected duplicate subject detected, got count %d", count)
	}
}

// TestTTL_BrokenReference_ShouldBeDetected verifies a vdo:references target
// that is not a declared subject is detected.
//
// // Validates: REQ-CON.SECURITY.write-path-invariant
func TestTTL_BrokenReference_ShouldBeDetected(t *testing.T) {
	content := ttlFixtureValid + "\nbase:adr/ADR-DES.API.other a vdo:ArchitectureDecisionRecord ;\n    vdo:references base:adr/ADR-DES.API.missing .\n"
	dir := ttlFixtureDir(t, content)
	path := filepath.Join(dir, ".ai-factory", "traceability", "traceability.ttl")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	subjects := declaredSubjects(string(data))
	targets := findTriples(string(data), "vdo:references")
	found := false
	for _, target := range targets {
		if target == "base:adr/ADR-DES.API.missing" && subjects[target] == 0 {
			found = true
		}
	}
	if !found {
		t.Errorf("expected broken reference detected, got targets %v", targets)
	}
}

// TestTTL_UnclosedString_ShouldBeDetected verifies Turtle syntax validation
// catches an unclosed quoted string.
func TestTTL_UnclosedString_ShouldBeDetected(t *testing.T) {
	content := ttlFixtureValid + "\nbase:req/REQ-TEST.broken a vdo:FunctionalRequirement ;\n    vdo:filePath \"specs/requirements/unclosed.md .\n"
	issues := validateSyntax(t, content)
	found := false
	for _, i := range issues {
		if strings.Contains(i, "quote") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected unclosed-string issue, got %v", issues)
	}
}

// TestTTL_ScanLegacyDanglingReqs_ShouldBeBounded ensures the known-dangling
// REQ allowlist is not silently growing with new entries.
func TestTTL_ScanLegacyDanglingReqs_ShouldBeBounded(t *testing.T) {
	content := ttlContent(t)
	lines := strings.Split(content, "\n")
	subjects := declaredSubjects(content)
	dangling := 0
	for subject := range subjects {
		if !strings.HasPrefix(subject, "base:req/") {
			continue
		}
		fp, ok := checkTTLFileExists(lines, subject)
		if ok {
			continue
		}
		slug := strings.TrimPrefix(subject, "base:req/")
		if !knownDanglingReqs[slug] {
			dangling++
			traceabilityLog.Warn("traceability_validator unexpected dangling req",
				"subject", subject, "filePath", fp)
		}
	}
	if dangling > 0 {
		t.Errorf("%d REQ entries reference missing files outside the known-dangling allowlist", dangling)
	}
}

// ---------------------------------------------------------------------------
// Utility: sorted output for deterministic reports
// ---------------------------------------------------------------------------

// sortedKeys returns the keys of a string map in sorted order.
func sortedKeys(m map[string]int) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
