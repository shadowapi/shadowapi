import { Typography, Alert, Steps } from 'antd';

const { Title, Paragraph, Text, Link } = Typography;

function GmailDocumentation() {
  return (
    <>
      <Title level={2}>Gmail Datasource</Title>
      <Paragraph>
        This guide explains how to configure Gmail as a datasource in ShadowAPI. It is based on
        the official{' '}
        <Link href="https://developers.google.com/gmail/api/guides" target="_blank">
          Google Gmail API documentation
        </Link>
        .
      </Paragraph>

      <Title level={3}>Prerequisites</Title>
      <Paragraph>
        To enable the Gmail API, you need a Google Cloud project with the appropriate APIs enabled
        and OAuth credentials configured.
      </Paragraph>

      <Title level={3}>Setup Steps</Title>

      <Title level={4}>1. Create a Google Cloud Project</Title>
      <Paragraph>
        If you don't already have a Google Cloud project, create one by following the{' '}
        <Link href="https://developers.google.com/workspace/guides/create-project" target="_blank">
          Create a Google Cloud project
        </Link>{' '}
        guide.
      </Paragraph>

      <Title level={4}>2. Enable the Gmail API</Title>
      <Steps
        orientation="vertical"
        size="small"
        current={-1}
        items={[
          {
            title: 'Open the API Library',
            content: (
              <>
                Navigate to the{' '}
                <Link
                  href="https://console.cloud.google.com/workspace-api/products"
                  target="_blank"
                >
                  Google Workspace API Products
                </Link>{' '}
                page in your Google Cloud Console.
              </>
            ),
          },
          {
            title: 'Find Gmail API',
            content: 'Search for "Gmail API" in the product library.',
          },
          {
            title: 'Enable the API',
            content: 'Click on the Gmail API and then click the "Enable" button.',
          },
        ]}
      />

      <Title level={4}>3. Configure OAuth Consent Screen</Title>
      <Steps
        orientation="vertical"
        size="small"
        current={-1}
        items={[
          {
            title: 'Open OAuth Consent Screen',
            content: (
              <>
                Go to the{' '}
                <Link
                  href="https://console.cloud.google.com/apis/credentials/consent"
                  target="_blank"
                >
                  OAuth consent screen
                </Link>{' '}
                in your Google Cloud Console.
              </>
            ),
          },
          {
            title: 'Select User Type',
            content: (
              <>
                For <Text strong>User type</Text>, select <Text strong>Internal</Text> if you are
                using Google Workspace. For personal Gmail accounts, select{' '}
                <Text strong>External</Text>.
              </>
            ),
          },
          {
            title: 'Complete Registration',
            content:
              'Fill in the required application information (app name, user support email, developer contact).',
          },
          {
            title: 'Add Scopes',
            content:
              'Add the necessary Gmail API scopes for your use case (e.g., gmail.readonly, gmail.send).',
          },
        ]}
      />

      <Title level={4}>4. Create OAuth Credentials</Title>
      <Steps
        orientation="vertical"
        size="small"
        current={-1}
        items={[
          {
            title: 'Navigate to Credentials',
            content: (
              <>
                Go to{' '}
                <Link href="https://console.cloud.google.com/apis/credentials" target="_blank">
                  APIs & Services &gt; Credentials
                </Link>
                .
              </>
            ),
          },
          {
            title: 'Create Credentials',
            content:
              'Click "Create Credentials" and select "OAuth client ID". Choose "Web application" as the application type.',
          },
          {
            title: 'Configure Redirect URIs',
            content:
              'Add the ShadowAPI callback URL as an authorized redirect URI (check your ShadowAPI configuration for the exact URL).',
          },
          {
            title: 'Download Credentials',
            content: (
              <>
                Download the <Text code>credentials.json</Text> file and store it in a secure
                location. You will need the Client ID and Client Secret for ShadowAPI
                configuration.
              </>
            ),
          },
        ]}
      />

      <Alert
        type="warning"
        showIcon
        title="Security Notice"
        description={
          <>
            Never commit your <Text code>credentials.json</Text> file or OAuth secrets to version
            control. Store them securely and use environment variables or secret management
            solutions in production.
          </>
        }
        style={{ marginTop: 16, marginBottom: 16 }}
      />

      <Title level={3}>Configuring ShadowAPI</Title>
      <Paragraph>
        Once you have your OAuth credentials, configure the Gmail datasource in ShadowAPI by
        providing:
      </Paragraph>
      <ul>
        <li>
          <Text strong>Client ID</Text> - From your OAuth credentials
        </li>
        <li>
          <Text strong>Client Secret</Text> - From your OAuth credentials
        </li>
        <li>
          <Text strong>Redirect URI</Text> - Must match the one configured in Google Cloud Console
        </li>
      </ul>

      <Title level={3}>Troubleshooting</Title>
      <Paragraph>
        <Text strong>Error: "Access blocked"</Text> - Ensure your OAuth consent screen is properly
        configured and the user has been added as a test user (for apps in testing mode).
      </Paragraph>
      <Paragraph>
        <Text strong>Error: "Redirect URI mismatch"</Text> - Verify that the redirect URI in
        ShadowAPI exactly matches the one configured in your Google Cloud OAuth credentials.
      </Paragraph>
    </>
  );
}

export default GmailDocumentation;
