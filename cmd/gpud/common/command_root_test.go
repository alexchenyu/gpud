package common

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	// Alias to avoid conflict if common.log is used
)

func TestCommand(t *testing.T) {
	cmd := RootCommand()
	assert.Equal(t, "gpud", cmd.Use)
	assert.Equal(t, "GPUd tool", cmd.Short)
	assert.True(t, strings.Contains(cmd.Example, "gpud scan"))
	assert.True(t, strings.Contains(cmd.Example, "sudo gpud up"))
}

func TestAddCommand(t *testing.T) {
	// Save the original commands to restore after test
	originalCommands := cmdRoot.Commands()
	defer func() {
		// Reset the cmdRoot commands
		cmdRoot.ResetCommands()
		for _, cmd := range originalCommands {
			cmdRoot.AddCommand(cmd)
		}
	}()

	// Create a test subcommand
	testCmd := &cobra.Command{
		Use:   "test",
		Short: "Test command",
	}

	// Get initial number of subcommands
	initialCmdCount := len(cmdRoot.Commands())

	// Add the test command
	cmdRoot.AddCommand(testCmd)

	// Verify the command was added
	assert.Equal(t, initialCmdCount+1, len(cmdRoot.Commands()))

	// Find the added command
	var found bool
	for _, cmd := range cmdRoot.Commands() {
		if cmd.Use == "test" {
			found = true
			break
		}
	}
	assert.True(t, found, "Added command was not found")
}

func TestFlags(t *testing.T) {
	cmd := RootCommand()

	// Test log-level flag
	logLevelFlag := cmd.PersistentFlags().Lookup("log-level")
	assert.NotNil(t, logLevelFlag)
	assert.Equal(t, "info", logLevelFlag.DefValue)

	// Test log-file flag
	logFileFlag := cmd.PersistentFlags().Lookup("log-file")
	assert.NotNil(t, logFileFlag)
	assert.Equal(t, "", logFileFlag.DefValue)

	// Test ibstat-command flag
	ibstatFlag := cmd.PersistentFlags().Lookup("ibstat-command")
	assert.NotNil(t, ibstatFlag)
	assert.Equal(t, "", ibstatFlag.DefValue)

	// Test ibstatus-command flag
	ibstatusFlag := cmd.PersistentFlags().Lookup("ibstatus-command")
	assert.NotNil(t, ibstatusFlag)
	assert.Equal(t, "", ibstatusFlag.DefValue)
}

func TestFlagLogLevel(t *testing.T) {
	t.Run("default value", func(t *testing.T) {
		// Create a test command with log-level flag
		cmd := &cobra.Command{}
		// Important: Use cmd.Flags() directly since FlagLogLevel uses cmd.Flags().GetString()
		cmd.Flags().StringP("log-level", "l", "info", "help")

		level, err := FlagLogLevel(cmd)
		assert.NoError(t, err)
		assert.Equal(t, zapcore.InfoLevel, level.Level())
	})

	testCases := []struct {
		name          string
		logLevelArg   string
		expectedLevel zapcore.Level
	}{
		{"debug", "debug", zapcore.DebugLevel},
		{"info", "info", zapcore.InfoLevel},
		{"warn", "warn", zapcore.WarnLevel},
		{"error", "error", zapcore.ErrorLevel},
		{"dpanic", "dpanic", zapcore.DPanicLevel},
		{"panic", "panic", zapcore.PanicLevel},
		{"fatal", "fatal", zapcore.FatalLevel},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			cmd.Flags().StringP("log-level", "l", "info", "help")

			require.NoError(t, cmd.Flags().Set("log-level", tc.logLevelArg))

			level, err := FlagLogLevel(cmd)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedLevel, level.Level())
		})
	}

	t.Run("invalid level string", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.Flags().StringP("log-level", "l", "info", "help")

		require.NoError(t, cmd.Flags().Set("log-level", "invalid"))

		level, err := FlagLogLevel(cmd)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), `unrecognized level: "invalid"`)

		// Don't call level.Level() when there's an error, since ParseLogLevel
		// returns an empty zap.AtomicLevel{} which would cause a nil pointer panic
		if err == nil {
			assert.Equal(t, zapcore.InfoLevel, level.Level(), "Expected InfoLevel as default")
		}
	})
}

func TestFlagLogLevelGetStringError(t *testing.T) {
	cmd := &cobra.Command{Use: "test"} // No flags defined

	level, err := FlagLogLevel(cmd)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "flag accessed but not defined: log-level")
	// FlagLogLevel returns (zap.NewAtomicLevelAt(zap.InfoLevel), err) if GetString fails.
	assert.Equal(t, zapcore.InfoLevel, level.Level(), "Expected InfoLevel on GetString error")
}

func TestFlagLogFile(t *testing.T) {
	t.Run("default value", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.Flags().String("log-file", "", "help")

		file, err := FlagLogFile(cmd)
		assert.NoError(t, err)
		assert.Equal(t, "", file)
	})

	t.Run("changed value", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.Flags().String("log-file", "", "help")

		expectedPath := "/tmp/logfile.log"
		require.NoError(t, cmd.Flags().Set("log-file", expectedPath))

		file, err := FlagLogFile(cmd)
		assert.NoError(t, err)
		assert.Equal(t, expectedPath, file)
	})
}

func TestFlagLogFileGetStringError(t *testing.T) {
	cmd := &cobra.Command{Use: "test"} // No flags defined

	_, err := FlagLogFile(cmd)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "flag accessed but not defined: log-file")
}

func TestFlagIbstatCommand(t *testing.T) {
	t.Run("default value", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.Flags().String("ibstat-command", "", "help")

		command, err := FlagIbstatCommand(cmd)
		assert.NoError(t, err)
		assert.Equal(t, "", command)
	})

	t.Run("changed value", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.Flags().String("ibstat-command", "", "help")

		expectedCmd := "/custom/ibstat"
		require.NoError(t, cmd.Flags().Set("ibstat-command", expectedCmd))

		command, err := FlagIbstatCommand(cmd)
		assert.NoError(t, err)
		assert.Equal(t, expectedCmd, command)
	})
}

func TestFlagIbstatusCommand(t *testing.T) {
	t.Run("default value", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.Flags().String("ibstatus-command", "", "help")

		command, err := FlagIbstatusCommand(cmd)
		assert.NoError(t, err)
		assert.Equal(t, "", command)
	})

	t.Run("changed value", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.Flags().String("ibstatus-command", "", "help")

		expectedCmd := "/custom/ibstatus"
		require.NoError(t, cmd.Flags().Set("ibstatus-command", expectedCmd))

		command, err := FlagIbstatusCommand(cmd)
		assert.NoError(t, err)
		assert.Equal(t, expectedCmd, command)
	})
}

// setupTestCommandForCreateLogger creates a command for use in TestCreateLoggerFromFlags
func setupTestCommandForCreateLogger() *cobra.Command {
	cmd := &cobra.Command{}
	// Define flags on both flagsets to ensure they're accessible from FlagLogLevel and FlagLogFile
	cmd.PersistentFlags().StringP("log-level", "l", "info", "set the logging level")
	cmd.PersistentFlags().String("log-file", "", "set the log file path")

	// Also add them to cmd.Flags() because FlagLogLevel/FlagLogFile use cmd.Flags().GetString
	cmd.Flags().AddFlag(cmd.PersistentFlags().Lookup("log-level"))
	cmd.Flags().AddFlag(cmd.PersistentFlags().Lookup("log-file"))

	return cmd
}

func TestCreateLoggerFromFlags(t *testing.T) {
	t.Run("success with debug level and log file", func(t *testing.T) {
		tempDir := t.TempDir()
		logFilePath := filepath.Join(tempDir, "test.log")

		cmd := setupTestCommandForCreateLogger()
		// Set values using PersistentFlags since that's where we defined them
		require.NoError(t, cmd.PersistentFlags().Set("log-level", "debug"))
		require.NoError(t, cmd.PersistentFlags().Set("log-file", logFilePath))

		logger, logLevel, err := CreateLoggerFromFlags(cmd)
		require.NoError(t, err)
		require.NotNil(t, logger)
		assert.Equal(t, zapcore.DebugLevel, logLevel.Level())

		assert.True(t, logger.Desugar().Core().Enabled(zap.DebugLevel), "Debug level should be enabled")
		assert.True(t, logger.Desugar().Core().Enabled(zap.InfoLevel), "Info level should be enabled when debug is on")
	})

	t.Run("success with default info level and no log file", func(t *testing.T) {
		cmd := setupTestCommandForCreateLogger()

		logger, logLevel, err := CreateLoggerFromFlags(cmd)
		require.NoError(t, err)
		require.NotNil(t, logger)
		assert.Equal(t, zapcore.InfoLevel, logLevel.Level())

		assert.True(t, logger.Desugar().Core().Enabled(zap.InfoLevel), "Info level should be enabled")
		assert.False(t, logger.Desugar().Core().Enabled(zap.DebugLevel), "Debug level should NOT be enabled for info level")
	})

	t.Run("error with invalid log level", func(t *testing.T) {
		cmd := setupTestCommandForCreateLogger()
		require.NoError(t, cmd.PersistentFlags().Set("log-level", "invalid-level"))

		logger, logLevel, err := CreateLoggerFromFlags(cmd)
		require.Error(t, err)
		assert.Nil(t, logger)
		assert.Equal(t, zapcore.InfoLevel, logLevel.Level(), "Should return InfoLevel on error")
		assert.Contains(t, err.Error(), `unrecognized level: "invalid-level"`)
	})

	t.Run("success with warn level and no log file", func(t *testing.T) {
		cmd := setupTestCommandForCreateLogger()
		require.NoError(t, cmd.PersistentFlags().Set("log-level", "warn"))

		logger, logLevel, err := CreateLoggerFromFlags(cmd)
		require.NoError(t, err)
		require.NotNil(t, logger)
		assert.Equal(t, zapcore.WarnLevel, logLevel.Level())

		assert.True(t, logger.Desugar().Core().Enabled(zap.WarnLevel), "Warn level should be enabled")
		assert.False(t, logger.Desugar().Core().Enabled(zap.InfoLevel), "Info level should NOT be enabled for warn level")
		assert.False(t, logger.Desugar().Core().Enabled(zap.DebugLevel), "Debug level should NOT be enabled for warn level")
	})

	t.Run("error when log-level flag is missing from command", func(t *testing.T) {
		cmdMissingFlags := &cobra.Command{}
		// Add log-file to both PersistentFlags and Flags
		cmdMissingFlags.PersistentFlags().String("log-file", "", "set the log file path")
		cmdMissingFlags.Flags().AddFlag(cmdMissingFlags.PersistentFlags().Lookup("log-file"))

		logger, logLevel, err := CreateLoggerFromFlags(cmdMissingFlags)
		require.Error(t, err, "Expected error when log-level flag is missing")
		assert.Nil(t, logger)
		assert.Equal(t, zapcore.InfoLevel, logLevel.Level())
		assert.Contains(t, err.Error(), "flag accessed but not defined: log-level")
	})

	t.Run("error when log-file flag is missing from command", func(t *testing.T) {
		cmdMissingFlags := &cobra.Command{}
		// Add log-level to both PersistentFlags and Flags
		cmdMissingFlags.PersistentFlags().StringP("log-level", "l", "info", "set the logging level")
		cmdMissingFlags.Flags().AddFlag(cmdMissingFlags.PersistentFlags().Lookup("log-level"))

		logger, logLevel, err := CreateLoggerFromFlags(cmdMissingFlags)
		require.Error(t, err, "Expected error when log-file flag is missing")
		assert.Nil(t, logger)
		assert.Equal(t, zapcore.InfoLevel, logLevel.Level())
		assert.Contains(t, err.Error(), "flag accessed but not defined: log-file")
	})
}
