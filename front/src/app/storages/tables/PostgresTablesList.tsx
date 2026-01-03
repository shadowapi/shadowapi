import { useState, useEffect, useCallback } from 'react';
import { useNavigate, useParams } from 'react-router';
import { Table, Button, Space, Typography, Tag, Popconfirm, message, Empty, Spin, Card, Divider, Alert } from 'antd';
import { PlusOutlined, ArrowLeftOutlined, DeleteOutlined, ImportOutlined, ReloadOutlined, DatabaseOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import { useWorkspace } from '../../../lib/workspace/WorkspaceContext';
import client from '../../../api/client';
import type { components } from '../../../api/v1';

type PostgresTable = components['schemas']['storage_postgres_table'];
type IntrospectTable = components['schemas']['StoragePostgresIntrospectTable'];

const { Title, Text } = Typography;

function PostgresTablesList() {
  const navigate = useNavigate();
  const { uuid: rawUuid } = useParams<{ uuid: string }>();
  const { slug } = useWorkspace();
  const uuid = rawUuid || 'new';
  const isNewStorage = uuid === 'new';

  // Configured tables from storage settings
  const [configuredTables, setConfiguredTables] = useState<PostgresTable[]>([]);
  // Available tables from database introspection
  const [dbTables, setDbTables] = useState<IntrospectTable[]>([]);
  const [loading, setLoading] = useState(true);
  const [introspecting, setIntrospecting] = useState(false);
  const [introspectError, setIntrospectError] = useState<string | null>(null);
  const [deleting, setDeleting] = useState<number | null>(null);

  // Redirect to storage edit page if trying to access tables for a new (unsaved) storage
  useEffect(() => {
    if (isNewStorage) {
      navigate(`/w/${slug}/storages/new`, { replace: true });
    }
  }, [isNewStorage, navigate, slug]);

  // Load configured tables from storage settings
  const loadConfiguredTables = useCallback(async () => {
    if (isNewStorage || !uuid) return;

    const { data, error } = await client.GET('/storage/postgres/{uuid}', {
      params: { path: { uuid } },
    });

    if (error) {
      message.error('Failed to load storage');
      return;
    }

    setConfiguredTables(data?.tables || []);
  }, [uuid, isNewStorage]);

  // Introspect database tables
  const introspectTables = useCallback(async () => {
    if (isNewStorage || !uuid) return;

    setIntrospecting(true);
    setIntrospectError(null);

    const { data, error } = await client.GET('/storage/postgres/{uuid}/introspect/tables', {
      params: { path: { uuid } },
    });

    setIntrospecting(false);

    if (error) {
      setIntrospectError(error.detail || 'Failed to connect to database');
      return;
    }

    setDbTables(data?.tables || []);
  }, [uuid, isNewStorage]);

  // Initial load
  useEffect(() => {
    const load = async () => {
      setLoading(true);
      await loadConfiguredTables();
      await introspectTables();
      setLoading(false);
    };
    load();
  }, [loadConfiguredTables, introspectTables]);

  const handleDelete = async (index: number) => {
    const newTables = configuredTables.filter((_, i) => i !== index);

    setDeleting(index);
    try {
      const { error } = await client.PUT('/storage/postgres/{uuid}/tables', {
        params: { path: { uuid } },
        body: newTables,
      });

      if (error) {
        message.error(error.detail || 'Failed to save changes');
        setDeleting(null);
        return;
      }

      setConfiguredTables(newTables);
      message.success('Table removed');
    } catch {
      message.error('Failed to save changes');
    }
    setDeleting(null);
  };

  // Filter out tables that are already configured
  const configuredNames = new Set(configuredTables.map((t) => t.name));
  const unconfiguredDbTables = dbTables.filter((t) => !configuredNames.has(t.name));

  const configuredColumns: ColumnsType<PostgresTable> = [
    {
      title: 'Table Name',
      dataIndex: 'name',
      key: 'name',
      render: (name: string) => <code>{name}</code>,
    },
    {
      title: 'Mode',
      dataIndex: 'creation_mode',
      key: 'creation_mode',
      width: 150,
      render: (mode: string) => (
        <Tag color={mode === 'auto_create' ? 'green' : 'blue'}>
          {mode === 'auto_create' ? 'Create' : 'Validate'}
        </Tag>
      ),
    },
    {
      title: 'Fields',
      key: 'fields',
      width: 100,
      render: (_: unknown, record: PostgresTable) => record.fields?.length || 0,
    },
    {
      title: '',
      key: 'actions',
      width: 60,
      render: (_: unknown, __: PostgresTable, index: number) => (
        <Popconfirm
          title="Remove table"
          description="Are you sure you want to remove this table configuration?"
          onConfirm={(e) => {
            e?.stopPropagation();
            handleDelete(index);
          }}
          onCancel={(e) => e?.stopPropagation()}
          okButtonProps={{ danger: true }}
          okText="Remove"
        >
          <Button
            type="text"
            danger
            icon={<DeleteOutlined />}
            size="small"
            loading={deleting === index}
            onClick={(e) => e.stopPropagation()}
          />
        </Popconfirm>
      ),
    },
  ];

  const dbTableColumns: ColumnsType<IntrospectTable> = [
    {
      title: 'Table Name',
      dataIndex: 'name',
      key: 'name',
      render: (name: string) => <code>{name}</code>,
    },
    {
      title: 'Rows',
      dataIndex: 'row_count',
      key: 'row_count',
      width: 120,
      render: (count: number) =>
        count !== undefined ? (
          <Text type="secondary">~{count.toLocaleString()}</Text>
        ) : (
          <Text type="secondary">-</Text>
        ),
    },
    {
      title: 'Primary Key',
      dataIndex: 'has_primary_key',
      key: 'has_primary_key',
      width: 120,
      render: (hasPK: boolean) =>
        hasPK ? <Tag color="green">Yes</Tag> : <Tag>No</Tag>,
    },
    {
      title: '',
      key: 'actions',
      width: 100,
      render: (_: unknown, record: IntrospectTable) => (
        <Button
          type="link"
          icon={<ImportOutlined />}
          onClick={() =>
            navigate(`/w/${slug}/storages/${uuid}/tables/new?import=${encodeURIComponent(record.name)}`)
          }
        >
          Import
        </Button>
      ),
    },
  ];

  if (loading) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', padding: 48 }}>
        <Spin size="large" />
      </div>
    );
  }

  return (
    <>
      <Space style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between' }}>
        <Space>
          <Button
            icon={<ArrowLeftOutlined />}
            onClick={() => navigate(`/w/${slug}/storages/${uuid}`)}
          >
            Back to Storage
          </Button>
          <Title level={4} style={{ margin: 0 }}>
            Target Tables
          </Title>
        </Space>
        <Button
          type="primary"
          icon={<PlusOutlined />}
          onClick={() => navigate(`/w/${slug}/storages/${uuid}/tables/new`)}
        >
          Add Table
        </Button>
      </Space>

      {/* Configured Tables Section */}
      <Card
        title="Configured Tables"
        size="small"
        style={{ marginBottom: 16 }}
        extra={
          <Text type="secondary">{configuredTables.length} table(s)</Text>
        }
      >
        {configuredTables.length === 0 ? (
          <Empty
            image={Empty.PRESENTED_IMAGE_SIMPLE}
            description="No tables configured yet. Add tables to define where data will be stored."
          >
            <Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={() => navigate(`/w/${slug}/storages/${uuid}/tables/new`)}
            >
              Add Table
            </Button>
          </Empty>
        ) : (
          <Table
            dataSource={configuredTables}
            columns={configuredColumns}
            rowKey={(_, index) => `configured-${index}`}
            pagination={false}
            size="small"
            onRow={(_, index) => ({
              onClick: () => navigate(`/w/${slug}/storages/${uuid}/tables/${index}`),
              style: { cursor: 'pointer' },
            })}
          />
        )}
      </Card>

      <Divider />

      {/* Available Database Tables Section */}
      <Card
        title={
          <Space>
            <DatabaseOutlined />
            Available in Database
          </Space>
        }
        size="small"
        extra={
          <Space>
            <Text type="secondary">{unconfiguredDbTables.length} table(s)</Text>
            <Button
              icon={<ReloadOutlined />}
              size="small"
              loading={introspecting}
              onClick={introspectTables}
            >
              Refresh
            </Button>
          </Space>
        }
      >
        {introspectError ? (
          <Alert
            type="warning"
            message="Could not connect to database"
            description={introspectError}
            showIcon
            action={
              <Button size="small" onClick={introspectTables} loading={introspecting}>
                Retry
              </Button>
            }
          />
        ) : unconfiguredDbTables.length === 0 ? (
          <Empty
            image={Empty.PRESENTED_IMAGE_SIMPLE}
            description={
              dbTables.length === 0
                ? 'No tables found in database. You can create new tables.'
                : 'All database tables are already configured.'
            }
          />
        ) : (
          <Table
            dataSource={unconfiguredDbTables}
            columns={dbTableColumns}
            rowKey="name"
            pagination={false}
            size="small"
          />
        )}
      </Card>
    </>
  );
}

export default PostgresTablesList;
