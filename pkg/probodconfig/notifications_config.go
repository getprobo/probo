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

package probodconfig

type NotificationsConfig struct {
	Mailer   MailerConfig               `json:"mailer"`
	Slack    SlackConfig                `json:"slack"`
	Webhook  WebhookConfig              `json:"webhook"`
	Document DocumentNotificationConfig `json:"document"`
}

type WebhookConfig struct {
	SenderInterval int `json:"sender-interval"`
	CacheTTL       int `json:"cache-ttl"`
}

// DocumentNotificationConfig configures the debounced worker that batches
// signature and approval request notifications. All durations are in seconds.
type DocumentNotificationConfig struct {
	// Interval is how often the worker scans for pending requests.
	Interval int `json:"interval"`
	// DebounceDelay is how long a request must have been pending before its
	// first notification is sent.
	DebounceDelay int `json:"debounce-delay"`
	// ReminderInterval is the base reminder cadence. Reminders are sent at
	// 1x, 2x and 3x this interval after the previous email, then stop.
	ReminderInterval int `json:"reminder-interval"`
}
