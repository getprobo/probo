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

package auth

import (
	"fmt"
	"strings"

	"github.com/crewjam/saml"
)

func ExtractAttributeValue(assertion *saml.Assertion, attributeName string) (string, error) {
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

func ExtractEmailFromAssertion(assertion *saml.Assertion) (string, error) {
	commonEmailAttributes := []string{
		"email",
		"Email",
		"emailAddress",
		"mail",
		"http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress",
		"http://schemas.xmlsoap.org/claims/EmailAddress",
	}

	for _, attrName := range commonEmailAttributes {
		email, err := ExtractAttributeValue(assertion, attrName)
		if err == nil && email != "" {
			return email, nil
		}
	}

	if assertion.Subject != nil && assertion.Subject.NameID != nil && assertion.Subject.NameID.Value != "" {
		return assertion.Subject.NameID.Value, nil
	}

	return "", fmt.Errorf("could not extract email from assertion")
}

func ExtractEmailDomain(email string) (string, error) {
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

func MapSAMLRoleToSystemRole(samlRole string) string {
	if samlRole != "" && isValidRole(samlRole) {
		return samlRole
	}

	return "MEMBER"
}

func isValidRole(role string) bool {
	switch role {
	case "OWNER", "ADMIN", "MEMBER", "VIEWER":
		return true
	default:
		return false
	}
}

func ExtractUserAttributes(
	assertion *saml.Assertion,
	attributeEmail, attributeFirstname, attributeLastname, attributeRole string,
) (email, fullname, role string, err error) {
	if len(assertion.AttributeStatements) == 0 {
		if assertion.Subject != nil && assertion.Subject.NameID != nil {
			email = assertion.Subject.NameID.Value
			fullname = email
			role = ""
			return email, fullname, role, nil
		}
		return "", "", "", fmt.Errorf("no attribute statement and no NameID in assertion")
	}

	email, err = ExtractAttributeValue(assertion, attributeEmail)
	if err != nil {
		if assertion.Subject != nil && assertion.Subject.NameID != nil {
			email = assertion.Subject.NameID.Value
		} else {
			return "", "", "", fmt.Errorf("cannot extract email: %w", err)
		}
	}

	firstname, err := ExtractAttributeValue(assertion, attributeFirstname)
	if err != nil {
		firstname = ""
	}

	lastname, err := ExtractAttributeValue(assertion, attributeLastname)
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
		fullname = email
	}

	role, err = ExtractAttributeValue(assertion, attributeRole)
	if err != nil {
		role = ""
	}

	return email, fullname, role, nil
}
