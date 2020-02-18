package cmd

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"gerrit.instructure.com/muss/config"
)

func testRootCmd(args ...string) (int, string, string) {
	var stdout, stderr strings.Builder

	cfg, _ := config.NewConfigFromMap(nil)
	rootCmd := NewRootCommand(cfg)

	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)

	exitCode := ExecuteRoot(rootCmd, args)

	return exitCode, stdout.String(), stderr.String()
}

func getLines(s string, want int) []string {
	lines := strings.SplitAfter(s, "\n")
	end := len(lines)
	if want < end {
		end = want
	}
	return lines[0:end]
}

func TestRootCommand(t *testing.T) {
	withTestPath(t, func(*testing.T) {
		t.Run("bad flag", func(*testing.T) {
			exitCode, stdout, stderr := testRootCmd("--foo")

			assert.Equal(t, 1, exitCode, "exit 1")
			assert.Equal(t, "", stdout)
			assert.Equal(t,
				[]string{"Error:  unknown flag: --foo\n", "\n", "Usage:\n", "  muss [command]\n"},
				getLines(stderr, 4),
			)
		})

		t.Run("bad subcmd flag", func(*testing.T) {
			exitCode, stdout, stderr := testRootCmd("wrap", "--foo")

			assert.Equal(t, 1, exitCode, "exit 1")
			assert.Equal(t, "", stdout)
			assert.Equal(t,
				[]string{"Error:  unknown flag: --foo\n", "\n", "Usage:\n", "  muss wrap [flags]\n"},
				getLines(stderr, 4),
			)
		})

		t.Run("non-zero delegated command exit", func(*testing.T) {
			os.Setenv("MUSS_TEST_DC_ERROR", "2")
			defer os.Unsetenv("MUSS_TEST_DC_ERROR")
			exitCode, stdout, stderr := testRootCmd("pull")

			assert.Equal(t, 2, exitCode, "exit 2")
			assert.Equal(t, "", stdout)
			assert.Equal(t, "", stderr)
		})
	})
}
