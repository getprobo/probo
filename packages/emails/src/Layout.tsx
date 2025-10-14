import {
  Body,
  Button,
  Container,
  Head,
  Heading,
  Html,
  Link,
  Preview,
  Section,
  Text,
} from '@react-email/components';
import * as React from 'react';

interface LayoutProps {
  subject: string;
  header: string;
  fullName: string;
  body: string;
  buttonText?: string;
  buttonURL?: string;
  footer?: string;
}

export const Layout = ({
  subject,
  header,
  fullName,
  body,
  buttonText,
  buttonURL,
  footer,
}: LayoutProps) => {
  return (
    <Html lang="en">
      <Head>
        <meta name="x-apple-disable-message-reformatting" />
        <meta httpEquiv="X-UA-Compatible" content="IE=edge" />
        <title>{subject}</title>
      </Head>
      <Preview>{header}</Preview>
      <Body style={main}>
        <Container style={container}>
          <Section style={headerSection}>
            <Heading style={h1}>{header}</Heading>
          </Section>

          <Section style={content}>
            <Text style={text}>Hi {fullName},</Text>

            <Text style={textWithMargin}>{body}</Text>

            <span dangerouslySetInnerHTML={{ __html: '<!-- CONDITIONAL_BUTTON_START -->' }} />
            <Section style={buttonContainer}>
              <Button style={button} href={buttonURL}>
                {buttonText}
              </Button>
            </Section>
            <span dangerouslySetInnerHTML={{ __html: '<!-- CONDITIONAL_BUTTON_END -->' }} />

            <span dangerouslySetInnerHTML={{ __html: '<!-- CONDITIONAL_FOOTER_START -->' }} />
            <Text style={footerText}>{footer}</Text>
            <span dangerouslySetInnerHTML={{ __html: '<!-- CONDITIONAL_FOOTER_END -->' }} />
          </Section>

          <Section style={footerSection}>
            <Text style={footerSignature}>
              Thanks,
              <br />
              Probo Team
            </Text>
            <Text style={poweredBy}>
              <Link href="https://www.getprobo.com" style={poweredByLink}>
                Powered by{' '}
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg" style={{ display: 'inline-block', verticalAlign: 'middle' }}>
                  <rect width="24" height="24" fill="#6b716a" rx="5.14"></rect>
                  <path stroke="#fffffb" strokeWidth="3.53" d="M20.57 6.17c2.96 4.47-6.43 13.97-20.32 5.66"></path>
                  <path fill="#fffffb" d="M19.07 7.1a1.77 1.77 0 0 0 3-1.86l-1.5.93-1.5.94ZM8.82 25.18l1.71-.41c-2.3-9.67.01-14.66 2.54-16.83A6.25 6.25 0 0 1 17 6.33c1.22-.01 1.87.44 2.08.78l1.5-.94 1.5-.93c-1.1-1.74-3.16-2.46-5.13-2.44-2.03.03-4.26.81-6.17 2.45-3.9 3.35-6.16 9.92-3.67 20.33l1.72-.41Z"></path>
                </svg>
                {' '}Probo
              </Link>
            </Text>
          </Section>
        </Container>
      </Body>
    </Html>
  );
};

export default Layout;

const main: React.CSSProperties = {
  margin: '0',
  padding: '0',
  fontFamily:
    "-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif",
  backgroundColor: '#fffffb',
  WebkitTextSizeAdjust: '100%',
};

const container: React.CSSProperties = {
  maxWidth: '600px',
  width: '100%',
  backgroundColor: '#ffffff',
  borderRadius: '8px',
  boxShadow: '0 2px 4px rgba(0, 0, 0, 0.1)',
  margin: '40px auto',
};

const headerSection: React.CSSProperties = {
  padding: '40px 40px 30px 40px',
  textAlign: 'center',
  backgroundColor: '#044E4114',
  borderRadius: '8px 8px 0 0',
};

const h1: React.CSSProperties = {
  margin: '0',
  color: '#141e12',
  fontSize: '24px',
  fontWeight: '600',
  lineHeight: '1.3',
};

const content: React.CSSProperties = {
  padding: '40px',
};

const text: React.CSSProperties = {
  margin: '0 0 20px 0',
  color: '#141e12',
  fontSize: '16px',
  lineHeight: '24px',
};

const textWithMargin: React.CSSProperties = {
  margin: '0 0 30px 0',
  color: '#141e12',
  fontSize: '16px',
  lineHeight: '24px',
};

const buttonContainer: React.CSSProperties = {
  padding: '10px 0 30px 0',
  textAlign: 'center',
};

const button: React.CSSProperties = {
  display: 'inline-block',
  padding: '14px 28px',
  backgroundColor: '#93c926',
  color: '#141e12',
  textDecoration: 'none',
  borderRadius: '6px',
  fontWeight: '600',
  fontSize: '16px',
  lineHeight: '20px',
};

const footerText: React.CSSProperties = {
  margin: '0',
  color: '#6b716a',
  fontSize: '14px',
  lineHeight: '20px',
};

const footerSection: React.CSSProperties = {
  padding: '30px 40px',
  textAlign: 'center',
  borderTop: '1px solid #ecefec',
};

const footerSignature: React.CSSProperties = {
  margin: '0 0 10px 0',
  color: '#141e12',
  fontSize: '14px',
  lineHeight: '20px',
};

const poweredBy: React.CSSProperties = {
  margin: '20px 0 0 0',
  fontSize: '12px',
  lineHeight: '18px',
};

const poweredByLink: React.CSSProperties = {
  color: '#6b716a',
  textDecoration: 'none',
};
