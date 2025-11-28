import { Typography, List } from 'antd';
import { Link } from 'react-router';
import { MailOutlined, SendOutlined } from '@ant-design/icons';

const { Title } = Typography;

const datasources = [
  { name: 'Gmail', path: '/page/documentation/datasource/gmail', icon: <MailOutlined /> },
  { name: 'Telegram', path: '/page/documentation/datasource/telegram', icon: <SendOutlined /> },
];

function DatasourceIndex() {
  return (
    <>
      <Title level={2}>Datasources</Title>
      <List
        bordered
        dataSource={datasources}
        renderItem={(item) => (
          <List.Item>
            <Link to={item.path}>{item.icon} {item.name}</Link>
          </List.Item>
        )}
      />
    </>
  );
}

export default DatasourceIndex;
