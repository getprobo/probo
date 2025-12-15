package validator

import (
	_ "embed"
	"strings"
)

var (
	//go:embed data/disposable-email-domains/disposable_email_blocklist.conf
	disposableEmailsRaw []byte
	testEmails          = []string{
		"acme.com",
		"acme.net",
		"acme.org",
		"ethereal.email",
		"example.com",
		"example.net",
		"example.org",
		"mailhog.local",
		"mailslurp.com",
		"test.com",
		"test.net",
		"test.org",
		"localhost.localdomain",
	}
	blacklistedEmails = append(
		strings.Split(strings.TrimSpace(string(disposableEmailsRaw)), "\n"),
		testEmails...,
	)
	notOneOfBlacklisted = NotOneOfSlice(blacklistedEmails)
)

func NotBlacklisted() ValidatorFunc {
	return func(value any) *ValidationError {
		err := notOneOfBlacklisted(value)

		if err != nil {
			return newValidationError(
				ErrorCodeInvalidEmail,
				"must not be blacklisted",
			)
		}

		return nil
	}
}
