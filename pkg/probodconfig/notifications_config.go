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
