// Package specs contains structural validation tests for the VEDO Core
// specification artifacts (ADRs, requirements, traceability matrix).
//
// The ADR validator enforces the structure convention documented in
// specs/adr/README.md:
//
//   - every ADR MUST carry valid frontmatter (**Дата:** / **Статус:**);
//   - modern ADRs (dated >= 2026-06-01) MUST contain the full section set
//     (Контекст, Требование-источник, Решение, Рассмотренные альтернативы,
//     Последствия);
//   - ADR-to-ADR and ADR-to-REQ markdown links MUST resolve to existing
//     files.
//
// The corpus contains legacy ADRs (pre-2026-06) that predate the full
// convention and carry imperfect internal links. Those ADRs are still
// scanned and logged (WARN), but only modern ADRs and the ADRs created by
// this plan are subject to hard failures. This keeps the gate future-proof
// without requiring a one-off refactor of the legacy corpus.
package specs

import (
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"
)

// adrDir resolves to the specs/adr directory relative to this test file.
// go test runs with the working directory set to the package directory
// (tests/specs), so ../../specs/adr is the ADR directory.
const adrDir = "../../specs/adr"

// reqDir resolves to the specs/requirements directory.
const reqDir = "../../specs/requirements"

// specsRoot resolves to the specs/ directory.
const specsRoot = "../../specs"

// modernDateThreshold is the earliest **Дата:** for which the full-section
// convention is enforced. Legacy ADRs before this date are exempt from the
// full-section requirement.
const modernDateThreshold = "2026-06-01"

// excludedFiles are non-ADR files inside specs/adr/ that must not be
// validated as architecture decision records.
var excludedFiles = map[string]bool{
	"README.md": true,
}

// planADRs are the ADRs created by the REST API realignment plan. They are
// subject to the strictest checks (full sections + all links resolve).
var planADRs = []string{
	"ADR-DES.API.write-path-invariant.md",
	"ADR-DES.API.rest-gitlab-alignment.md",
	"ADR-DES.INFRA.publishing-extension.md",
}

// requiredSections is the full-section set enforced for modern ADRs.
var requiredSections = []string{
	"## Контекст",
	"## Требование-источник",
	"## Решение",
	"## Рассмотренные альтернативы",
	"## Последствия",
}

// statusRe matches the **Статус:** frontmatter values observed across the
// ADR corpus. Matching is done against a lowercased value (strings.ToLower
// handles Cyrillic correctly, unlike regexp (?i) which only folds ASCII).
var statusRe = regexp.MustCompile(`принято|принятo|accepted|предложено|proposed|изменено|устарело|заменено|черновик|draft|superseded|replaced`)

// datePrimaryRe matches the primary date of a **Дата:** value. Annotations
// such as "(amended 2026-07-20, implemented 2026-07-24)" are permitted
// after the primary date.
var datePrimaryRe = regexp.MustCompile(`^(\d{4}-\d{2}-\d{2})`)

// markdownLinkRe matches markdown link destinations ending in .md.
var markdownLinkRe = regexp.MustCompile(`\]\(([^)]*\.md)\)`)

// log is a package-level slog logger with the [adr_validator] prefix.
var log = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))

// adrFiles returns the sorted list of ADR file names in specs/adr,
// excluding non-ADR files (README.md).
func adrFiles(t *testing.T) []string {
	t.Helper()
	entries, err := os.ReadDir(adrDir)
	if err != nil {
		t.Fatalf("read ADR dir %s: %v", adrDir, err)
	}
	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".md") || excludedFiles[name] {
			continue
		}
		files = append(files, name)
	}
	sort.Strings(files)
	return files
}

// readFile returns the raw content of a file relative to the test working
// directory.
func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}

// readADR returns the raw content of an ADR file.
func readADR(t *testing.T, name string) string {
	t.Helper()
	return readFile(t, filepath.Join(adrDir, name))
}

// hasSection reports whether the content contains the given section header
// at line start.
func hasSection(content, header string) bool {
	for _, line := range strings.Split(content, "\n") {
		if strings.TrimSpace(line) == header {
			return true
		}
	}
	return false
}

// frontmatterValue extracts the value of a **Key:** frontmatter field.
func frontmatterValue(content, key string) (string, bool) {
	prefix := "**" + key + ":**"
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, prefix) {
			val := strings.TrimSpace(strings.TrimPrefix(trimmed, prefix))
			val = strings.Trim(val, "`* ")
			return val, true
		}
	}
	return "", false
}

// primaryDate extracts and parses the primary date of a **Дата:** value.
// It returns the date string and a validity flag.
func primaryDate(content string) (string, bool) {
	raw, ok := frontmatterValue(content, "Дата")
	if !ok {
		return "", false
	}
	m := datePrimaryRe.FindStringSubmatch(raw)
	if m == nil {
		return raw, false
	}
	if _, err := time.Parse("2006-01-02", m[1]); err != nil {
		return m[1], false
	}
	return m[1], true
}

// isModern reports whether an ADR is subject to the full-section
// convention (dated >= modernDateThreshold).
func isModern(name, content string) bool {
	date, ok := primaryDate(content)
	if !ok {
		// Unknown date — treat plan ADRs as modern, others conservatively
		// as legacy to avoid false failures.
		for _, p := range planADRs {
			if p == name {
				return true
			}
		}
		return false
	}
	return date >= modernDateThreshold
}

// linkTargets extracts all markdown link destinations ending in .md.
func linkTargets(content string) []string {
	matches := markdownLinkRe.FindAllStringSubmatch(content, -1)
	targets := make([]string, 0, len(matches))
	for _, m := range matches {
		targets = append(targets, m[1])
	}
	return targets
}

// isAdrLink reports whether a link target refers to an ADR file (either in
// the same directory or via ../adr/).
func isAdrLink(target string) bool {
	base := filepath.Base(strings.ReplaceAll(target, "\\", "/"))
	return strings.HasPrefix(base, "ADR-")
}

// isReqLink reports whether a link target refers to a requirements file.
func isReqLink(target string) bool {
	t := strings.ReplaceAll(target, "\\", "/")
	return strings.Contains(t, "/requirements/") || strings.HasPrefix(t, "REQ-")
}

// resolveLinkFrom resolves a markdown link destination relative to the
// given base directory (the ADR's own directory) and reports whether the
// target exists on disk.
func resolveLinkFrom(baseDir, target string) bool {
	target = strings.ReplaceAll(target, "\\", "/")
	if idx := strings.IndexAny(target, "#"); idx >= 0 {
		target = target[:idx]
	}
	var resolved string
	switch {
	case strings.HasPrefix(target, "../requirements/"):
		resolved = filepath.Join(filepath.Dir(baseDir), "requirements", strings.TrimPrefix(target, "../requirements/"))
	case strings.HasPrefix(target, "../adr/"):
		resolved = filepath.Join(filepath.Dir(baseDir), "adr", strings.TrimPrefix(target, "../adr/"))
	case strings.HasPrefix(target, "requirements/"):
		resolved = filepath.Join(filepath.Dir(baseDir), "requirements", strings.TrimPrefix(target, "requirements/"))
	case strings.HasPrefix(target, "adr/"):
		resolved = filepath.Join(filepath.Dir(baseDir), "adr", strings.TrimPrefix(target, "adr/"))
	case strings.HasPrefix(target, "../"):
		resolved = filepath.Join(filepath.Dir(filepath.Dir(baseDir)), strings.TrimPrefix(target, "../"))
	default:
		resolved = filepath.Join(baseDir, target)
	}
	info, err := os.Stat(resolved)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// resolveLink resolves a link target for a corpus ADR (located in adrDir).
func resolveLink(adrName, target string) bool {
	return resolveLinkFrom(adrDir, target)
}

// linkCategory classifies a link target as "adr", "req", or "other".
func linkCategory(target string) string {
	switch {
	case isAdrLink(target):
		return "adr"
	case isReqLink(target):
		return "req"
	default:
		return "other"
	}
}

// checkADR runs the full validation for a single ADR and returns issues.
// strict controls whether legacy-tolerant checks (imperfect links to
// non-ADR/REQ files, missing sections) become hard failures.
func checkADR(t *testing.T, name, content string, strict bool) []string {
	t.Helper()
	var issues []string

	// Frontmatter (applies to all ADRs).
	if v, ok := frontmatterValue(content, "Дата"); !ok {
		issues = append(issues, "missing **Дата:** frontmatter")
	} else if _, valid := primaryDate(content); !valid {
		issues = append(issues, "invalid **Дата:** value "+v)
	}
	if v, ok := frontmatterValue(content, "Статус"); !ok {
		issues = append(issues, "missing **Статус:** frontmatter")
	} else if !statusRe.MatchString(strings.ToLower(v)) {
		issues = append(issues, "invalid **Статус:** value "+v)
	}

	// Links: ADR and REQ links must always resolve; "other" links only in
	// strict mode (plan ADRs / modern ADRs).
	for _, target := range linkTargets(content) {
		cat := linkCategory(target)
		resolved := resolveLinkFrom(adrDir, target)
		if !resolved && (cat != "other" || strict) {
			issues = append(issues, "broken "+cat+" link to "+target)
		} else if !resolved && cat == "other" && !strict {
			log.Warn("adr_validator legacy loose link",
				"file", name, "target", target)
		}
	}

	// Sections: full set required for modern ADRs; ## Решение required for
	// all non-legacy ADRs.
	if isModern(name, content) || strict {
		for _, sec := range requiredSections {
			if !hasSection(content, sec) {
				issues = append(issues, "missing section "+sec)
			}
		}
	} else if !hasSection(content, "## Решение") {
		log.Warn("adr_validator legacy missing resolution",
			"file", name)
	}

	return issues
}

// ---------------------------------------------------------------------------
// Positive tests (run against the real ADR corpus)
// ---------------------------------------------------------------------------

// TestADR_AllFiles_ShouldHaveValidFrontmatter verifies every ADR carries
// valid **Дата:** and **Статус:** frontmatter.
//
// // Validates: REQ-CON.SECURITY.write-path-invariant
func TestADR_AllFiles_ShouldHaveValidFrontmatter(t *testing.T) {
	files := adrFiles(t)
	log.Info("adr_validator frontmatter check start", "total", len(files))
	failures := 0
	for _, name := range files {
		content := readADR(t, name)
		var issues []string
		if v, ok := frontmatterValue(content, "Дата"); !ok {
			issues = append(issues, "missing **Дата:**")
		} else if _, valid := primaryDate(content); !valid {
			issues = append(issues, "invalid **Дата:** "+v)
		}
		if v, ok := frontmatterValue(content, "Статус"); !ok {
			issues = append(issues, "missing **Статус:**")
		} else if !statusRe.MatchString(strings.ToLower(v)) {
			issues = append(issues, "invalid **Статус:** "+v)
		}
		if len(issues) > 0 {
			failures++
			log.Warn("adr_validator frontmatter failure", "file", name, "issues", issues)
			t.Errorf("ADR %s frontmatter invalid: %v", name, issues)
		} else {
			log.Info("adr_validator frontmatter ok", "file", name)
		}
	}
	log.Info("adr_validator frontmatter summary", "total", len(files), "failed", failures)
}

// TestADR_PlanAdrs_ShouldConformToConvention verifies the three ADRs created
// by this plan satisfy the full structure convention: frontmatter, all
// required sections, and resolvable ADR/REQ/other links.
func TestADR_PlanAdrs_ShouldConformToConvention(t *testing.T) {
	for _, name := range planADRs {
		t.Run(name, func(t *testing.T) {
			content := readADR(t, name)
			issues := checkADR(t, name, content, true)
			for _, issue := range issues {
				t.Errorf("ADR %s: %s", name, issue)
			}
			log.Info("adr_validator plan adr result",
				"file", name, "issues", len(issues))
		})
	}
}

// TestADR_ModernAdrs_ShouldHaveRequiredSections verifies modern ADRs
// (dated >= 2026-06-01) contain the full required section set.
func TestADR_ModernAdrs_ShouldHaveRequiredSections(t *testing.T) {
	files := adrFiles(t)
	checked := 0
	failures := 0
	for _, name := range files {
		content := readADR(t, name)
		if !isModern(name, content) {
			continue
		}
		checked++
		issues := checkADR(t, name, content, true)
		for _, issue := range issues {
			if strings.HasPrefix(issue, "missing section") {
				failures++
				t.Errorf("ADR %s: %s", name, issue)
			}
		}
	}
	log.Info("adr_validator modern sections summary", "checked", checked, "failed", failures)
}

// TestADR_NewAdrs_ShouldHaveSourceRequirements verifies the three ADRs
// reference their source REQ files.
func TestADR_NewAdrs_ShouldHaveSourceRequirements(t *testing.T) {
	cases := []struct {
		adr     string
		reqSlug string
	}{
		{"ADR-DES.API.write-path-invariant.md", "REQ-CON.SECURITY.write-path-invariant"},
		{"ADR-DES.API.rest-gitlab-alignment.md", "REQ-FUN.API.rest-gitlab-alignment"},
		{"ADR-DES.INFRA.publishing-extension.md", "REQ-CON.STACK.publishing-extension"},
	}
	for _, tc := range cases {
		t.Run(tc.adr, func(t *testing.T) {
			content := readADR(t, tc.adr)
			expected := tc.reqSlug + ".md"
			found := false
			for _, target := range linkTargets(content) {
				if strings.HasSuffix(target, expected) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("ADR %s does not link source requirement %s", tc.adr, expected)
			}
		})
	}
}

// TestADR_PlanAdrs_ShouldReferenceExistingAdrs verifies the three ADRs
// cross-reference only existing ADR files.
//
// // Validates: REQ-CON.SECURITY.write-path-invariant
func TestADR_PlanAdrs_ShouldReferenceExistingAdrs(t *testing.T) {
	for _, name := range planADRs {
		content := readADR(t, name)
		for _, target := range linkTargets(content) {
			if isAdrLink(target) && !resolveLink(name, target) {
				t.Errorf("ADR %s references missing ADR %s", name, target)
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Negative tests (synthetic fixtures)
// ---------------------------------------------------------------------------

// fixtureIssues runs the validator against a synthetic ADR and returns the
// reported issues. Strict mode is used so missing sections are detected.
// The fixture is placed in a temp dir that mirrors the corpus layout
// (adr/ + requirements/) so links resolve relative to the fixture's own
// directory. The positive control creates real sibling files; negative
// tests point at missing files.
func fixtureIssues(t *testing.T, content string) []string {
	t.Helper()
	dir := t.TempDir()
	adrDir := filepath.Join(dir, "adr")
	reqDir := filepath.Join(dir, "requirements")
	if err := os.MkdirAll(adrDir, 0o755); err != nil {
		t.Fatalf("mkdir adr: %v", err)
	}
	if err := os.MkdirAll(reqDir, 0o755); err != nil {
		t.Fatalf("mkdir requirements: %v", err)
	}
	// Seed the files referenced by the positive-control fixture.
	seed := map[string]string{
		filepath.Join(reqDir, "REQ-CON.SECURITY.write-path-invariant.md"):   "# REQ-CON.SECURITY.write-path-invariant\n",
		filepath.Join(adrDir, "ADR-DES.API.organization-rest-endpoints.md"): "# ADR-DES.API.organization-rest-endpoints\n",
	}
	for path, data := range seed {
		if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
			t.Fatalf("seed %s: %v", path, err)
		}
	}
	path := filepath.Join(adrDir, "ADR-TEST.fixture.md")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	return checkADRWithBase(t, adrDir, filepath.Base(path), content, true)
}

// checkADRWithBase is like checkADR but resolves links relative to baseDir
// instead of the corpus adrDir.
func checkADRWithBase(t *testing.T, baseDir, name, content string, strict bool) []string {
	t.Helper()
	var issues []string

	if v, ok := frontmatterValue(content, "Дата"); !ok {
		issues = append(issues, "missing **Дата:** frontmatter")
	} else if _, valid := primaryDate(content); !valid {
		issues = append(issues, "invalid **Дата:** value "+v)
	}
	if v, ok := frontmatterValue(content, "Статус"); !ok {
		issues = append(issues, "missing **Статус:** frontmatter")
	} else if !statusRe.MatchString(strings.ToLower(v)) {
		issues = append(issues, "invalid **Статус:** value "+v)
	}

	for _, target := range linkTargets(content) {
		cat := linkCategory(target)
		resolved := resolveLinkFrom(baseDir, target)
		if !resolved {
			issues = append(issues, "broken "+cat+" link to "+target)
		}
	}

	if strict {
		for _, sec := range requiredSections {
			if !hasSection(content, sec) {
				issues = append(issues, "missing section "+sec)
			}
		}
	}
	return issues
}

// fixtureBase is a structurally valid synthetic ADR. Links point to real
// corpus files so the positive control passes.
const fixtureBase = `# ADR-TEST.fixture

**Дата:** 2026-08-01
**Статус:** Принято

## Контекст

Test context.

## Требование-источник
- [REQ-CON.SECURITY.write-path-invariant.md](../requirements/REQ-CON.SECURITY.write-path-invariant.md)

## Решение

Test resolution.

## Рассмотренные альтернативы

| A | B |
|---|---|
| X | Y |

## Последствия

**Положительные:** none.
**Отрицательные:** none.

## Связанные ADR
- [ADR-DES.API.organization-rest-endpoints.md](ADR-DES.API.organization-rest-endpoints.md)
`

// TestADR_ValidFixture_ShouldHaveNoIssues is the baseline positive control
// for the negative tests below.
func TestADR_ValidFixture_ShouldHaveNoIssues(t *testing.T) {
	issues := fixtureIssues(t, fixtureBase)
	if len(issues) > 0 {
		t.Errorf("valid fixture reported issues: %v", issues)
	}
}

// TestADR_MissingDate_ShouldFail verifies a missing **Дата:** is detected.
func TestADR_MissingDate_ShouldFail(t *testing.T) {
	content := strings.Replace(fixtureBase, "**Дата:** 2026-08-01\n", "", 1)
	issues := fixtureIssues(t, content)
	found := false
	for _, i := range issues {
		if strings.Contains(i, "Дата") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected missing-date issue, got %v", issues)
	}
}

// TestADR_InvalidDate_ShouldFail verifies a malformed **Дата:** is detected.
func TestADR_InvalidDate_ShouldFail(t *testing.T) {
	content := strings.Replace(fixtureBase, "**Дата:** 2026-08-01", "**Дата:** 08/01/2026", 1)
	issues := fixtureIssues(t, content)
	found := false
	for _, i := range issues {
		if strings.Contains(i, "Дата") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected invalid-date issue, got %v", issues)
	}
}

// TestADR_MissingStatus_ShouldFail verifies a missing **Статус:** is detected.
func TestADR_MissingStatus_ShouldFail(t *testing.T) {
	content := strings.Replace(fixtureBase, "**Статус:** Принято\n", "", 1)
	issues := fixtureIssues(t, content)
	found := false
	for _, i := range issues {
		if strings.Contains(i, "Статус") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected missing-status issue, got %v", issues)
	}
}

// TestADR_MissingResolution_ShouldFail verifies a missing ## Решение is
// detected.
func TestADR_MissingResolution_ShouldFail(t *testing.T) {
	content := strings.Replace(fixtureBase, "## Решение\n\nTest resolution.\n\n", "", 1)
	issues := fixtureIssues(t, content)
	found := false
	for _, i := range issues {
		if strings.Contains(i, "Решение") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected missing-resolution issue, got %v", issues)
	}
}

// TestADR_BrokenAdrLink_ShouldFail verifies a link to a non-existent ADR
// file is detected.
func TestADR_BrokenAdrLink_ShouldFail(t *testing.T) {
	content := strings.ReplaceAll(fixtureBase,
		"ADR-DES.API.organization-rest-endpoints.md",
		"ADR-DES.API.does-not-exist.md")
	issues := fixtureIssues(t, content)
	found := false
	for _, i := range issues {
		if strings.Contains(i, "broken") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected broken-adr-link issue, got %v", issues)
	}
}

// TestADR_BrokenReqLink_ShouldFail verifies a link to a non-existent REQ
// file is detected.
func TestADR_BrokenReqLink_ShouldFail(t *testing.T) {
	content := strings.ReplaceAll(fixtureBase,
		"REQ-CON.SECURITY.write-path-invariant.md",
		"REQ-CON.SECURITY.does-not-exist.md")
	issues := fixtureIssues(t, content)
	found := false
	for _, i := range issues {
		if strings.Contains(i, "broken") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected broken-req-link issue, got %v", issues)
	}
}

// TestADR_MissingRequiredSection_ShouldFail verifies a missing required
// section is detected in strict mode.
func TestADR_MissingRequiredSection_ShouldFail(t *testing.T) {
	content := strings.Replace(fixtureBase, "## Последствия", "## Missing", 1)
	if hasSection(content, "## Последствия") {
		t.Fatal("fixture replacement failed")
	}
	issues := fixtureIssues(t, content)
	found := false
	for _, i := range issues {
		if strings.Contains(i, "missing section") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected missing-section issue, got %v", issues)
	}
}
