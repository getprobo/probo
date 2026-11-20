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

package repair

import (
	"fmt"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/spf13/cobra"
	"go.gearno.de/kit/httpclient"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/awsconfig"
	"go.probo.inc/probo/pkg/esign"
	"go.probo.inc/probo/pkg/filemanager"
	"go.probo.inc/probo/pkg/html2pdf"
	"go.probo.inc/probo/pkg/proboctl/cmdutil"
)

func newCmdSignatureBundles(f *cmdutil.Factory) *cobra.Command {
	var (
		flagBucket          string
		flagRegion          string
		flagAccessKeyID     string
		flagSecretAccessKey string
		flagEndpoint        string
		flagUsePathStyle    bool
		flagDryRun          bool
		flagYes             bool
		flagContinueOnError bool
	)

	cmd := &cobra.Command{
		Use:   "signature-bundles",
		Short: "Store zips for signatures that already have a certificate",
		Long: "Read each completed signature that has a certificate and no zip, " +
			"then store the signed file and certificate as one private file. " +
			"Run it after deploying the bundle column. Re-running skips " +
			"signatures that already have a zip. AWS settings default to the " +
			"PROBOD_AWS_* environment variables.",
		Args: cobra.NoArgs,
	}

	cmd.Flags().StringVar(&flagBucket, "bucket", "", "File bucket (default: PROBOD_AWS_BUCKET)")
	cmd.Flags().StringVar(&flagRegion, "aws-region", "", "AWS region (default: PROBOD_AWS_REGION)")
	cmd.Flags().StringVar(&flagAccessKeyID, "aws-access-key-id", "", "AWS access key ID (default: PROBOD_AWS_ACCESS_KEY_ID)")
	cmd.Flags().StringVar(&flagSecretAccessKey, "aws-secret-access-key", "", "AWS secret access key (default: PROBOD_AWS_SECRET_ACCESS_KEY)")
	cmd.Flags().StringVar(&flagEndpoint, "aws-endpoint", "", "S3 endpoint (default: PROBOD_AWS_ENDPOINT)")
	cmd.Flags().BoolVar(&flagUsePathStyle, "aws-use-path-style", false, "Use path-style S3 URLs (default: PROBOD_AWS_USE_PATH_STYLE)")
	cmd.Flags().BoolVar(&flagDryRun, "dry-run", false, "List signatures that need a zip without writing")
	cmd.Flags().BoolVar(&flagYes, "yes", false, "Skip confirmation")
	cmd.Flags().BoolVar(&flagContinueOnError, "continue-on-error", false, "Keep going when a signature fails; exit non-zero if any failed")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		out := f.IOStreams.Out

		pgClient, err := f.PgClient()
		if err != nil {
			return err
		}

		logger := log.NewLogger(log.WithOutput(f.IOStreams.ErrOut))
		svc := esign.NewService(
			pgClient,
			nil,
			html2pdf.NewConverter(""),
			"",
			flagBucket,
			logger,
		)

		ids, err := svc.MissingBundleIDs(ctx)
		if err != nil {
			return err
		}

		if len(ids) == 0 {
			_, _ = fmt.Fprintln(out, "No signatures are missing a bundle.")

			return nil
		}

		_, _ = fmt.Fprintf(out, "Found %d signature(s) missing a bundle.\n", len(ids))

		if flagDryRun {
			for _, id := range ids {
				_, _ = fmt.Fprintln(out, id.String())
			}

			return nil
		}

		if !flagYes {
			return fmt.Errorf(
				"about to store %d signature bundle(s); pass --yes to proceed or --dry-run to preview",
				len(ids),
			)
		}

		bucket, region, accessKeyID, secretAccessKey, endpoint, usePathStyle := signatureBundleAWS(
			flagBucket,
			flagRegion,
			flagAccessKeyID,
			flagSecretAccessKey,
			flagEndpoint,
			flagUsePathStyle,
			cmd.Flags().Changed("aws-use-path-style"),
		)
		if bucket == "" {
			return fmt.Errorf("set --bucket or PROBOD_AWS_BUCKET")
		}

		files, err := newFileManager(logger, region, accessKeyID, secretAccessKey, endpoint, usePathStyle)
		if err != nil {
			return err
		}

		svc = esign.NewService(
			pgClient,
			files,
			html2pdf.NewConverter(""),
			"",
			bucket,
			logger,
		)

		var failures int

		for _, id := range ids {
			if err := svc.StoreBundle(ctx, id); err != nil {
				failures++
				_, _ = fmt.Fprintf(f.IOStreams.ErrOut, "%s: %v\n", id.String(), err)

				if !flagContinueOnError {
					return fmt.Errorf("stopped after %d failure(s)", failures)
				}

				continue
			}

			_, _ = fmt.Fprintf(out, "stored %s\n", id.String())
		}

		_, _ = fmt.Fprintf(out, "Done. Stored %d of %d signature bundle(s); %d failure(s).\n", len(ids)-failures, len(ids), failures)

		if failures > 0 {
			return fmt.Errorf("%d signature(s) failed", failures)
		}

		return nil
	}

	return cmd
}

func newFileManager(
	logger *log.Logger,
	region string,
	accessKeyID string,
	secretAccessKey string,
	endpoint string,
	usePathStyle bool,
) (*filemanager.Service, error) {
	awsConfig, err := awsconfig.NewConfig(
		logger,
		httpclient.DefaultPooledClient(httpclient.WithLogger(logger)),
		awsconfig.Options{
			Region:          region,
			AccessKeyID:     accessKeyID,
			SecretAccessKey: secretAccessKey,
			Endpoint:        endpoint,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot initialize AWS config: %w", err)
	}

	client := s3.NewFromConfig(awsConfig, func(o *s3.Options) {
		o.UsePathStyle = usePathStyle
	})

	return filemanager.NewService(nil, nil, client, logger), nil
}

func signatureBundleAWS(
	bucket string,
	region string,
	accessKeyID string,
	secretAccessKey string,
	endpoint string,
	usePathStyle bool,
	usePathStyleSet bool,
) (string, string, string, string, string, bool) {
	if bucket == "" {
		bucket = os.Getenv("PROBOD_AWS_BUCKET")
	}

	if region == "" {
		region = os.Getenv("PROBOD_AWS_REGION")
	}

	if accessKeyID == "" {
		accessKeyID = os.Getenv("PROBOD_AWS_ACCESS_KEY_ID")
	}

	if secretAccessKey == "" {
		secretAccessKey = os.Getenv("PROBOD_AWS_SECRET_ACCESS_KEY")
	}

	if endpoint == "" {
		endpoint = os.Getenv("PROBOD_AWS_ENDPOINT")
	}

	if !usePathStyleSet {
		usePathStyle = envBool(os.Getenv("PROBOD_AWS_USE_PATH_STYLE"))
	}

	return bucket, region, accessKeyID, secretAccessKey, endpoint, usePathStyle
}

func envBool(value string) bool {
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false
	}

	return parsed
}
