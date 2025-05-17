// Package run implements the "run" command.
package run

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/leptonai/gpud/cmd/gpud/common"
	"github.com/leptonai/gpud/pkg/config"
	gpud_manager "github.com/leptonai/gpud/pkg/gpud-manager"
	"github.com/leptonai/gpud/pkg/log"
	gpudserver "github.com/leptonai/gpud/pkg/server"
	pkgsystemd "github.com/leptonai/gpud/pkg/systemd"
	"github.com/leptonai/gpud/version"
)

// Command returns the cobra command for the "run" command.
func Command() *cobra.Command {
	return cmdRoot
}

var cmdRoot = &cobra.Command{
	Use:   "run",
	Short: "starts gpud without any login/checkin ('gpud up' is recommended for linux)",
	RunE:  cmdRootFunc,
}

var (
	flagAnnotations        string
	flagListenAddr         string
	flagPprof              bool
	flagRetentionPeriod    time.Duration
	flagEndpoint           string
	flagEnableAutoUpdate   bool
	flagAutoUpdateExitCode int
	flagPluginSpecsFile    string
	flagEnablePluginAPI    bool
)

func init() {
	cmdRoot.PersistentFlags().StringVar(&flagAnnotations, "annotations", "", "set the annotations in JSON map")
	cmdRoot.PersistentFlags().StringVar(&flagListenAddr, "listen-address", fmt.Sprintf("0.0.0.0:%d", config.DefaultGPUdPort), "set the listen address")
	cmdRoot.PersistentFlags().BoolVar(&flagPprof, "pprof", false, "enable pprof (default: false)")
	cmdRoot.PersistentFlags().DurationVar(&flagRetentionPeriod, "retention-period", config.DefaultRetentionPeriod.Duration, "set the time period to retain metrics for (once elapsed, old records are compacted/purged)")
	cmdRoot.PersistentFlags().StringVar(&flagEndpoint, "endpoint", "mothership-machine.app.lepton.ai", "set the endpoint for control plane")
	cmdRoot.PersistentFlags().BoolVar(&flagEnableAutoUpdate, "enable-auto-update", true, "enable auto update of gpud (default: true)")
	cmdRoot.PersistentFlags().IntVar(&flagAutoUpdateExitCode, "auto-update-exit-code", -1, "specifies the exit code to exit with when auto updating (default: -1 to disable exit code)")
	cmdRoot.PersistentFlags().StringVar(&flagPluginSpecsFile, "plugin-specs-file", "", "sets the plugin specs file (leave empty for default, useful for testing)")
	cmdRoot.PersistentFlags().BoolVar(&flagEnablePluginAPI, "enable-plugin-api", false, "enable plugin API (default: false)")
}

func cmdRootFunc(cmd *cobra.Command, args []string) error {
	logger, logLevel, err := common.CreateLoggerFromFlags(cmd)
	if err != nil {
		return err
	}
	log.Logger = logger

	ibstatCommand, err := common.FlagIbstatCommand(cmd)
	if err != nil {
		return err
	}
	ibstatusCommand, err := common.FlagIbstatusCommand(cmd)
	if err != nil {
		return err
	}

	log.Logger.Debugw("starting run command")

	if logLevel.Level() > zap.DebugLevel { // e.g., info, warn, error
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	configOpts := []config.OpOption{
		config.WithIbstatCommand(ibstatCommand),
		config.WithIbstatusCommand(ibstatusCommand),
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	cfg, err := config.DefaultConfig(ctx, configOpts...)
	cancel()
	if err != nil {
		return err
	}

	if flagAnnotations != "" {
		annot := make(map[string]string)
		err = json.Unmarshal([]byte(flagAnnotations), &annot)
		if err != nil {
			return err
		}
		cfg.Annotations = annot
	}

	if flagListenAddr != "" {
		cfg.Address = flagListenAddr
	}

	cfg.Pprof = flagPprof

	if flagRetentionPeriod > 0 {
		cfg.RetentionPeriod = metav1.Duration{Duration: flagRetentionPeriod}
	}

	cfg.CompactPeriod = config.DefaultCompactPeriod

	cfg.EnableAutoUpdate = flagEnableAutoUpdate
	cfg.AutoUpdateExitCode = flagAutoUpdateExitCode

	cfg.PluginSpecsFile = flagPluginSpecsFile
	cfg.EnablePluginAPI = flagEnablePluginAPI

	if err := cfg.Validate(); err != nil {
		return err
	}

	rootCtx, rootCancel := context.WithCancel(context.Background())
	defer rootCancel()

	start := time.Now()

	signals := make(chan os.Signal, 2048)
	serverC := make(chan gpudserver.ServerStopper, 1)

	log.Logger.Infof("starting gpud %v", version.Version)

	done := gpudserver.HandleSignals(rootCtx, rootCancel, signals, serverC, func(ctx context.Context) error {
		if pkgsystemd.SystemctlExists() {
			if err := pkgsystemd.NotifyStopping(ctx); err != nil {
				log.Logger.Errorw("notify stopping failed")
			}
		}
		return nil
	})

	// start the signal handler as soon as we can to make sure that
	// we don't miss any signals during boot
	signal.Notify(signals, gpudserver.DefaultSignalsToHandle...)
	m, err := gpud_manager.New()
	if err != nil {
		return err
	}
	m.Start(rootCtx)

	server, err := gpudserver.New(rootCtx, cfg, m)
	if err != nil {
		return err
	}
	serverC <- server

	if pkgsystemd.SystemctlExists() {
		if err := pkgsystemd.NotifyReady(rootCtx); err != nil {
			log.Logger.Warnw("notify ready failed")
		}
	} else {
		log.Logger.Debugw("skipped sd notify as systemd is not available")
	}

	log.Logger.Infow("successfully booted", "tookSeconds", time.Since(start).Seconds())
	<-done

	return nil
}
