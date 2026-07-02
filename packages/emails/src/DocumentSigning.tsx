// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

import { Button, Section, Text } from 'react-email';
import * as React from 'react';
import EmailLayout, { bodyText, button, buttonContainer, footerText } from './components/EmailLayout';

const documentList = {
  ...bodyText,
  paddingLeft: '20px',
};

export const DocumentSigning = () => {
  return (
    <EmailLayout subject={'Action Required – Please review and sign {{.OrganizationName}} compliance documents'}>
      <Text style={bodyText}>
        You're receiving this message because your company, <strong>{'{{.OrganizationName}}'}</strong>, has shared compliance documents that require your review and signature:
      </Text>

      <ul style={documentList} dangerouslySetInnerHTML={{ __html: '{{range .Documents}}<li><a href="{{.URL}}">{{.Title}}</a> ({{.Type}})</li>{{end}}' }} />

      <Text style={bodyText}>
        To stay compliant with company policies, please take a moment to review and sign them by clicking the button below:
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
