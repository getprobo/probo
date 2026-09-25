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

package drivers

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/account"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"go.gearno.de/kit/log"
	cloudaws "go.probo.inc/probo/pkg/cloud/aws"
)

// awsNameResolver names the connected account for the source-name worker.
// The official account name is preferred, then the IAM sign-in alias, then
// the 12-digit account ID so two sources in one organization still differ.
type awsNameResolver struct {
	session *cloudaws.Session
	logger  *log.Logger
}

func NewAWSNameResolver(session *cloudaws.Session, logger *log.Logger) NameResolver {
	return &awsNameResolver{session: session, logger: logger}
}

func (r *awsNameResolver) ResolveInstanceName(ctx context.Context) (string, error) {
	accountName, err := r.accountName(ctx)
	if err != nil {
		if ctx.Err() != nil {
			return "", err
		}

		r.logger.WarnCtx(ctx, "cannot read aws account name, trying alias", log.Error(err))
	}

	if name := strings.TrimSpace(accountName); name != "" {
		return name, nil
	}

	alias, err := r.accountAlias(ctx)
	if err != nil {
		if ctx.Err() != nil {
			return "", err
		}

		r.logger.WarnCtx(ctx, "cannot read aws account alias, using account id", log.Error(err))
	}

	return pickAWSInstanceName(accountName, alias, r.session.AccountID()), nil
}

func (r *awsNameResolver) accountName(ctx context.Context) (string, error) {
	out, err := account.NewFromConfig(r.session.Config()).GetAccountInformation(
		ctx,
		&account.GetAccountInformationInput{},
	)
	if err != nil {
		return "", fmt.Errorf("cannot get aws account information: %w", err)
	}

	if out.AccountName == nil {
		return "", nil
	}

	return *out.AccountName, nil
}

func (r *awsNameResolver) accountAlias(ctx context.Context) (string, error) {
	out, err := iam.NewFromConfig(r.session.Config()).ListAccountAliases(
		ctx,
		&iam.ListAccountAliasesInput{},
	)
	if err != nil {
		return "", fmt.Errorf("cannot list aws account aliases: %w", err)
	}

	if len(out.AccountAliases) == 0 {
		return "", nil
	}

	return out.AccountAliases[0], nil
}

func pickAWSInstanceName(accountName, alias, accountID string) string {
	if name := strings.TrimSpace(accountName); name != "" {
		return name
	}

	if name := strings.TrimSpace(alias); name != "" {
		return name
	}

	return accountID
}
