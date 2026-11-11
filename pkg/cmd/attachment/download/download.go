// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package download

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

type downloadedFile struct {
	ID       string `json:"id"`
	FileName string `json:"fileName"`
	Path     string `json:"path"`
	Size     int64  `json:"size"`
}

func NewCmdDownload(f *cmdutil.Factory) *cobra.Command {
	var flagOutput *string

	cmd := &cobra.Command{
		Use:   "download <id> [path]",
		Short: "Download a private file",
		Example: `  # Download a file into the current directory
  prb file download <id>

  # Choose the output path
  prb file download <id> ./evidence.pdf

  # Write the bytes to stdout
  prb file download <id> -`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.ValidateOutputFlag(flagOutput); err != nil {
				return fmt.Errorf("cannot validate output: %w", err)
			}

			dest := ""
			if len(args) == 2 {
				dest = args[1]
			}

			if dest == "-" && *flagOutput == cmdutil.OutputJSON {
				return fmt.Errorf("cannot download file: --json writes metadata and cannot share stdout with the file")
			}

			fileID, err := api.ParseFileID(args[0])
			if err != nil {
				return err
			}

			cfg, err := f.Config()
			if err != nil {
				return fmt.Errorf("cannot load config: %w", err)
			}

			host, hc, err := cfg.DefaultHost()
			if err != nil {
				return fmt.Errorf("cannot load default host: %w", err)
			}

			client := api.NewClient(
				host,
				hc.Token,
				cmdutil.ConsoleEndpoint,
				cfg.HTTPTimeoutDuration(),
				cmdutil.TokenRefreshOption(cfg, host, hc),
			)

			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}

			body, err := client.DownloadFile(ctx, fileID)
			if err != nil {
				return err
			}

			defer func() { _ = body.Body.Close() }()

			path, size, err := saveDownload(dest, body.FileName, body.Body, f.IOStreams.Out)
			if err != nil {
				return err
			}

			downloaded := downloadedFile{
				ID:       fileID,
				FileName: body.FileName,
				Path:     path,
				Size:     size,
			}

			if *flagOutput == cmdutil.OutputJSON {
				if err := cmdutil.PrintJSON(f.IOStreams.Out, downloaded); err != nil {
					return fmt.Errorf("cannot print file: %w", err)
				}

				return nil
			}

			if dest == "-" {
				_, _ = fmt.Fprintf(f.IOStreams.ErrOut, "Downloaded %s\n", body.FileName)

				return nil
			}

			_, _ = fmt.Fprintf(f.IOStreams.Out, "Downloaded %s\n", path)

			return nil
		},
	}

	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}

func saveDownload(dest, name string, body io.Reader, stdout io.Writer) (string, int64, error) {
	if dest == "-" {
		n, err := io.Copy(stdout, body)
		if err != nil {
			return "", 0, fmt.Errorf("cannot download file: %w", err)
		}

		return "-", n, nil
	}

	path := name

	if dest != "" {
		info, err := os.Stat(dest)
		switch {
		case err == nil && info.IsDir():
			path = filepath.Join(dest, name)
		case err == nil || os.IsNotExist(err):
			path = dest
		default:
			return "", 0, fmt.Errorf("cannot download file: %w", err)
		}
	}

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return "", 0, fmt.Errorf("cannot download file: %s already exists", path)
		}

		return "", 0, fmt.Errorf("cannot download file: %w", err)
	}

	remove := true

	defer func() {
		if remove {
			_ = os.Remove(path)
		}
	}()

	n, err := io.Copy(file, body)
	if err != nil {
		_ = file.Close()

		return "", 0, fmt.Errorf("cannot download file: %w", err)
	}

	if err := file.Close(); err != nil {
		return "", 0, fmt.Errorf("cannot download file: %w", err)
	}

	remove = false

	return path, n, nil
}
