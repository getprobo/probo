// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

package awsconfig

import (
	"context"
	"fmt"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"go.gearno.de/kit/httpclient"
	"go.gearno.de/kit/log"
)

type (
	Options struct {
		SessionName     string
		Endpoint        string
		Region          string
		AccessKeyID     string
		SecretAccessKey string
	}
)

const (
	DefaultRegion      = "us-east-2"
	DefaultSessionName = "go.probo.inc/probo"
)

func NewConfig(logger *log.Logger, httpClient *http.Client, opts Options) (aws.Config, error) {
	if opts.Region == "" {
		opts.Region = DefaultRegion
	}

	logger = logger.Named(
		"aws.client",
		log.WithAttributes(
			log.String("region", opts.Region),
			log.String("endpoint", opts.Endpoint),
			log.String("session_name", opts.SessionName),
		),
	)

	if httpClient == nil {
		httpClient = httpclient.DefaultPooledClient(httpclient.WithLogger(logger))
	}

	loadOpts := []func(*config.LoadOptions) error{
		config.WithRegion(opts.Region),
		config.WithHTTPClient(httpClient),
	}

	if opts.AccessKeyID != "" && opts.SecretAccessKey != "" {
		loadOpts = append(loadOpts, config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				opts.AccessKeyID,
				opts.SecretAccessKey,
				opts.SessionName,
			),
		))
	}

	cfg, err := config.LoadDefaultConfig(context.Background(), loadOpts...)
	if err != nil {
		return aws.Config{}, fmt.Errorf("cannot load AWS config: %w", err)
	}

	if opts.Endpoint != "" {
		cfg.BaseEndpoint = new(opts.Endpoint)
	}

	return cfg, nil
}
