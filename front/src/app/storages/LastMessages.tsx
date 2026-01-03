import { useState, useEffect, useCallback } from 'react';
import {
  Typography,
  Space,
  Button,
  Table,
  message,
  Select,
  Empty,
  Drawer,
  Tag,
  Modal,
  Form,
  InputNumber,
  Input,
  Alert,
} from 'antd';
import {
  ReloadOutlined,
  EyeOutlined,
  MailOutlined,
  CloudDownloadOutlined,
  SyncOutlined,
  DatabaseOutlined,
} from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import client from '../../api/client';
import type { components } from '../../api/v1';

const { Title, Text, Paragraph } = Typography;

type NatsMessage = components['schemas']['NatsMessage'];
type Storage = components['schemas']['Storage'];

// Message record from PostgreSQL storage (the actual data)
interface MessageRecord {
  id?: number;
  uuid?: string;
  subject?: string;
  body?: string;
  body_text?: string;
  body_html?: string;
  sender_email?: string;
  sender_name?: string;
  recipient_email?: string;
  recipient_name?: string;
  message_type?: string;
  source?: string;
  created_at?: string;
  received_at?: string;
  [key: string]: unknown;
}

const LIMIT_OPTIONS = [
  { value: 20, label: '20 messages' },
  { value: 50, label: '50 messages' },
  { value: 100, label: '100 messages' },
  { value: 200, label: '200 messages' },
  { value: 500, label: '500 messages' },
];

function formatDate(dateStr?: string): string {
  if (!dateStr) return '-';
  const date = new Date(dateStr);
  return date.toLocaleString();
}

function formatTimeAgo(dateStr?: string): string {
  if (!dateStr) return '-';
  const date = new Date(dateStr);
  const now = Date.now();
  const diffMs = now - date.getTime();
  const diffSecs = Math.floor(diffMs / 1000);

  if (diffSecs < 60) return `${diffSecs}s ago`;
  const diffMins = Math.floor(diffSecs / 60);
  if (diffMins < 60) return `${diffMins}m ago`;
  const diffHours = Math.floor(diffMins / 60);
  if (diffHours < 24) return `${diffHours}h ${diffMins % 60}m ago`;
  const diffDays = Math.floor(diffHours / 24);
  return `${diffDays}d ago`;
}

function truncateText(text: string | undefined, maxLen: number): string {
  if (!text) return '';
  if (text.length <= maxLen) return text;
  return text.substring(0, maxLen) + '...';
}

function LastMessages() {
  const [messages, setMessages] = useState<NatsMessage[]>([]);
  const [loading, setLoading] = useState(true);
  const [limit, setLimit] = useState(50);
  const [selectedMessage, setSelectedMessage] = useState<NatsMessage | null>(null);
  const [drawerVisible, setDrawerVisible] = useState(false);

  // Fetch modal state
  const [fetchModalVisible, setFetchModalVisible] = useState(false);
  const [storages, setStorages] = useState<Storage[]>([]);
  const [selectedStorage, setSelectedStorage] = useState<Storage | null>(null);
  const [loadingStorages, setLoadingStorages] = useState(false);
  const [fetching, setFetching] = useState(false);
  const [fetchForm] = Form.useForm();

  const loadMessages = useCallback(async () => {
    setLoading(true);
    const { data, error } = await client.GET('/nats/messages', {
      params: { query: { limit } },
    });
    if (error) {
      message.error('Failed to load messages');
      setLoading(false);
      return;
    }
    setMessages(data?.messages || []);
    setLoading(false);
  }, [limit]);

  const loadStorages = useCallback(async () => {
    setLoadingStorages(true);
    const { data, error } = await client.GET('/storage');
    if (error) {
      message.error('Failed to load storages');
      setLoadingStorages(false);
      return;
    }
    // API returns array directly
    setStorages(data || []);
    setLoadingStorages(false);
  }, []);

  const handleOpenFetchModal = () => {
    loadStorages();
    setSelectedStorage(null);
    fetchForm.resetFields();
    fetchForm.setFieldsValue({ limit: 100, table_name: 'messages' });
    setFetchModalVisible(true);
  };

  const handleStorageChange = (uuid: string) => {
    const storage = storages.find(s => s.uuid === uuid);
    setSelectedStorage(storage || null);
  };

  const handleFetchMessages = async () => {
    try {
      const values = await fetchForm.validateFields();

      if (!selectedStorage) {
        message.error('Please select a storage');
        return;
      }

      // Currently only PostgreSQL supports message query
      if (selectedStorage.type !== 'postgres') {
        message.error(`Message query is not yet supported for ${selectedStorage.type} storage type`);
        return;
      }

      setFetching(true);

      const { error } = await client.POST('/storage/postgres/{uuid}/messages/query', {
        params: { path: { uuid: values.storage_uuid } },
        body: {
          limit: values.limit,
          table_name: values.table_name,
        },
      });

      if (error) {
        message.error('Failed to start message query');
        setFetching(false);
        return;
      }

      message.success(`Query job started. Fetching up to ${values.limit} messages...`);
      setFetchModalVisible(false);
      setFetching(false);

      // Poll for results after a short delay
      setTimeout(() => {
        loadMessages();
      }, 2000);

      // Poll again after more time
      setTimeout(() => {
        loadMessages();
      }, 5000);
    } catch {
      // Validation error
    }
  };

  useEffect(() => {
    loadMessages();
  }, [loadMessages]);

  const handleViewMessage = (record: NatsMessage) => {
    setSelectedMessage(record);
    setDrawerVisible(true);
  };

  // Extract message record from NATS message data
  const getMessageData = (natsMsg: NatsMessage): MessageRecord => {
    return (natsMsg.data as MessageRecord) || {};
  };

  const columns: ColumnsType<NatsMessage> = [
    {
      title: 'Date',
      key: 'date',
      width: 140,
      render: (_, record) => {
        const msg = getMessageData(record);
        const dateStr = msg.created_at || msg.received_at;
        return (
          <Text title={formatDate(dateStr)} style={{ fontSize: 12 }}>
            {formatTimeAgo(dateStr)}
          </Text>
        );
      },
      sorter: (a, b) => {
        const dateA = getMessageData(a).created_at || '';
        const dateB = getMessageData(b).created_at || '';
        return dateA.localeCompare(dateB);
      },
      defaultSortOrder: 'descend',
    },
    {
      title: 'From',
      key: 'sender',
      width: 200,
      ellipsis: true,
      render: (_, record) => {
        const msg = getMessageData(record);
        const name = msg.sender_name || msg.sender_email || '-';
        return (
          <Text title={msg.sender_email}>
            {truncateText(name, 30)}
          </Text>
        );
      },
    },
    {
      title: 'Subject',
      key: 'subject',
      ellipsis: true,
      render: (_, record) => {
        const msg = getMessageData(record);
        return (
          <Text strong>
            {msg.subject || '(no subject)'}
          </Text>
        );
      },
    },
    {
      title: 'Type',
      key: 'type',
      width: 100,
      render: (_, record) => {
        const msg = getMessageData(record);
        const type = msg.message_type || msg.source || 'email';
        const colorMap: Record<string, string> = {
          email: 'blue',
          telegram: 'cyan',
          whatsapp: 'green',
          linkedin: 'geekblue',
        };
        return <Tag color={colorMap[type] || 'default'}>{type}</Tag>;
      },
    },
    {
      title: '',
      key: 'actions',
      width: 50,
      render: (_, record) => (
        <Button
          type="text"
          icon={<EyeOutlined />}
          onClick={() => handleViewMessage(record)}
          title="View Message"
        />
      ),
    },
  ];

  const selectedMsgData = selectedMessage ? getMessageData(selectedMessage) : null;

  return (
    <>
      <Space style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
        <Space>
          <Title level={4} style={{ margin: 0 }}>Last Messages</Title>
          <Select
            value={limit}
            onChange={setLimit}
            options={LIMIT_OPTIONS}
            style={{ width: 140 }}
          />
          <Text type="secondary">
            {messages.length} message{messages.length !== 1 ? 's' : ''} fetched
          </Text>
        </Space>
        <Space>
          <Button
            type="primary"
            icon={<CloudDownloadOutlined />}
            onClick={handleOpenFetchModal}
          >
            Fetch from Storage
          </Button>
          <Button icon={<ReloadOutlined />} onClick={loadMessages}>
            Refresh
          </Button>
        </Space>
      </Space>

      <Table
        columns={columns}
        dataSource={messages}
        rowKey="sequence"
        loading={loading}
        pagination={{ pageSize: 20, showSizeChanger: true, pageSizeOptions: ['10', '20', '50', '100'] }}
        locale={{
          emptyText: (
            <Empty
              description="No messages fetched yet. Run a message query from a PostgreSQL storage to see data here."
              image={Empty.PRESENTED_IMAGE_SIMPLE}
            />
          ),
        }}
        size="small"
      />

      <Drawer
        title={
          <Space>
            <MailOutlined />
            {selectedMsgData?.subject || 'Message Details'}
          </Space>
        }
        placement="right"
        width={700}
        onClose={() => setDrawerVisible(false)}
        open={drawerVisible}
      >
        {selectedMsgData && (
          <Space direction="vertical" style={{ width: '100%' }} size="middle">
            {/* Header info */}
            <div style={{ borderBottom: '1px solid #f0f0f0', paddingBottom: 16 }}>
              <Space direction="vertical" size="small" style={{ width: '100%' }}>
                {selectedMsgData.sender_email && (
                  <div>
                    <Text type="secondary">From: </Text>
                    <Text strong>
                      {selectedMsgData.sender_name && `${selectedMsgData.sender_name} `}
                      &lt;{selectedMsgData.sender_email}&gt;
                    </Text>
                  </div>
                )}
                {selectedMsgData.recipient_email && (
                  <div>
                    <Text type="secondary">To: </Text>
                    <Text>
                      {selectedMsgData.recipient_name && `${selectedMsgData.recipient_name} `}
                      &lt;{selectedMsgData.recipient_email}&gt;
                    </Text>
                  </div>
                )}
                {(selectedMsgData.created_at || selectedMsgData.received_at) && (
                  <div>
                    <Text type="secondary">Date: </Text>
                    <Text>{formatDate(selectedMsgData.created_at || selectedMsgData.received_at)}</Text>
                  </div>
                )}
                {selectedMsgData.subject && (
                  <div>
                    <Text type="secondary">Subject: </Text>
                    <Text strong>{selectedMsgData.subject}</Text>
                  </div>
                )}
              </Space>
            </div>

            {/* Message body */}
            <div>
              <Text type="secondary" style={{ display: 'block', marginBottom: 8 }}>Message Body</Text>
              {selectedMsgData.body_html ? (
                <div
                  style={{
                    backgroundColor: '#fff',
                    border: '1px solid #f0f0f0',
                    borderRadius: 8,
                    padding: 16,
                    maxHeight: 400,
                    overflow: 'auto',
                  }}
                  dangerouslySetInnerHTML={{ __html: selectedMsgData.body_html }}
                />
              ) : (
                <Paragraph
                  style={{
                    backgroundColor: '#fafafa',
                    padding: 16,
                    borderRadius: 8,
                    whiteSpace: 'pre-wrap',
                    maxHeight: 400,
                    overflow: 'auto',
                  }}
                >
                  {selectedMsgData.body_text || selectedMsgData.body || '(no body)'}
                </Paragraph>
              )}
            </div>

            {/* Raw data */}
            <div>
              <Text type="secondary" style={{ display: 'block', marginBottom: 8 }}>Raw Data</Text>
              <pre style={{
                backgroundColor: '#f5f5f5',
                padding: 16,
                borderRadius: 8,
                overflow: 'auto',
                maxHeight: 300,
                fontSize: 11,
              }}>
                {JSON.stringify(selectedMsgData, null, 2)}
              </pre>
            </div>
          </Space>
        )}
      </Drawer>

      {/* Fetch from Storage Modal */}
      <Modal
        title={
          <Space>
            <DatabaseOutlined />
            Fetch Messages from Storage
          </Space>
        }
        open={fetchModalVisible}
        onCancel={() => setFetchModalVisible(false)}
        onOk={handleFetchMessages}
        okText={fetching ? 'Fetching...' : 'Fetch Messages'}
        okButtonProps={{
          loading: fetching,
          icon: fetching ? <SyncOutlined spin /> : <CloudDownloadOutlined />,
          disabled: !selectedStorage || selectedStorage.type !== 'postgres',
        }}
        width={500}
      >
        <Alert
          message="Query messages from a data storage and stream them to this page."
          type="info"
          showIcon
          style={{ marginBottom: 16 }}
        />
        <Form form={fetchForm} layout="vertical">
          <Form.Item
            name="storage_uuid"
            label="Data Storage"
            rules={[{ required: true, message: 'Please select a storage' }]}
          >
            <Select
              loading={loadingStorages}
              placeholder="Select a data storage"
              onChange={handleStorageChange}
              notFoundContent={loadingStorages ? 'Loading...' : 'No storages found'}
            >
              {storages.map(s => (
                <Select.Option key={s.uuid} value={s.uuid}>
                  <Space>
                    <Tag color={s.type === 'postgres' ? 'blue' : s.type === 's3' ? 'orange' : 'default'}>
                      {s.type}
                    </Tag>
                    {s.name || s.uuid}
                  </Space>
                </Select.Option>
              ))}
            </Select>
          </Form.Item>

          {selectedStorage && selectedStorage.type !== 'postgres' && (
            <Alert
              message={`Message query is not yet supported for ${selectedStorage.type} storage type. Only PostgreSQL is currently supported.`}
              type="warning"
              showIcon
              style={{ marginBottom: 16 }}
            />
          )}

          {selectedStorage?.type === 'postgres' && (
            <>
              <Form.Item
                name="table_name"
                label="Table Name"
                rules={[{ required: true, message: 'Please enter table name' }]}
                initialValue="messages"
              >
                <Input placeholder="messages" />
              </Form.Item>
              <Form.Item
                name="limit"
                label="Maximum Messages"
                rules={[{ required: true, message: 'Please enter limit' }]}
                initialValue={100}
              >
                <InputNumber min={1} max={1000} style={{ width: '100%' }} />
              </Form.Item>
            </>
          )}
        </Form>
      </Modal>
    </>
  );
}

export default LastMessages;
