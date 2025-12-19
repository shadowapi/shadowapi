import { Typography, Card, Flex } from 'antd';
import { Link } from 'react-router';
import { MailOutlined, SendOutlined } from '@ant-design/icons';

const { Title } = Typography;

const datasources = [
  { name: 'Gmail', path: '/documentation/datasource/gmail', icon: <MailOutlined /> },
  { name: 'Telegram', path: '/documentation/datasource/telegram', icon: <SendOutlined /> },
];

function DatasourceIndex() {
  return (
    <>
      <Title level={2}>Datasources</Title>
      <Flex vertical gap="small">
        {datasources.map((item) => (
          <Card key={item.path} size="small" hoverable>
            <Link to={item.path}>{item.icon} {item.name}</Link>
          </Card>
        ))}
      </Flex>
    </>
  );
}

export default DatasourceIndex;
