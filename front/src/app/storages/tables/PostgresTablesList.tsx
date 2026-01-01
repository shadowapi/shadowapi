import { useState } from 'react';
import { useNavigate, useParams } from 'react-router';
import { Table, Button, Space, Typography, Tag, Popconfirm, message, Empty } from 'antd';
import { PlusOutlined, ArrowLeftOutlined, DeleteOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import { useWorkspace } from '../../../lib/workspace/WorkspaceContext';
import { getStorageKey, getTables, setTables } from '../../../lib/storage/storageTablesStore';
import client from '../../../api/client';
import type { components } from '../../../api/v1';

type PostgresTable = components['schemas']['storage_postgres_table'];

const { Title } = Typography;

function PostgresTablesList() {
  const navigate = useNavigate();
  const { uuid } = useParams<{ uuid: string }>();
  const { slug } = useWorkspace();
  const storageKey = getStorageKey(uuid);
  const isNewStorage = uuid === 'new';

  // Read tables from sessionStorage and use state to trigger re-renders after mutations
  const [tables, setLocalTables] = useState<PostgresTable[]>(() => getTables(storageKey));
  const [deleting, setDeleting] = useState<number | null>(null);

  const handleDelete = async (index: number) => {
    const newTables = tables.filter((_, i) => i !== index);

    // Save to sessionStorage
    setTables(storageKey, newTables);
    setLocalTables(newTables);

    // For existing storages, also save to backend
    if (!isNewStorage && uuid) {
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

        message.success('Table removed');
      } catch (err) {
        message.error('Failed to save changes');
      }
      setDeleting(null);
    } else {
      message.success('Table removed');
    }
  };

  const columns: ColumnsType<PostgresTable> = [
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
          {mode === 'auto_create' ? 'Auto Create' : 'Validate'}
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
          description="Are you sure you want to remove this table definition?"
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
            onClick={(e) => e.stopPropagation()}
          />
        </Popconfirm>
      ),
    },
  ];

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

      {tables.length === 0 ? (
        <Empty
          description={
            isNewStorage
              ? 'No tables defined yet. Add tables to define where data will be stored.'
              : 'No tables defined. Add tables to define where data will be stored.'
          }
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
          dataSource={tables}
          columns={columns}
          rowKey={(_, index) => `table-${index}`}
          pagination={false}
          onRow={(_, index) => ({
            onClick: () => navigate(`/w/${slug}/storages/${uuid}/tables/${index}`),
            style: { cursor: 'pointer' },
          })}
        />
      )}
    </>
  );
}

export default PostgresTablesList;
