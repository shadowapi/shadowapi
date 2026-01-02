import { useEffect, useState, useCallback } from 'react';
import { Table, Select, Button, Space, Checkbox, Spin, Empty, Typography, message, Alert } from 'antd';
import { DeleteOutlined, PlusOutlined } from '@ant-design/icons';
import client from '../../../api/client';
import type { components } from '../../../api/v1';

type MapperFieldMapping = components['schemas']['mapper_field_mapping'];
type SourceFieldDefinition = components['schemas']['source_field_definition'];
type TransformDefinition = components['schemas']['transform_definition'];
type StoragePostgresTable = components['schemas']['storage_postgres_table'];

const { Text } = Typography;

interface MapperTableProps {
  storageUuid: string;
  datasourceType?: string;
  mappings: MapperFieldMapping[];
  onChange: (mappings: MapperFieldMapping[]) => void;
}

function generateId(): string {
  return `mapping-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
}

function MapperTable({ storageUuid, datasourceType, mappings, onChange }: MapperTableProps) {
  const [sourceFields, setSourceFields] = useState<SourceFieldDefinition[]>([]);
  const [transforms, setTransforms] = useState<TransformDefinition[]>([]);
  const [tables, setTables] = useState<StoragePostgresTable[]>([]);
  const [loading, setLoading] = useState(true);

  const loadData = useCallback(async () => {
    const [sourceRes, transformRes, storageRes] = await Promise.all([
      client.GET('/mapper/source-fields', {
        params: {
          query: datasourceType ? { datasource_type: datasourceType as 'email' | 'email_oauth' | 'telegram' | 'whatsapp' | 'linkedin' } : {},
        },
      }),
      client.GET('/mapper/transforms'),
      client.GET('/storage/postgres/{uuid}', {
        params: { path: { uuid: storageUuid } },
      }),
    ]);

    if (sourceRes.error) {
      console.error('Failed to load source fields:', sourceRes.error);
      message.error('Failed to load source fields');
    } else if (sourceRes.data?.fields) {
      setSourceFields(sourceRes.data.fields);
    }

    if (transformRes.error) {
      console.error('Failed to load transforms:', transformRes.error);
      message.error('Failed to load transforms');
    } else if (transformRes.data?.transforms) {
      setTransforms(transformRes.data.transforms);
    }

    if (storageRes.error) {
      console.error('Failed to load storage:', storageRes.error);
      message.error('Failed to load storage tables');
    } else if (storageRes.data?.tables) {
      setTables(storageRes.data.tables);
    }

    setLoading(false);
  }, [storageUuid, datasourceType]);

  useEffect(() => {
    void loadData();
  }, [loadData]);

  const updateMapping = (index: number, updates: Partial<MapperFieldMapping>) => {
    const newMappings = [...mappings];
    newMappings[index] = { ...newMappings[index], ...updates };
    onChange(newMappings);
  };

  const deleteMapping = (index: number) => {
    const newMappings = mappings.filter((_, i) => i !== index);
    onChange(newMappings);
  };

  const addMapping = () => {
    const newMapping: MapperFieldMapping = {
      id: generateId(),
      source_entity: 'message',
      source_field: '',
      transform: { type: 'set' },
      target_table: '',
      target_field: '',
      is_enabled: true,
    };
    onChange([...mappings, newMapping]);
  };

  // Build source field options grouped by entity (filter out empty groups)
  const sourceFieldOptions = [
    {
      label: 'Message',
      options: sourceFields
        .filter((f) => f.entity === 'message')
        .map((f) => ({
          value: `message:${f.name}`,
          label: f.name,
        })),
    },
    {
      label: 'Contact',
      options: sourceFields
        .filter((f) => f.entity === 'contact')
        .map((f) => ({
          value: `contact:${f.name}`,
          label: f.name,
        })),
    },
  ].filter((group) => group.options.length > 0);

  // Build transform options
  const transformOptions = transforms.map((t) => ({
    value: t.type,
    label: t.display_name,
  }));

  // Build target field options grouped by table
  const targetFieldOptions = tables.map((table) => ({
    label: table.name,
    options: (table.fields || []).map((field) => ({
      value: `${table.name}:${field.name}`,
      label: `${table.name}.${field.name}`,
    })),
  }));

  const columns = [
    {
      title: 'Source Field',
      key: 'source',
      width: 200,
      render: (_: unknown, record: MapperFieldMapping, index: number) => {
        const value = record.source_field
          ? `${record.source_entity}:${record.source_field}`
          : undefined;
        return (
          <Select
            value={value}
            options={sourceFieldOptions}
            placeholder="Select source field"
            onChange={(val: string) => {
              const [entity, field] = val.split(':');
              updateMapping(index, {
                source_entity: entity as 'message' | 'contact',
                source_field: field,
              });
            }}
            style={{ width: '100%' }}
            size="small"
            showSearch
            filterOption={(input, option) =>
              (option?.label ?? '').toString().toLowerCase().includes(input.toLowerCase())
            }
          />
        );
      },
    },
    {
      title: 'Transform',
      key: 'transform',
      width: 150,
      render: (_: unknown, record: MapperFieldMapping, index: number) => (
        <Select
          value={record.transform?.type || 'set'}
          options={transformOptions}
          onChange={(type: string) => {
            updateMapping(index, {
              transform: { type: type as MapperFieldMapping['transform']['type'] },
            });
          }}
          style={{ width: '100%' }}
          size="small"
        />
      ),
    },
    {
      title: 'Target Field',
      key: 'target',
      width: 200,
      render: (_: unknown, record: MapperFieldMapping, index: number) => {
        const value =
          record.target_table && record.target_field
            ? `${record.target_table}:${record.target_field}`
            : undefined;
        return (
          <Select
            value={value}
            options={targetFieldOptions}
            placeholder="Select target field"
            onChange={(val: string) => {
              const [table, field] = val.split(':');
              updateMapping(index, {
                target_table: table,
                target_field: field,
              });
            }}
            style={{ width: '100%' }}
            size="small"
            showSearch
            filterOption={(input, option) =>
              (option?.label ?? '').toString().toLowerCase().includes(input.toLowerCase())
            }
          />
        );
      },
    },
    {
      title: 'Enabled',
      key: 'enabled',
      width: 70,
      align: 'center' as const,
      render: (_: unknown, record: MapperFieldMapping, index: number) => (
        <Checkbox
          checked={record.is_enabled}
          onChange={(e) => updateMapping(index, { is_enabled: e.target.checked })}
        />
      ),
    },
    {
      title: '',
      key: 'actions',
      width: 50,
      render: (_: unknown, __: MapperFieldMapping, index: number) => (
        <Button
          type="text"
          danger
          icon={<DeleteOutlined />}
          onClick={() => deleteMapping(index)}
          size="small"
        />
      ),
    },
  ];

  if (loading) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', padding: 24 }}>
        <Spin />
      </div>
    );
  }

  if (tables.length === 0) {
    return (
      <Empty
        description={
          <Text type="secondary">
            No tables defined in the selected storage. Please add tables to the storage first.
          </Text>
        }
      />
    );
  }

  if (sourceFields.length === 0) {
    return (
      <Alert
        type="warning"
        message="Unable to load source fields"
        description="The mapper source fields API may not be available. Please check the browser console for errors."
      />
    );
  }

  return (
    <div>
      <Table
        dataSource={mappings}
        columns={columns}
        rowKey={(record) => record.id || `row-${Math.random()}`}
        pagination={false}
        size="small"
        bordered
      />
      <Space style={{ marginTop: 8 }}>
        <Button type="dashed" onClick={addMapping} icon={<PlusOutlined />} size="small">
          Add Mapping
        </Button>
      </Space>
    </div>
  );
}

export default MapperTable;
