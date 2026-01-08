import { Typography, Button, Space, Card, Row, Col } from 'antd';
import {
  MailOutlined,
  SendOutlined,
  MessageOutlined,
  UserOutlined,
} from '@ant-design/icons';
import { Link } from 'react-router';
import { SmartLink } from '../../lib/SmartLink';
import { colors } from '../../theme';

const { Title, Text, Paragraph } = Typography;

const integrations = [
  {
    key: 'gmail',
    icon: <MailOutlined />,
    title: 'Gmail',
    description: 'OAuth2 integration with Google Workspace. Sync emails, threads, and attachments.',
  },
  {
    key: 'telegram',
    icon: <SendOutlined />,
    title: 'Telegram',
    description: 'Connect to Telegram bots and channels. Real-time message sync.',
  },
  {
    key: 'whatsapp',
    icon: <MessageOutlined />,
    title: 'WhatsApp',
    description: 'WhatsApp Business API integration. Unified conversation management.',
  },
  {
    key: 'linkedin',
    icon: <UserOutlined />,
    title: 'LinkedIn',
    description: 'Professional messaging integration. Connect with your network.',
  },
];

function LandingPage() {
  return (
    <>
      {/* Hero Section */}
      <section
        style={{
          background: colors.oxfordBlue,
          padding: '80px 24px',
          textAlign: 'center',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          minHeight: '60vh',
        }}
      >
        <div style={{ maxWidth: 800 }}>
          <img
            src="/logo.svg"
            alt="MeshPump Logo"
            style={{ width: 120, height: 120, marginBottom: 24 }}
          />
          <Title
            level={1}
            style={{
              color: colors.white,
              marginBottom: 8,
              fontSize: 48,
            }}
          >
            MeshPump
          </Title>
          <Title
            level={2}
            style={{
              color: colors.orange,
              marginBottom: 24,
              fontWeight: 400,
              marginTop: 0,
            }}
          >
            Unified Messaging Platform
          </Title>
          <Paragraph
            style={{
              color: colors.lightGray,
              fontSize: 18,
              maxWidth: 600,
              margin: '0 auto 40px',
            }}
          >
            Normalize Gmail, Telegram, WhatsApp, and LinkedIn into a single REST + MCP surface.
            Connect all your messaging channels in one place.
          </Paragraph>
          <Space size="large">
            <SmartLink to="/login">
              <Button type="primary" size="large">
                Get Started
              </Button>
            </SmartLink>
            <Link to="/documentation">
              <Button
                size="large"
                ghost
                style={{ borderColor: colors.white, color: colors.white }}
              >
                Documentation
              </Button>
            </Link>
          </Space>
        </div>
      </section>

      {/* Features Section */}
      <section style={{ padding: '80px 24px', background: colors.white }}>
        <div style={{ maxWidth: 1200, margin: '0 auto' }}>
          <Title level={2} style={{ textAlign: 'center', marginBottom: 16 }}>
            Connect All Your Messaging Channels
          </Title>
          <Paragraph
            style={{
              textAlign: 'center',
              marginBottom: 48,
              color: 'rgba(0, 0, 0, 0.65)',
              fontSize: 16,
            }}
          >
            One API to rule them all. MeshPump normalizes disparate messaging platforms into a unified interface.
          </Paragraph>

          <Row gutter={[24, 24]} justify="center">
            {integrations.map((feature) => (
              <Col key={feature.key} xs={24} sm={12} lg={6}>
                <Card hoverable style={{ textAlign: 'center', height: '100%' }}>
                  <div
                    style={{
                      width: 64,
                      height: 64,
                      borderRadius: '50%',
                      background: `${colors.orange}20`,
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      margin: '0 auto 16px',
                      fontSize: 28,
                      color: colors.orange,
                    }}
                  >
                    {feature.icon}
                  </div>
                  <Title level={4} style={{ marginBottom: 8 }}>
                    {feature.title}
                  </Title>
                  <Text type="secondary">{feature.description}</Text>
                </Card>
              </Col>
            ))}
          </Row>
        </div>
      </section>
    </>
  );
}

export default LandingPage;
