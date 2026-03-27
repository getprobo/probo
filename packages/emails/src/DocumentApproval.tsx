import { Button, Section, Text } from '@react-email/components';
import * as React from 'react';
import EmailLayout, { bodyText, button, buttonContainer, footerText } from './components/EmailLayout';

export const DocumentApproval = () => {
  return (
    <EmailLayout subject={'Action Required – Please review and approve {{.DocumentName}}'}>
      <Text style={bodyText}>
        You're receiving this message because <strong>{'{{.OrganizationName}}'}</strong> has requested your approval on the document <strong>{'{{.DocumentName}}'}</strong>.
      </Text>
      <Text style={bodyText}>
        Please take a moment to review the document and approve or reject it by clicking the button below:
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
