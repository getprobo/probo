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

package provider

// NewBuiltinRegistry returns a *Registry populated with every
// connector provider compiled into the binary. It panics on duplicate
// registration or invalid Registration metadata — both are programmer
// errors caught at process start, not at runtime. Probod calls this
// once at startup and threads the *Registry into every consumer.
func NewBuiltinRegistry() *Registry {
	r := NewRegistry()
	for _, reg := range []*Registration{
		anthropicRegistration(),
		apolloRegistration(),
		asanaRegistration(),
		betterStackRegistration(),
		bitbucketRegistration(),
		brevoRegistration(),
		brexRegistration(),
		clickhouseRegistration(),
		clickupRegistration(),
		cloudflareRegistration(),
		crispRegistration(),
		cursorRegistration(),
		datadogRegistration(),
		deepgramRegistration(),
		docusignRegistration(),
		grafanaRegistration(),
		githubRegistration(),
		gitlabRegistration(),
		googleWorkspaceRegistration(),
		herokuRegistration(),
		hubspotRegistration(),
		incidentioRegistration(),
		intercomRegistration(),
		langfuseRegistration(),
		linearRegistration(),
		mercuryRegistration(),
		metabaseRegistration(),
		microsoft365Registration(),
		mondayRegistration(),
		neonRegistration(),
		netlifyRegistration(),
		notionRegistration(),
		oktaRegistration(),
		onePasswordRegistration(),
		openaiRegistration(),
		openrouterRegistration(),
		posthogRegistration(),
		pagerdutyRegistration(),
		pylonRegistration(),
		qoveryRegistration(),
		railwayRegistration(),
		renderRegistration(),
		resendRegistration(),
		scalewayRegistration(),
		sendgridRegistration(),
		sentryRegistration(),
		signozRegistration(),
		slackRegistration(),
		supabaseRegistration(),
		tailscaleRegistration(),
		tallyRegistration(),
		vercelRegistration(),
		yousignRegistration(),
		zendeskRegistration(),
	} {
		if err := r.Register(reg); err != nil {
			panic(err)
		}
	}

	return r
}
