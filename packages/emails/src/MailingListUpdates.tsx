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

import { Button, Hr, Link, Section, Text } from "react-email";
import * as React from "react";
import EmailLayout, {
  bodyText,
  button,
  buttonContainer,
  footerText,
} from "./components/EmailLayout";

export const MailingListUpdates = () => {
  return (
    <EmailLayout
      subject={`${"{{.OrganizationName}}"} – ${"{{.NewsTitle}}"}`}
    >
      <Text style={bodyText}>
        {"{{.OrganizationName}}"} has published a new compliance update.
      </Text>

      <Text style={{ ...bodyText, fontWeight: "600", fontSize: "18px" }}>
        {"{{.NewsTitle}}"}
      </Text>

      <Text style={bodyText}>{"{{.NewsBody}}"}</Text>

      <Section style={buttonContainer}>
        <Button style={button} href={"{{.CompliancePageURL}}"}>
          View Compliance Page
        </Button>
      </Section>

      <Hr style={{ borderColor: "#ecefec", margin: "8px 0 20px" }} />

      <Text style={{ ...footerText, textAlign: "center" }}>
        You are receiving this email because you subscribed to compliance
        updates from <strong>{"{{.OrganizationName}}"}</strong>. To
        unsubscribe,{" "}
        <Link
          href={"{{.UnsubscribeURL}}"}
          style={{ color: "#374151", textDecoration: "underline", fontWeight: "500" }}
        >
          click here.
        </Link>
      </Text>
    </EmailLayout>
  );
};

export default MailingListUpdates;
