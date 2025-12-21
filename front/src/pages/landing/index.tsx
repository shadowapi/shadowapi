import { Typography, Card, Flex, Button, Space } from 'antd';
import { Link } from 'react-router';
import { RocketOutlined } from '@ant-design/icons';
import { uiColors } from '../../theme';
import { SmartLink } from '../../lib/SmartLink';

const { Title, Text, Paragraph } = Typography;

function LandingPage() {
  return (
    <Flex justify="center" align="center" style={{ minHeight: '60vh' }}>
      <Card style={{ width: 500, textAlign: 'center' }}>
        <RocketOutlined style={{ fontSize: 64, color: uiColors.primary, marginBottom: 24 }} />
        <Title level={1} style={{ marginBottom: 8 }}>MeshPump</Title>
        <Text type="secondary" style={{ fontSize: 16 }}>
          Unified Messaging Platform
        </Text>
        <Paragraph style={{ marginTop: 24, fontSize: 16 }}>
          Normalize Gmail, Telegram, WhatsApp, and LinkedIn into a single REST + MCP surface.
          Connect all your messaging channels in one place.
        </Paragraph>
        <Space direction="vertical" size="middle" style={{ width: '100%', marginTop: 32 }}>
          <SmartLink to="/login">
            <Button type="primary" size="large" block>
              Get Started
            </Button>
          </SmartLink>
          <Link to="/documentation">
            <Button type="default" size="large" block>
              Documentation
            </Button>
          </Link>
        </Space>
        <Paragraph type="secondary" style={{ marginTop: 32, fontSize: 12 }}>
          Built with Go, React, and PostgreSQL
        </Paragraph>
      </Card>
    </Flex>
  );
}

export default LandingPage;
