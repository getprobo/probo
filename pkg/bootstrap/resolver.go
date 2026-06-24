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

package bootstrap

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

const awsSecretRefPrefix = "aws://"

type (
	SecretsManagerClient interface {
		GetSecretValue(
			ctx context.Context,
			params *secretsmanager.GetSecretValueInput,
			optFns ...func(*secretsmanager.Options),
		) (*secretsmanager.GetSecretValueOutput, error)
	}

	Resolver struct {
		lookup               EnvGetter
		secretsManagerClient SecretsManagerClient
		smCache              map[string]string
		err                  error
	}
)

func NewResolver(lookup EnvGetter) *Resolver {
	if lookup == nil {
		lookup = os.Getenv
	}

	return &Resolver{lookup: lookup}
}

func (r *Resolver) Err() error {
	return r.err
}

func (r *Resolver) getEnv(key string) string {
	if r.err != nil {
		return ""
	}

	value, err := r.resolve(key)
	if err != nil {
		r.err = err
		return ""
	}

	return value
}

func (r *Resolver) getEnvOrDefault(key, defaultValue string) string {
	if value := r.getEnv(key); value != "" {
		return value
	}

	return defaultValue
}

func (r *Resolver) getEnvIntOrDefault(key string, defaultValue int) int {
	if value := r.getEnv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 32); err == nil {
			return int(intValue)
		}
	}

	return defaultValue
}

func (r *Resolver) getEnvFloatOrDefault(key string, defaultValue float64) float64 {
	if value := r.getEnv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}

	return defaultValue
}

func (r *Resolver) getEnvFloatPtr(key string) *float64 {
	if value := r.getEnv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return &floatValue
		}
	}

	return nil
}

func (r *Resolver) getEnvIntPtr(key string) *int {
	if value := r.getEnv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 32); err == nil {
			v := int(intValue)
			return &v
		}
	}

	return nil
}

func (r *Resolver) getEnvBoolOrDefault(key string, defaultValue bool) bool {
	if value := r.getEnv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}

	return defaultValue
}

func (r *Resolver) resolve(key string) (string, error) {
	raw := r.lookup(key)

	secretID, ok := parseAWSSecretRef(raw)
	if !ok {
		return raw, nil
	}

	value, err := r.loadPlaintextSecret(secretID)
	if err != nil {
		return "", fmt.Errorf("cannot resolve %s: %w", key, err)
	}

	return value, nil
}

func (r *Resolver) loadPlaintextSecret(secretID string) (string, error) {
	if r.smCache != nil {
		if value, ok := r.smCache[secretID]; ok {
			return value, nil
		}
	}

	value, err := fetchPlaintextSecret(
		context.Background(),
		secretsManagerOptions{
			SecretID: secretID,
			Client:   r.secretsManagerClient,
		},
	)
	if err != nil {
		return "", err
	}

	if r.smCache == nil {
		r.smCache = make(map[string]string)
	}

	r.smCache[secretID] = value

	return value, nil
}

func parseAWSSecretRef(value string) (string, bool) {
	if !strings.HasPrefix(value, awsSecretRefPrefix) {
		return "", false
	}

	secretID := strings.TrimPrefix(value, awsSecretRefPrefix)
	if secretID == "" {
		return "", false
	}

	return secretID, true
}

type secretsManagerOptions struct {
	SecretID string
	Client   SecretsManagerClient
}

func fetchPlaintextSecret(ctx context.Context, opts secretsManagerOptions) (string, error) {
	client, err := secretsManagerClient(ctx, opts)
	if err != nil {
		return "", err
	}

	out, err := client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(opts.SecretID),
	})
	if err != nil {
		return "", fmt.Errorf("cannot load secret from AWS Secrets Manager: %w", err)
	}

	if out.SecretString == nil || *out.SecretString == "" {
		return "", fmt.Errorf("secret %q has an empty SecretString", opts.SecretID)
	}

	return *out.SecretString, nil
}

func secretsManagerClient(ctx context.Context, opts secretsManagerOptions) (SecretsManagerClient, error) {
	if opts.Client != nil {
		return opts.Client, nil
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot load AWS config: %w", err)
	}

	return secretsmanager.NewFromConfig(cfg), nil
}
