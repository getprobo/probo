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
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

const (
	awsSecretsManagerRefPrefix       = "awssm://"
	awsSecretsManagerLegacyRefPrefix = "aws://"
	awsParameterStoreRefPrefix       = "awsps://"
)

type Resolver struct {
	lookup   EnvGetter
	smClient *secretsmanager.Client
	psClient *ssm.Client
	smCache  map[string]string
	psCache  map[string]string
	err      error
}

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

	if prefix, empty := emptyAWSRefPrefix(raw); empty {
		return "", fmt.Errorf("cannot resolve %s: empty AWS reference after %s", key, prefix)
	}

	if secretID, ok := parseAWSSecretsManagerRef(raw); ok {
		value, err := r.loadPlaintextSecret(secretID)
		if err != nil {
			return "", fmt.Errorf("cannot resolve %s: %w", key, err)
		}

		return value, nil
	}

	if paramName, ok := parseAWSRef(raw, awsParameterStoreRefPrefix); ok {
		value, err := r.loadParameter(paramName)
		if err != nil {
			return "", fmt.Errorf("cannot resolve %s: %w", key, err)
		}

		return value, nil
	}

	return raw, nil
}

func (r *Resolver) loadPlaintextSecret(secretID string) (string, error) {
	if r.smCache != nil {
		if value, ok := r.smCache[secretID]; ok {
			return value, nil
		}
	}

	client, err := r.secretsManagerClient(context.Background())
	if err != nil {
		return "", err
	}

	out, err := client.GetSecretValue(context.Background(), &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretID),
	})
	if err != nil {
		return "", fmt.Errorf("cannot load secret from AWS Secrets Manager: %w", err)
	}

	if out.SecretString == nil || *out.SecretString == "" {
		return "", fmt.Errorf("secret %q has an empty SecretString", secretID)
	}

	value := *out.SecretString

	if r.smCache == nil {
		r.smCache = make(map[string]string)
	}

	r.smCache[secretID] = value

	return value, nil
}

func (r *Resolver) loadParameter(name string) (string, error) {
	if r.psCache != nil {
		if value, ok := r.psCache[name]; ok {
			return value, nil
		}
	}

	client, err := r.parameterStoreClient(context.Background())
	if err != nil {
		return "", err
	}

	out, err := client.GetParameter(context.Background(), &ssm.GetParameterInput{
		Name:           aws.String(name),
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		return "", fmt.Errorf("cannot load parameter from AWS Systems Manager Parameter Store: %w", err)
	}

	if out.Parameter == nil || out.Parameter.Value == nil || *out.Parameter.Value == "" {
		return "", fmt.Errorf("parameter %q has an empty value", name)
	}

	value := *out.Parameter.Value

	if r.psCache == nil {
		r.psCache = make(map[string]string)
	}

	r.psCache[name] = value

	return value, nil
}

func (r *Resolver) secretsManagerClient(ctx context.Context) (*secretsmanager.Client, error) {
	if r.smClient != nil {
		return r.smClient, nil
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot load AWS config: %w", err)
	}

	r.smClient = secretsmanager.NewFromConfig(cfg)

	return r.smClient, nil
}

func (r *Resolver) parameterStoreClient(ctx context.Context) (*ssm.Client, error) {
	if r.psClient != nil {
		return r.psClient, nil
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot load AWS config: %w", err)
	}

	r.psClient = ssm.NewFromConfig(cfg)

	return r.psClient, nil
}

func parseAWSRef(value, prefix string) (string, bool) {
	if !strings.HasPrefix(value, prefix) {
		return "", false
	}

	ref := strings.TrimPrefix(value, prefix)
	if ref == "" {
		return "", false
	}

	return ref, true
}

func emptyAWSRefPrefix(value string) (string, bool) {
	for _, prefix := range []string{
		awsSecretsManagerRefPrefix,
		awsSecretsManagerLegacyRefPrefix,
		awsParameterStoreRefPrefix,
	} {
		if strings.HasPrefix(value, prefix) && strings.TrimPrefix(value, prefix) == "" {
			return prefix, true
		}
	}

	return "", false
}

func parseAWSSecretsManagerRef(value string) (string, bool) {
	if secretID, ok := parseAWSRef(value, awsSecretsManagerRefPrefix); ok {
		return secretID, true
	}

	return parseAWSRef(value, awsSecretsManagerLegacyRefPrefix)
}

func parseAWSParameterStoreRef(value string) (string, bool) {
	return parseAWSRef(value, awsParameterStoreRefPrefix)
}
