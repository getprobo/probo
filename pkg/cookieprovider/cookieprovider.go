// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

package cookieprovider

import (
	"go.probo.inc/probo/pkg/coredata"
)

type (
	Category string

	Cookie struct {
		Name        string
		Duration    string
		Description string
	}

	Provider struct {
		Key         string
		Name        string
		Description string
		Category    Category
		WebsiteURL  string
		Cookies     []Cookie
	}
)

const (
	CategoryNecessary   Category = "Necessary"
	CategoryAnalytics   Category = "Analytics"
	CategoryMarketing   Category = "Marketing"
	CategoryPreferences Category = "Preferences"
)

var providersByKey map[string]*Provider

func init() {
	providersByKey = make(map[string]*Provider, len(providers))
	for i := range providers {
		providersByKey[providers[i].Key] = &providers[i]
	}
}

func All() []Provider {
	result := make([]Provider, len(providers))
	copy(result, providers)
	return result
}

func ByCategory(category Category) []Provider {
	var result []Provider
	for _, p := range providers {
		if p.Category == category {
			result = append(result, p)
		}
	}
	return result
}

func ByKey(key string) (Provider, bool) {
	p, ok := providersByKey[key]
	if !ok {
		return Provider{}, false
	}
	return *p, true
}

func (p Provider) CookieItems() coredata.CookieItems {
	items := make(coredata.CookieItems, len(p.Cookies))
	for i, c := range p.Cookies {
		items[i] = coredata.CookieItem{
			Name:        c.Name,
			Duration:    c.Duration,
			Description: c.Description,
		}
	}
	return items
}

var providers = []Provider{
	{
		Key:         "google-analytics",
		Name:        "Google Analytics",
		Description: "Web analytics service that tracks and reports website traffic.",
		Category:    CategoryAnalytics,
		WebsiteURL:  "https://analytics.google.com",
		Cookies: []Cookie{
			{Name: "_ga", Duration: "2 years", Description: "Distinguishes unique users by assigning a randomly generated number as a client identifier."},
			{Name: "_ga_*", Duration: "2 years", Description: "Used by Google Analytics to persist session state."},
			{Name: "_gid", Duration: "24 hours", Description: "Distinguishes unique users."},
			{Name: "_gat", Duration: "1 minute", Description: "Used to throttle the request rate to Google Analytics."},
		},
	},
	{
		Key:         "google-tag-manager",
		Name:        "Google Tag Manager",
		Description: "Tag management system that allows managing JavaScript and HTML tags for tracking and analytics.",
		Category:    CategoryNecessary,
		WebsiteURL:  "https://tagmanager.google.com",
		Cookies: []Cookie{
			{Name: "_gcl_au", Duration: "90 days", Description: "Used by Google AdSense to store and track conversions."},
		},
	},
	{
		Key:         "google-ads",
		Name:        "Google Ads",
		Description: "Online advertising platform for displaying ads on Google search results and partner websites.",
		Category:    CategoryMarketing,
		WebsiteURL:  "https://ads.google.com",
		Cookies: []Cookie{
			{Name: "_gcl_aw", Duration: "90 days", Description: "Stores click information from Google Ads for conversion tracking."},
			{Name: "_gcl_dc", Duration: "90 days", Description: "Stores click information from DoubleClick for conversion tracking."},
			{Name: "IDE", Duration: "1 year", Description: "Used by DoubleClick to register and report ad clicks."},
			{Name: "test_cookie", Duration: "15 minutes", Description: "Used to check if the browser supports cookies."},
		},
	},
	{
		Key:         "facebook-pixel",
		Name:        "Facebook Pixel",
		Description: "Analytics tool for measuring the effectiveness of advertising on Facebook.",
		Category:    CategoryMarketing,
		WebsiteURL:  "https://www.facebook.com/business/tools/meta-pixel",
		Cookies: []Cookie{
			{Name: "_fbp", Duration: "90 days", Description: "Used by Facebook to deliver advertising products."},
			{Name: "_fbc", Duration: "90 days", Description: "Stores click identifier from Facebook ad campaigns."},
			{Name: "fr", Duration: "90 days", Description: "Used by Facebook for advertising and tracking."},
		},
	},
	{
		Key:         "hubspot",
		Name:        "HubSpot",
		Description: "Marketing, sales, and service software platform with tracking and analytics capabilities.",
		Category:    CategoryMarketing,
		WebsiteURL:  "https://www.hubspot.com",
		Cookies: []Cookie{
			{Name: "__hssc", Duration: "30 minutes", Description: "Tracks sessions and determines if a new session should be created."},
			{Name: "__hssrc", Duration: "Session", Description: "Indicates if the visitor has restarted the browser."},
			{Name: "__hstc", Duration: "180 days", Description: "Tracks visitors across sessions for HubSpot analytics."},
			{Name: "hubspotutk", Duration: "180 days", Description: "Keeps track of a visitor's identity and is passed to HubSpot on form submission."},
		},
	},
	{
		Key:         "hotjar",
		Name:        "Hotjar",
		Description: "Behavior analytics and user feedback service that helps understand user interactions.",
		Category:    CategoryAnalytics,
		WebsiteURL:  "https://www.hotjar.com",
		Cookies: []Cookie{
			{Name: "_hj*", Duration: "Varies", Description: "Used by Hotjar to collect usage statistics and behavior data."},
			{Name: "_hjSession_*", Duration: "30 minutes", Description: "Ensures subsequent requests in a session window are attributed to the same session."},
			{Name: "_hjSessionUser_*", Duration: "1 year", Description: "Ensures data from subsequent visits is attributed to the same user."},
		},
	},
	{
		Key:         "intercom",
		Name:        "Intercom",
		Description: "Customer messaging platform for sales, marketing, and support.",
		Category:    CategoryPreferences,
		WebsiteURL:  "https://www.intercom.com",
		Cookies: []Cookie{
			{Name: "intercom-id-*", Duration: "9 months", Description: "Anonymous visitor identifier for Intercom messenger."},
			{Name: "intercom-session-*", Duration: "1 week", Description: "Identifies sessions so that users can access their conversations."},
		},
	},
	{
		Key:         "linkedin-insight",
		Name:        "LinkedIn Insight",
		Description: "Analytics tool for tracking conversions and retargeting from LinkedIn campaigns.",
		Category:    CategoryMarketing,
		WebsiteURL:  "https://business.linkedin.com",
		Cookies: []Cookie{
			{Name: "li_sugr", Duration: "90 days", Description: "Used by LinkedIn for browser identification."},
			{Name: "bcookie", Duration: "1 year", Description: "LinkedIn browser identifier."},
			{Name: "lidc", Duration: "24 hours", Description: "Used for routing and data center selection."},
			{Name: "UserMatchHistory", Duration: "30 days", Description: "Used for LinkedIn Ads ID syncing."},
		},
	},
	{
		Key:         "stripe",
		Name:        "Stripe",
		Description: "Payment processing platform for internet businesses.",
		Category:    CategoryNecessary,
		WebsiteURL:  "https://stripe.com",
		Cookies: []Cookie{
			{Name: "__stripe_mid", Duration: "1 year", Description: "Used for fraud prevention and detection."},
			{Name: "__stripe_sid", Duration: "30 minutes", Description: "Used for fraud prevention and detection during a session."},
		},
	},
	{
		Key:         "cloudflare",
		Name:        "Cloudflare",
		Description: "Web infrastructure and security company providing CDN, DDoS protection, and bot management.",
		Category:    CategoryNecessary,
		WebsiteURL:  "https://www.cloudflare.com",
		Cookies: []Cookie{
			{Name: "__cf_bm", Duration: "30 minutes", Description: "Used by Cloudflare Bot Management to identify and distinguish bots from humans."},
			{Name: "cf_clearance", Duration: "30 minutes", Description: "Stores proof that a visitor has successfully completed a challenge."},
		},
	},
	{
		Key:         "cookiebot",
		Name:        "Cookiebot",
		Description: "Cookie consent management platform for GDPR and ePrivacy compliance.",
		Category:    CategoryNecessary,
		WebsiteURL:  "https://www.cookiebot.com",
		Cookies: []Cookie{
			{Name: "CookieConsent", Duration: "1 year", Description: "Stores the user's cookie consent state."},
			{Name: "CookieConsentBulkTicket", Duration: "1 year", Description: "Enables consent across multiple websites."},
		},
	},
	{
		Key:         "onetrust",
		Name:        "OneTrust",
		Description: "Privacy management and cookie compliance platform.",
		Category:    CategoryNecessary,
		WebsiteURL:  "https://www.onetrust.com",
		Cookies: []Cookie{
			{Name: "OptanonConsent", Duration: "1 year", Description: "Stores the user's cookie consent preferences."},
			{Name: "OptanonAlertBoxClosed", Duration: "1 year", Description: "Records that the cookie consent banner has been dismissed."},
		},
	},
	{
		Key:         "twitter-x",
		Name:        "Twitter / X",
		Description: "Social media platform advertising and tracking pixels.",
		Category:    CategoryMarketing,
		WebsiteURL:  "https://ads.x.com",
		Cookies: []Cookie{
			{Name: "muc_ads", Duration: "2 years", Description: "Used by Twitter/X for advertising purposes."},
			{Name: "guest_id", Duration: "2 years", Description: "Unique identifier for the visitor assigned by Twitter/X."},
			{Name: "personalization_id", Duration: "2 years", Description: "Used by Twitter/X to deliver personalized ads."},
		},
	},
	{
		Key:         "tiktok",
		Name:        "TikTok",
		Description: "Social media platform advertising and tracking pixel for measuring ad performance.",
		Category:    CategoryMarketing,
		WebsiteURL:  "https://ads.tiktok.com",
		Cookies: []Cookie{
			{Name: "_ttp", Duration: "13 months", Description: "Used by TikTok to track and attribute conversions."},
			{Name: "tt_webid", Duration: "13 months", Description: "Used by TikTok to identify the visitor."},
			{Name: "tt_webid_v2", Duration: "13 months", Description: "Used by TikTok to identify the visitor (updated version)."},
		},
	},
	{
		Key:         "zendesk",
		Name:        "Zendesk",
		Description: "Customer service and engagement platform with live chat and support widgets.",
		Category:    CategoryPreferences,
		WebsiteURL:  "https://www.zendesk.com",
		Cookies: []Cookie{
			{Name: "__zlcmid", Duration: "1 year", Description: "Used by Zendesk live chat to store a unique user identifier."},
		},
	},
	{
		Key:         "drift",
		Name:        "Drift",
		Description: "Conversational marketing and sales platform for live chat and chatbots.",
		Category:    CategoryMarketing,
		WebsiteURL:  "https://www.drift.com",
		Cookies: []Cookie{
			{Name: "drift_aid", Duration: "2 years", Description: "Anonymous identifier for the visitor."},
			{Name: "drift_campaign_refresh", Duration: "30 minutes", Description: "Controls campaign display frequency."},
			{Name: "driftt_aid", Duration: "2 years", Description: "Anonymous identifier used for tracking across sessions."},
		},
	},
	{
		Key:         "segment",
		Name:        "Segment",
		Description: "Customer data platform for collecting, cleaning, and controlling customer data.",
		Category:    CategoryAnalytics,
		WebsiteURL:  "https://segment.com",
		Cookies: []Cookie{
			{Name: "ajs_anonymous_id", Duration: "1 year", Description: "Anonymous user identifier for tracking before a user is identified."},
			{Name: "ajs_user_id", Duration: "1 year", Description: "Stores the identified user ID after identification."},
		},
	},
	{
		Key:         "mixpanel",
		Name:        "Mixpanel",
		Description: "Product analytics platform for tracking user interactions and engagement.",
		Category:    CategoryAnalytics,
		WebsiteURL:  "https://mixpanel.com",
		Cookies: []Cookie{
			{Name: "mp_*_mixpanel", Duration: "1 year", Description: "Stores tracking information for Mixpanel analytics."},
		},
	},
	{
		Key:         "amplitude",
		Name:        "Amplitude",
		Description: "Product analytics platform for understanding user behavior and engagement.",
		Category:    CategoryAnalytics,
		WebsiteURL:  "https://amplitude.com",
		Cookies: []Cookie{
			{Name: "AMP_*", Duration: "1 year", Description: "Stores user and session information for Amplitude analytics."},
		},
	},
	{
		Key:         "microsoft-clarity",
		Name:        "Microsoft Clarity",
		Description: "Free analytics tool that captures how users interact with your site via session recordings and heatmaps.",
		Category:    CategoryAnalytics,
		WebsiteURL:  "https://clarity.microsoft.com",
		Cookies: []Cookie{
			{Name: "_clck", Duration: "1 year", Description: "Persists the Clarity user ID and preferences."},
			{Name: "_clsk", Duration: "1 day", Description: "Connects multiple page views by a user into a single session recording."},
			{Name: "CLID", Duration: "1 year", Description: "Identifies the first-time Clarity saw this user on any site."},
		},
	},
}
