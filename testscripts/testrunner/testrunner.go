package testrunner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/rogpeppe/go-internal/testscript"
	"github.com/stretchr/testify/require"
	"go.jetpack.io/devbox/internal/boxcli"
)

func Main(m *testing.M) int {
	commands := map[string]func() int{
		"devbox": func() int {
			// Call the devbox CLI directly:
			return boxcli.Execute(context.Background(), os.Args[1:])
		},
		"print": func() int { // Not 'echo' because we don't expand variables
			fmt.Println(strings.Join(os.Args[1:], " "))
			return 0
		},
	}
	return testscript.RunMain(m, commands)
}

func RunTestscripts(t *testing.T, testscriptsDir string) {
	globPattern := filepath.Join(testscriptsDir, "**/*.test.txt")
	dirs := globDirs(globPattern)
	require.NotEmpty(t, dirs, "no test scripts found")

	// Loop through all the directories and run all tests scripts (files ending
	// in .test.txt)
	for _, dir := range dirs {
		// The testrunner dir has the testscript we use for projects in examples/ directory.
		// We should skip that one since it is run separately (see RunExamplesTestscripts).
		if filepath.Base(dir) == "testrunner" {
			continue
		}

		testscript.Run(t, getTestscriptParams(dir))
	}
}

// Return directories that contain files matching the pattern.
func globDirs(pattern string) []string {
	scripts, err := doublestar.FilepathGlob(pattern)
	if err != nil {
		return nil
	}

	// List of directories with test scripts.
	directories := []string{}
	dups := map[string]bool{}
	for _, script := range scripts {
		dir := filepath.Dir(script)
		if _, ok := dups[dir]; !ok {
			directories = append(directories, dir)
			dups[dir] = true
		}
	}

	return directories
}

func getTestscriptParams(dir string) testscript.Params {
	return testscript.Params{
		Dir:                 dir,
		RequireExplicitExec: true,
		TestWork:            false, // Set to true if you're trying to debug a test.
		Setup:               setupTestEnv,
		Cmds: map[string]func(ts *testscript.TestScript, neg bool, args []string){
			"env.path.len":  assertPathLength,
			"json.superset": assertJSONSuperset,
			"path.order":    assertPathOrder,
			"source.path":   sourcePath,
		},
	}
}
