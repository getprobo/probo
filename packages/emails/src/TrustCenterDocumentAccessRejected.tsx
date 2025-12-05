import { Column, Row, Text } from '@react-email/components';
import * as React from 'react';
import EmailLayout, { bodyText, footerText } from './components/EmailLayout';

export const TrustCenterDocumentAccessRejected = () => {
  return (
    <EmailLayout subject={`Trust Center Access Invitation - ${'{{.OrganizationName}}'}`} organizationName={'{{.OrganizationName}}'}>
      <Text style={bodyText}>
        Your access request to the following files in <strong>{'{{.OrganizationName}}'}</strong>'s Trust Center has been rejected:
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
