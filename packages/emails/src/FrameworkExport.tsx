import { Button, Section, Text } from '@react-email/components';
import * as React from 'react';
import EmailLayout, { bodyText, button, buttonContainer, footerText } from './components/EmailLayout';

export const FrameworkExport = () => {
  return (
    <EmailLayout subject="Your framework export is ready">
      <Text style={bodyText}>
        Your framework export has been completed successfully. Click the button below to download it:
      </Text>

      <Section style={buttonContainer}>
        <Button style={button} href={'{{.DownloadUrl}}'}>
          Download Export
        </Button>
      </Section>

      <Text style={footerText}>
        This link will expire in 24 hours.
      </Text>
    </EmailLayout>
  );
};

export default FrameworkExport;
