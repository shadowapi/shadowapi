import { Typography, Card, Flex } from 'antd';

const { Title, Text, Paragraph } = Typography;

function AboutPage() {
  return (
    <Flex justify="center" align="center" style={{ minHeight: '60vh' }}>
      <Card style={{ width: 400, textAlign: 'center' }}>
        <Title level={2}>ShadowAPI</Title>
        <Text strong style={{ fontSize: 18 }}>Version 0.0.1</Text>
        <Paragraph style={{ marginTop: 24 }}>
          A unified messaging platform that normalizes Gmail, Telegram, WhatsApp,
          and LinkedIn into a single REST + MCP surface.
        </Paragraph>
        <Paragraph type="secondary">
          Built with Go, React, and PostgreSQL.
        </Paragraph>
        <Text type="secondary">© {new Date().getFullYear()} ShadowAPI</Text>
      </Card>
    </Flex>
  );
}

export default AboutPage;
