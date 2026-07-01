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
		clerkRegistration(),
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
