import { Typography, Card, Flex } from 'antd';
import { Link } from 'react-router';
import { DatabaseOutlined } from '@ant-design/icons';

const { Title } = Typography;

const sections = [
  { name: 'Datasources', path: '/documentation/datasource', icon: <DatabaseOutlined /> },
];

function DocumentationIndex() {
  return (
    <>
      <Title level={2}>Documentation</Title>
      <Flex vertical gap="small">
        {sections.map((item) => (
          <Card key={item.path} size="small" hoverable>
            {/* Using Link here is fine - navigation within /page/* stays as SPA */}
            <Link to={item.path}>{item.icon} {item.name}</Link>
          </Card>
        ))}
      </Flex>
    </>
  );
}

export default DocumentationIndex;
