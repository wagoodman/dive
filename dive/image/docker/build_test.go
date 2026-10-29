package docker

import (
	"github.com/stretchr/testify/require"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestIsFileFlagsAreSet(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		flags    []string
		expected bool
	}{
		{
			name:     "flag present in the middle with value",
			args:     []string{"arg1", "-f", "dockerfile", "arg2"},
			flags:    []string{"-f"},
			expected: true,
		},
		{
			name:     "flag present at the beginning with value",
			args:     []string{"-f", "dockerfile", "arg1", "arg2"},
			flags:    []string{"-f"},
			expected: true,
		},
		{
			name:     "flag present at the end with no value",
			args:     []string{"arg1", "arg2", "-f"},
			flags:    []string{"-f"},
			expected: false,
		},
		{
			name:     "flag not present",
			args:     []string{"arg1", "arg2", "arg3"},
			flags:    []string{"-f"},
			expected: false,
		},
		{
			name:     "one of multiple flags present",
			args:     []string{"arg1", "--file", "dockerfile", "arg2"},
			flags:    []string{"-f", "--file"},
			expected: true,
		},
		{
			name:     "none of multiple flags present",
			args:     []string{"arg1", "-x", "value", "arg2"},
			flags:    []string{"-f", "--file"},
			expected: false,
		},
		{
			name:     "empty args",
			args:     []string{},
			flags:    []string{"-f", "--file"},
			expected: false,
		},
		{
			name:     "empty flags",
			args:     []string{"arg1", "-f", "value"},
			flags:    []string{},
			expected: false,
		},
		{
			name:     "flag with multiple values",
			args:     []string{"arg1", "-f", "value1", "value2"},
			flags:    []string{"-f"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isFileFlagsAreSet(tt.args, tt.flags...)
			assert.Equal(t, tt.expected, result, "isFileFlagsAreSet() = %v, want %v", result, tt.expected)
		})
	}
}

func TestTryFindContainerfile(t *testing.T) {
	tests := []struct {
		name           string
		buildArgs      []string
		setupFs        func(t testing.TB, fs afero.Fs)
		expectedPath   string
		expectedErrMsg string
	}{
		{
			name:      "find Containerfile (uppercase)",
			buildArgs: []string{"testdir"},
			setupFs: func(t testing.TB, fs afero.Fs) {
				create(t, fs, "testdir/Containerfile", "FROM alpine")
			},
			expectedPath:   filepath.Join("testdir", "Containerfile"),
			expectedErrMsg: "",
		},
		{
			name:      "find containerfile (lowercase)",
			buildArgs: []string{"testdir"},
			setupFs: func(t testing.TB, fs afero.Fs) {
				create(t, fs, "testdir/containerfile", "FROM alpine")
			},
			expectedPath:   filepath.Join("testdir", "containerfile"),
			expectedErrMsg: "",
		},
		{
			name:      "find Dockerfile when no Containerfile exists",
			buildArgs: []string{"testdir"},
			setupFs: func(t testing.TB, fs afero.Fs) {
				create(t, fs, "testdir/Dockerfile", "FROM alpine")
			},
			expectedPath:   filepath.Join("testdir", "Dockerfile"),
			expectedErrMsg: "",
		},
		{
			name:      "find dockerfile (lowercase)",
			buildArgs: []string{"testdir"},
			setupFs: func(t testing.TB, fs afero.Fs) {
				create(t, fs, "testdir/dockerfile", "FROM alpine")
			},
			expectedPath:   filepath.Join("testdir", "dockerfile"),
			expectedErrMsg: "",
		},
		{
			name:      "prefer Containerfile over Dockerfile",
			buildArgs: []string{"testdir"},
			setupFs: func(t testing.TB, fs afero.Fs) {
				create(t, fs, "testdir/Containerfile", "FROM alpine")
				create(t, fs, "testdir/Dockerfile", "FROM ubuntu")
			},
			expectedPath:   filepath.Join("testdir", "Containerfile"),
			expectedErrMsg: "",
		},
		{
			name:           "non-existent directory",
			buildArgs:      []string{"nonexistentdir"},
			setupFs:        func(t testing.TB, fs afero.Fs) {},
			expectedPath:   "",
			expectedErrMsg: "could not find Containerfile or Dockerfile",
		},
		{
			name:           "empty build args",
			buildArgs:      []string{},
			setupFs:        func(t testing.TB, fs afero.Fs) {},
			expectedPath:   "",
			expectedErrMsg: "could not find Containerfile or Dockerfile",
		},
		{
			name:      "directory exists but no container files",
			buildArgs: []string{"testdir"},
			setupFs: func(t testing.TB, fs afero.Fs) {
				create(t, fs, "testdir/somefile.txt", "content")
			},
			expectedPath:   "",
			expectedErrMsg: "could not find Containerfile or Dockerfile",
		},
		{
			name:      "find in second directory",
			buildArgs: []string{"firstdir", "seconddir"},
			setupFs: func(t testing.TB, fs afero.Fs) {
				err := fs.MkdirAll("firstdir", 0755)
				require.NoError(t, err, "Failed to create directory: firstdir")
				create(t, fs, "seconddir/Dockerfile", "FROM alpine")
			},
			expectedPath:   filepath.Join("seconddir", "Dockerfile"),
			expectedErrMsg: "",
		},
		{
			name:      "find in first directory when both have files",
			buildArgs: []string{"firstdir", "seconddir"},
			setupFs: func(t testing.TB, fs afero.Fs) {
				create(t, fs, "firstdir/Containerfile", "FROM alpine")
				create(t, fs, "seconddir/Dockerfile", "FROM ubuntu")
			},
			expectedPath:   filepath.Join("firstdir", "Containerfile"),
			expectedErrMsg: "",
		},
		{
			name:      "file argument not a directory",
			buildArgs: []string{"testfile.txt"},
			setupFs: func(t testing.TB, fs afero.Fs) {
				create(t, fs, "testfile.txt", "content")
			},
			expectedPath:   "",
			expectedErrMsg: "could not find Containerfile or Dockerfile",
		},
		{
			name:      "mixed args with valid directory",
			buildArgs: []string{"testfile.txt", "testdir"},
			setupFs: func(t testing.TB, fs afero.Fs) {
				create(t, fs, "testfile.txt", "content")
				create(t, fs, "testdir/Dockerfile", "FROM alpine")
			},
			expectedPath:   filepath.Join("testdir", "Dockerfile"),
			expectedErrMsg: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			tt.setupFs(t, fs)

			result, err := tryFindContainerfile(fs, tt.buildArgs)

			if tt.expectedErrMsg != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedPath, result)
			}
		})
	}
}

func create(t testing.TB, fs afero.Fs, path, contents string) {
	t.Helper()

	dir := filepath.Dir(path)
	if dir != "." {
		err := fs.MkdirAll(dir, 0755)
		require.NoError(t, err, "Failed to create directory: %s", dir)
	}

	err := afero.WriteFile(fs, path, []byte(contents), 0644)
	require.NoError(t, err, "Failed to write file: %s", path)
}
