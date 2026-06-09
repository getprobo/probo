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

import { Button, Section, Text } from "@react-email/components";
import * as React from "react";
import EmailLayout, {
  bodyText,
  button,
  buttonContainer,
} from "./components/EmailLayout";

export const MagicLink = () => {
  return (
    <EmailLayout subject="Probo Magic Link">
      <Text style={bodyText}>{"Please use this link to connect to {{.OrganizationName}}'s Compliance Page:"}</Text>

      <Section style={buttonContainer}>
        <Button style={button} href={"{{.MagicLinkURL}}"}>
          Connect
        </Button>
      </Section>

      <Section style={expiryBox}>
        <Text style={expiryText}>
          {"⚠️ This link expires in {{.DurationInMinutes}} minutes. Use it promptly after receiving this email."}
        </Text>
      </Section>
    </EmailLayout>
  );
};

const expiryBox: React.CSSProperties = {
  backgroundColor: "#fff8e1",
  border: "1px solid #f9a825",
  borderRadius: "6px",
  padding: "12px 16px",
  marginTop: "8px",
};

const expiryText: React.CSSProperties = {
  margin: "0",
  color: "#7a5900",
  fontSize: "14px",
  fontWeight: "600",
  lineHeight: "20px",
};

export default MagicLink;
