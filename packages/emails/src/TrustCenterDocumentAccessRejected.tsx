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
