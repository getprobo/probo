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

import { Button, Section, Text } from 'react-email';
import * as React from 'react';
import EmailLayout, { bodyText, button, buttonContainer, footerText } from './components/EmailLayout';

const documentList = {
  ...bodyText,
  paddingLeft: '20px',
};

export const DocumentApproval = () => {
  return (
    <EmailLayout subject={'Action Required – Please review and approve {{.OrganizationName}} compliance documents'}>
      <Text style={bodyText}>
        You're receiving this message because <strong>{'{{.OrganizationName}}'}</strong> has requested your approval on the following documents:
      </Text>

      <ul style={documentList} dangerouslySetInnerHTML={{ __html: '{{range .Documents}}<li><a href="{{.URL}}">{{.Title}}</a> ({{.Type}})</li>{{end}}' }} />

      <Text style={bodyText}>
        Please take a moment to review them and approve or reject each one by clicking the button below:
      </Text>

      <Section style={buttonContainer}>
        <Button style={button} href={'{{.ApprovalUrl}}'}>
          Review and Approve
        </Button>
      </Section>

      <Text style={bodyText}>
        If you have any questions, please contact your security team.
      </Text>

      <Text style={footerText}>
        This process is managed securely by Probo, acting as the compliance partner on behalf of <strong>{'{{.OrganizationName}}'}</strong>.
        Thank you for your prompt attention to this matter.
      </Text>
      <Text style={footerText}>
        Best regards,
      </Text>
    </EmailLayout>
  );
};

export default DocumentApproval;
