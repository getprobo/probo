// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

import { Button, Section, Text } from "@react-email/components";
import * as React from "react";
import EmailLayout, {
  bodyText,
  button,
  buttonContainer,
  footerText,
} from "./components/EmailLayout";

export const TrustCenterAccess = () => {
  return (
    <EmailLayout
      subject={`Compliance Page Access Invitation - ${"{{.OrganizationName}}"}`}
    >
      <Text style={bodyText}>
        You have been granted access to{" "}
        <strong>{"{{.OrganizationName}}"}</strong>'s compliance page! Click the
        button below to access it:
      </Text>

      <Section style={buttonContainer}>
        <Button style={button} href={"{{.BaseURL}}"}>
          Access Compliance Page
        </Button>
      </Section>
    </EmailLayout>
  );
};

export default TrustCenterAccess;
