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

import { Button, Section, Text } from "react-email";
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
