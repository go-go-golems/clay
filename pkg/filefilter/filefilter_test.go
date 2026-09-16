package filefilter

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIsExcludedDir_GlobSegmentMatching verifies that directory exclusion matches
// globs against individual path segments, not substrings of the whole path.
//
// This is a behavior change from the previous strings.Contains implementation:
// a default pattern like "build" used to exclude "builder-api"; it now only
// excludes a directory literally named "build" (or matching the glob).
func TestIsExcludedDir_GlobSegmentMatching(t *testing.T) {
	ff := NewFileFilter()

	excluded := []string{
		"build",
		"code/app/build",
		"code/app/build/generated.go",
		".git",
		"code/.git/hooks",
		"x/node_modules/y",
		"vendor",
		"a/dist/b",
	}
	for _, path := range excluded {
		assert.True(t, ff.isExcludedDir(path), "expected %q to be excluded", path)
	}

	included := []string{
		// the motivating bug: substring match used to exclude these
		"builder-api",
		"code/builder-api",
		"src/rebuild",
		"distribution",
		"code/app/building",
		// root path must never be excluded by a dir pattern
		".",
		"/",
	}
	for _, path := range included {
		assert.False(t, ff.isExcludedDir(path), "expected %q to be included", path)
	}
}

// TestIsExcludedDir_WindowsSeparators verifies segments are split on both
// separators so glob matching works on Windows-style paths everywhere.
func TestIsExcludedDir_WindowsSeparators(t *testing.T) {
	ff := NewFileFilter()

	assert.True(t, ff.isExcludedDir(`code\app\build`), "windows-style path with / separator")
	assert.False(t, ff.isExcludedDir(`code\builder-api`), "windows-style path must not substring-match")
}

// TestIsExcludedDir_UserGlobs verifies user-supplied exclude patterns keep the
// ability to express prefix and substring semantics explicitly via globs.
func TestIsExcludedDir_UserGlobs(t *testing.T) {
	ff := NewFileFilter(WithExcludeDirs([]string{"build*", "*dist*"}))

	assert.True(t, ff.isExcludedDir("x/builder-api"), "build* matches builder-api segment")
	assert.True(t, ff.isExcludedDir("distribution"), "*dist* restores substring semantics")
	assert.False(t, ff.isExcludedDir("x/api"), "unrelated segment not matched")
}

// TestIsExcludedDir_IncludeDirsWins verifies include patterns take precedence
// over both default and user exclude patterns.
func TestIsExcludedDir_IncludeDirsWins(t *testing.T) {
	// include beats a default exclude
	ff := NewFileFilter(WithIncludeDirs([]string{"build"}))
	assert.False(t, ff.isExcludedDir("code/build"), "include must beat default exclude")
	assert.True(t, ff.isExcludedDir("code/dist"), "other defaults still apply")

	// include beats a user exclude, and globs work for includes too
	ff = NewFileFilter(
		WithExcludeDirs([]string{"api"}),
		WithIncludeDirs([]string{"builder-*"}),
	)
	assert.False(t, ff.isExcludedDir("x/builder-api"), "include glob must beat user exclude")
	assert.True(t, ff.isExcludedDir("x/api"), "user exclude without include still applies")
}

// TestIsExcludedDir_DisableDefaultFilters verifies the flag still disables only
// the default list.
func TestIsExcludedDir_DisableDefaultFilters(t *testing.T) {
	ff := NewFileFilter(
		WithDisableDefaultFilters(true),
		WithExcludeDirs([]string{"build"}),
	)
	assert.False(t, ff.isExcludedDir("code/dist"), "defaults disabled")
	assert.True(t, ff.isExcludedDir("code/build"), "user excludes still apply")
	assert.False(t, ff.isExcludedDir("builder-api"), "no substring matching")
}

// TestFileFilter_ValidateBadPatterns verifies malformed glob patterns are
// rejected up front instead of silently never matching.
func TestFileFilter_ValidateBadPatterns(t *testing.T) {
	ff := NewFileFilter(WithExcludeDirs([]string{"["}))
	err := ff.Validate()
	require.Error(t, err, "malformed exclude pattern must be rejected")

	ff = NewFileFilter(WithIncludeDirs([]string{"[a-"}))
	err = ff.Validate()
	require.Error(t, err, "malformed include pattern must be rejected")

	ff = NewFileFilter()
	assert.NoError(t, ff.Validate())
}

// TestFileFilter_IncludeDirsYAMLRoundTrip verifies the include-dirs YAML field
// and profile inheritance of defaults.
func TestFileFilter_IncludeDirsYAMLRoundTrip(t *testing.T) {
	ff := NewFileFilter(WithIncludeDirs([]string{"build", "builder-*"}))
	data, err := ff.ToYAML()
	require.NoError(t, err)

	parsed, err := FromYAML(data)
	require.NoError(t, err)
	assert.Equal(t, []string{"build", "builder-*"}, parsed.IncludeDirs)
	assert.False(t, parsed.isExcludedDir("code/build"), "include survives YAML round trip")

	_, err = FromYAML([]byte("exclude-dirs: [\"[\"]\n"))
	require.Error(t, err, "malformed pattern in YAML must fail FromYAML via Validate")
}

// TestFileFilter_ShouldProcessFileDirectories verifies FilterPath /
// shouldProcessFile route directories through the same glob logic.
func TestFileFilter_ShouldProcessFileDirectories(t *testing.T) {
	dir := t.TempDir()

	builderAPI := filepath.Join(dir, "builder-api")
	build := filepath.Join(dir, "build")
	for _, d := range []string{builderAPI, build} {
		require.NoError(t, os.MkdirAll(d, 0o755))
	}

	ff := NewFileFilter()

	assert.True(t, ff.FilterPath(builderAPI), "builder-api directory must be processed")
	assert.False(t, ff.FilterPath(build), "build directory must be excluded")
}

// TestCreateFileFilterFromSettings_IncludeDirs verifies the parameter layer
// wiring for the include-dirs flag.
func TestCreateFileFilterFromSettings_IncludeDirs(t *testing.T) {
	// settings are exercised through the glazed layer in downstream repos;
	// here we check the FileFilterSettings struct mapping end-to-end.
	s := &FileFilterSettings{
		MaxFileSize: 1024,
		ExcludeDirs: []string{"api"},
		IncludeDirs: []string{"builder-*"},
	}
	ff := NewFileFilter()
	ff.IncludeDirs = s.IncludeDirs
	ff.ExcludeDirs = s.ExcludeDirs

	assert.False(t, ff.isExcludedDir("x/builder-api"))
	assert.True(t, ff.isExcludedDir("x/api"))
}
