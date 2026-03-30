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

import { Text } from '@react-email/components';
import * as React from 'react';
import EmailLayout, { bodyText } from './components/EmailLayout';

export const TrustCenterDocumentAccessRejected = () => {
  return (
    <EmailLayout subject={`Compliance Page Document Access Rejected - ${'{{.OrganizationName}}'}`}>
      <Text style={bodyText}>
        Your access request to the following files in <strong>{'{{.OrganizationName}}'}</strong>'s compliance page has been rejected:
      </Text>

      <Text  style={bodyText}>
      {'{{range .FileNames}}'}
        • {'{{.}}'}<br/>
      {'{{end}}'}
      </Text>
    </EmailLayout>
  );
};

export default TrustCenterDocumentAccessRejected;
