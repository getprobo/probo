// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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

package saml

import (
	"fmt"
	"strings"

	"github.com/crewjam/saml"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/mail"
)

func extractUserAttributes(assertion *saml.Assertion, config *coredata.SAMLConfiguration) (mail.Addr, string, *coredata.MembershipRole, error) {
	var (
		email    mail.Addr
		fullname string
		role     *coredata.MembershipRole
	)

	if len(assertion.AttributeStatements) == 0 {
		if assertion.Subject != nil && assertion.Subject.NameID != nil {
			email, err := mail.ParseAddr(assertion.Subject.NameID.Value)
			if err != nil {
				return mail.Nil, "", nil, fmt.Errorf("cannot parse email: %w", err)
			}

			fullname = email.String()
			role = nil
			return email, fullname, role, nil
		}

		return mail.Nil, "", nil, fmt.Errorf("no attribute statement and no NameID in assertion")
	}

	emailString, err := extractAttributeValue(assertion, config.AttributeEmail)
	if err != nil {
		if assertion.Subject != nil && assertion.Subject.NameID != nil {
			emailString = assertion.Subject.NameID.Value
		} else {
			return mail.Nil, "", nil, fmt.Errorf("cannot extract email: %w", err)
		}
	}

	email, err = mail.ParseAddr(emailString)
	if err != nil {
		return mail.Nil, "", nil, fmt.Errorf("cannot parse email: %w", err)
	}

	firstname, err := extractAttributeValue(assertion, config.AttributeFirstname)
	if err != nil {
		firstname = ""
	}

	lastname, err := extractAttributeValue(assertion, config.AttributeLastname)
	if err != nil {
		lastname = ""
	}

	if firstname != "" && lastname != "" {
		fullname = strings.TrimSpace(firstname + " " + lastname)
	} else if firstname != "" {
		fullname = firstname
	} else if lastname != "" {
		fullname = lastname
	} else {
		fullname = email.String()
	}

	roleString, err := extractAttributeValue(assertion, config.AttributeRole)
	if err != nil {
		role = nil
	}

	if roleString != "" {
		role = mapSAMLRoleToSystemRole(roleString)
	}

	return email, fullname, role, nil
}

func extractAttributeValue(assertion *saml.Assertion, attributeName string) (string, error) {
	if len(assertion.AttributeStatements) == 0 {
		return "", fmt.Errorf("no attribute statement in assertion")
	}

	for _, statement := range assertion.AttributeStatements {
		for _, attr := range statement.Attributes {
			if attr.Name == attributeName {
				if len(attr.Values) == 0 {
					return "", fmt.Errorf("attribute %q has no values", attributeName)
				}

				return attr.Values[0].Value, nil
			}
		}
	}

	return "", fmt.Errorf("attribute %q not found in assertion", attributeName)
}

func extractEmailFromAssertion(assertion *saml.Assertion) (string, error) {
	commonEmailAttributes := []string{
		"email",
		"Email",
		"emailAddress",
		"mail",
		"http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress",
		"http://schemas.xmlsoap.org/claims/EmailAddress",
	}

	for _, attrName := range commonEmailAttributes {
		email, err := extractAttributeValue(assertion, attrName)
		if err == nil && email != "" {
			return email, nil
		}
	}

	if assertion.Subject != nil && assertion.Subject.NameID != nil && assertion.Subject.NameID.Value != "" {
		return assertion.Subject.NameID.Value, nil
	}

	return "", fmt.Errorf("could not extract email from assertion")
}

func extractEmailDomain(email string) (string, error) {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid email address: %s", email)
	}

	domain := strings.ToLower(strings.TrimSpace(parts[1]))
	if domain == "" {
		return "", fmt.Errorf("empty domain in email address: %s", email)
	}

	return domain, nil
}

func mapSAMLRoleToSystemRole(samlRole string) *coredata.MembershipRole {
	if samlRole != "" && isValidRole(samlRole) {
		role := coredata.MembershipRole(samlRole)
		return &role
	}

	return nil
}

func isValidRole(role string) bool {
	switch role {
	case "OWNER", "ADMIN", "EMPLOYEE", "VIEWER":
		return true
	default:
		return false
	}
}
