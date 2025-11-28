import { Typography, List } from 'antd';
import { Link } from 'react-router';
import { DatabaseOutlined } from '@ant-design/icons';

const { Title } = Typography;

const sections = [
  { name: 'Datasources', path: '/page/documentation/datasource', icon: <DatabaseOutlined /> },
];

function DocumentationIndex() {
  return (
    <>
      <Title level={2}>Documentation</Title>
      <List
        bordered
        dataSource={sections}
        renderItem={(item) => (
          <List.Item>
            <Link to={item.path}>{item.icon} {item.name}</Link>
          </List.Item>
        )}
      />
    </>
  );
}

export default DocumentationIndex;
