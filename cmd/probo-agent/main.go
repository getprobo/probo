// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/deviceagent"
	"go.probo.inc/probo/pkg/deviceagent/service"
	"go.probo.inc/probo/pkg/deviceagent/update"

	// Side-effect import: registers per-OS posture checks.
	_ "go.probo.inc/probo/pkg/deviceagent/checks"
)

var version = "dev"

// restartExitCode signals the OS service supervisor that the agent
// process exited because its binary was replaced and needs to be
// restarted. The value matches sysexits.h's EX_TEMPFAIL and is
// whitelisted in the systemd unit so the unit does not enter the
// "failed" state on a normal self-update.
const restartExitCode = 75

func main() {
	// Best-effort cleanup of a previous-version binary left aside by
	// a Windows self-update. No-op on Unix.
	if exe, err := os.Executable(); err == nil {
		update.CleanupAfterRestart(exe)
	}

	if err := newRootCmd().Execute(); err != nil {
		if errors.Is(err, deviceagent.ErrRestartRequired) {
			os.Exit(restartExitCode)
		}

		fmt.Fprintf(os.Stderr, "probo-agent: %s\n", err)
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "probo-agent",
		Short:         "Probo device posture agent",
		Long:          "probo-agent runs as a managed OS service, reporting device posture to Probo.",
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       version,
	}

	root.PersistentFlags().StringP("dir", "d", "", "agent config / keystore directory (defaults to platform-specific path)")

	root.AddCommand(newInstallCmd())
	root.AddCommand(newUninstallCmd())
	root.AddCommand(newRunCmd())
	root.AddCommand(newStatusCmd())
	root.AddCommand(newCollectCmd())
	root.AddCommand(newUpdateCmd())

	return root
}

// newUpdater returns an Updater scoped to the running binary, or nil
// when self-update cannot be performed (unresolvable binary path).
//
// dir is the agent state directory, used to host the Sigstore TUF
// metadata cache for cosign bundle verification.
func newUpdater(logger *log.Logger, dir string) *update.Updater {
	exePath, err := os.Executable()
	if err != nil || exePath == "" {
		return nil
	}

	return update.New(
		version,
		exePath,
		fmt.Sprintf("probo-agent/%s", version),
		filepath.Join(dir, "sigstore-cache"),
		logger,
	)
}

func resolveDir(cmd *cobra.Command) string {
	dir, _ := cmd.Flags().GetString("dir")
	if dir != "" {
		return dir
	}

	return deviceagent.DefaultConfigDir()
}

func newAgentLogger() *log.Logger {
	return log.NewLogger(
		log.WithName("probo-agent"),
		log.WithOutput(os.Stderr),
	)
}

func newInstallCmd() *cobra.Command {
	var (
		serverURL       string
		enrollmentToken string
		skipService     bool
		noAutoUpdate    bool
	)

	cmd := &cobra.Command{
		Use:   "install",
		Short: "Enroll this device and install the agent as a managed OS service",
		RunE: func(cmd *cobra.Command, args []string) error {
			if serverURL == "" {
				return errors.New("--server is required")
			}

			if enrollmentToken == "" {
				if v := os.Getenv("PROBO_TOKEN"); v != "" {
					enrollmentToken = v
				}
			}

			if enrollmentToken == "" {
				return errors.New("--enrollment-token (or PROBO_TOKEN env var) is required")
			}

			dir := resolveDir(cmd)

			ctx, cancel := context.WithTimeout(cmd.Context(), 60*time.Second)
			defer cancel()

			agent := deviceagent.New(dir, version, newAgentLogger())

			resp, err := agent.EnrollNewDevice(ctx, strings.TrimRight(serverURL, "/"), enrollmentToken)
			if err != nil {
				return fmt.Errorf("enrollment failed: %w", err)
			}

			fmt.Printf("Enrolled device %s (heartbeat %ds, posture %ds)\n",
				resp.DeviceID, resp.HeartbeatSeconds, resp.PostureSeconds)

			if noAutoUpdate {
				if err := persistAutoUpdate(dir, false); err != nil {
					return fmt.Errorf("cannot persist auto-update preference: %w", err)
				}

				fmt.Println("Auto-update disabled.")
			}

			if skipService {
				fmt.Println("Service installation skipped (--skip-service).")
				return nil
			}

			exePath, err := os.Executable()
			if err != nil {
				return fmt.Errorf("cannot resolve current executable path: %w", err)
			}

			if err := service.Install(
				service.Config{
					ExePath: exePath,
					Dir:     dir,
				},
			); err != nil {
				return fmt.Errorf("cannot install OS service: %w", err)
			}

			fmt.Println("Service installed and started.")

			return nil
		},
	}

	cmd.Flags().StringVar(&serverURL, "server", "", "Probo server base URL (e.g. https://app.getprobo.com)")
	cmd.Flags().StringVar(&enrollmentToken, "enrollment-token", "", "device enrollment token issued by an admin")
	cmd.Flags().BoolVar(&skipService, "skip-service", false, "register the device but do not install the OS service")
	cmd.Flags().BoolVar(&noAutoUpdate, "no-auto-update", false, "disable automatic upgrades of the agent binary")

	return cmd
}

// persistAutoUpdate flips the UpdatesDisabled flag in the agent's
// on-disk config without disturbing other fields.
func persistAutoUpdate(dir string, enabled bool) error {
	cfg, err := deviceagent.LoadConfig(dir)
	if err != nil {
		return err
	}

	cfg.UpdatesDisabled = !enabled

	return deviceagent.SaveConfig(dir, cfg)
}

func newUninstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "uninstall",
		Short: "Stop the service, unenroll this device, and remove local state",
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := resolveDir(cmd)

			ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
			defer cancel()

			agent := deviceagent.New(dir, version, newAgentLogger())
			if err := agent.Unenroll(ctx); err != nil {
				fmt.Fprintf(os.Stderr, "warning: unenroll failed: %v\n", err)
			}

			if err := service.Uninstall(service.Config{Dir: dir}); err != nil {
				fmt.Fprintf(os.Stderr, "warning: service uninstall failed: %v\n", err)
			}

			_ = os.Remove(deviceagent.ConfigPath(dir))

			fmt.Println("Uninstalled.")

			return nil
		},
	}
}

func newRunCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "run",
		Short: "Run the agent in the foreground (used by the OS service unit)",
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := resolveDir(cmd)

			ctx, stop := signal.NotifyContext(cmd.Context(), syscall.SIGINT, syscall.SIGTERM)
			defer stop()

			logger := newAgentLogger()
			agent := deviceagent.New(dir, version, logger)
			agent.Updater = newUpdater(logger, dir)

			err := agent.Run(ctx)
			if errors.Is(err, context.Canceled) {
				return nil
			}

			return err
		},
	}
}

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Print the agent's local state",
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := resolveDir(cmd)

			cfg, err := deviceagent.LoadConfig(dir)
			if err != nil {
				return err
			}

			haveKey := true
			if _, err := deviceagent.LoadAPIKey(dir); err != nil {
				haveKey = false
			}

			fmt.Printf("Server URL:           %s\n", cfg.ServerURL)
			fmt.Printf("Device ID:            %s\n", cfg.DeviceID)
			fmt.Printf("Heartbeat interval:   %s\n", cfg.HeartbeatInterval)
			fmt.Printf("Posture interval:     %s\n", cfg.PostureInterval)
			fmt.Printf("Update interval:      %s\n", cfg.UpdateInterval)
			fmt.Printf("Auto-update enabled:  %v\n", !cfg.UpdatesDisabled)
			fmt.Printf("API key on disk:      %v\n", haveKey)
			fmt.Printf("Config directory:     %s\n", dir)

			return nil
		},
	}
}

func newCollectCmd() *cobra.Command {
	var (
		once     bool
		asJSON   bool
		printDir bool
	)

	cmd := &cobra.Command{
		Use:   "collect",
		Short: "Run the posture check set once and print results (no server push)",
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := resolveDir(cmd)
			if printDir {
				fmt.Println(dir)
			}

			ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
			defer cancel()

			agent := deviceagent.New(dir, version, newAgentLogger())
			results := agent.CollectOnce(ctx)

			if asJSON {
				return json.NewEncoder(os.Stdout).Encode(results)
			}

			for _, r := range results {
				fmt.Printf("%-20s %-15s %v\n", r.CheckKey, r.Status, r.Evidence)
			}

			return nil
		},
	}
	cmd.Flags().BoolVar(&once, "once", true, "(default true) run the check set once and exit")
	cmd.Flags().BoolVar(&asJSON, "json", false, "emit JSON instead of the human-readable table")
	cmd.Flags().BoolVar(&printDir, "print-dir", false, "print the resolved agent dir before the results")

	return cmd
}

func newUpdateCmd() *cobra.Command {
	var checkOnly bool

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Check GitHub for a newer agent release and install it in place",
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := resolveDir(cmd)
			logger := newAgentLogger()

			updater := newUpdater(logger, dir)
			if updater == nil {
				return errors.New("cannot resolve current executable path")
			}

			ctx, cancel := context.WithTimeout(cmd.Context(), 10*time.Minute)
			defer cancel()

			rel, err := updater.CheckLatest(ctx)
			if err != nil {
				if errors.Is(err, update.ErrNoUpdateAvailable) {
					fmt.Printf("probo-agent is up to date (version %s).\n", version)
					return nil
				}

				return fmt.Errorf("cannot check for updates: %w", err)
			}

			fmt.Printf("Update available: %s -> %s\n", version, rel.Version)

			if checkOnly {
				return nil
			}

			if err := updater.Apply(ctx, rel); err != nil {
				return fmt.Errorf("cannot apply update: %w", err)
			}

			fmt.Printf("Installed probo-agent %s. Restart the service to use it.\n", rel.Version)

			return nil
		},
	}

	cmd.Flags().BoolVar(&checkOnly, "check", false, "only print the available version, do not install it")

	return cmd
}
