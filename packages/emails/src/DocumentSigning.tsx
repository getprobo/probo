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

import { Button, Section, Text } from '@react-email/components';
import * as React from 'react';
import EmailLayout, { bodyText, button, buttonContainer, footerText } from './components/EmailLayout';

export const DocumentSigning = () => {
  return (
    <EmailLayout subject={'Action Required – Please review and sign {{.OrganizationName}} compliance documents'}>
      <Text style={bodyText}>
        You're receiving this message because your company, <strong>{'{{.OrganizationName}}'}</strong>, has shared a new compliance document that requires your review and signature.
      </Text>
      <Text style={bodyText}>
        To stay compliant with company policies, please take a moment to review and sign the document by clicking the button below:
      </Text>

      <Section style={buttonContainer}>
        <Button style={button} href={'{{.SigningUrl}}'}>
          Review and Sign Documents
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

export default DocumentSigning;
