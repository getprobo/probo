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

package cmdutil

import (
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cli/config"
	"go.probo.inc/probo/pkg/cmd/iostreams"
)

type Factory struct {
	IOStreams *iostreams.IOStreams
	Version   string
	Config    func() (*config.Config, error)
}

// TokenRefreshOption returns an api.Option that enables automatic access
// token refresh using the stored OAuth2 refresh token. If the host config
// has no refresh token or token endpoint, a no-op option is returned.
func TokenRefreshOption(
	cfg *config.Config,
	host string,
	hc *config.HostConfig,
) api.Option {
	if hc.RefreshToken == "" || hc.TokenEndpoint == "" {
		return func(*api.Client) {}
	}

	return api.WithTokenRefresher(&api.TokenRefresher{
		RefreshToken:  hc.RefreshToken,
		TokenEndpoint: hc.TokenEndpoint,
		ClientID:      config.CLIClientID,
		OnRefresh: func(accessToken, refreshToken string) error {
			hc.Token = accessToken
			hc.RefreshToken = refreshToken
			cfg.Hosts[host] = hc

			return cfg.Save()
		},
	})
}
