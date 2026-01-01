import { Table, Input, Select, Checkbox, Button, Space } from 'antd';
import { DeleteOutlined, PlusOutlined } from '@ant-design/icons';
import type { components } from '../../../api/v1';

type PostgresField = components['schemas']['storage_postgres_field'];
type FieldType = PostgresField['type'];

interface FieldsTableProps {
  fields: PostgresField[];
  onChange: (fields: PostgresField[]) => void;
}

const fieldTypeOptions: { value: FieldType; label: string }[] = [
  { value: 'TEXT', label: 'TEXT' },
  { value: 'INTEGER', label: 'INTEGER' },
  { value: 'BOOLEAN', label: 'BOOLEAN' },
  { value: 'TIMESTAMP', label: 'TIMESTAMP' },
  { value: 'JSONB', label: 'JSONB' },
];

function FieldsTable({ fields, onChange }: FieldsTableProps) {
  const updateField = (index: number, updates: Partial<PostgresField>) => {
    const newFields = [...fields];
    newFields[index] = { ...newFields[index], ...updates };
    onChange(newFields);
  };

  const deleteField = (index: number) => {
    const newFields = fields.filter((_, i) => i !== index);
    onChange(newFields);
  };

  const addField = () => {
    const newField: PostgresField = {
      name: '',
      type: 'TEXT',
      nullable: true,
      is_primary_key: false,
    };
    onChange([...fields, newField]);
  };

  const columns = [
    {
      title: 'Name',
      dataIndex: 'name',
      key: 'name',
      width: 150,
      render: (value: string, _: PostgresField, index: number) => (
        <Input
          value={value}
          placeholder="column_name"
          onChange={(e) => updateField(index, { name: e.target.value })}
          size="small"
        />
      ),
    },
    {
      title: 'Type',
      dataIndex: 'type',
      key: 'type',
      width: 120,
      render: (value: FieldType, _: PostgresField, index: number) => (
        <Select
          value={value}
          options={fieldTypeOptions}
          onChange={(newType) => updateField(index, { type: newType })}
          size="small"
          style={{ width: '100%' }}
        />
      ),
    },
    {
      title: 'Nullable',
      dataIndex: 'nullable',
      key: 'nullable',
      width: 80,
      align: 'center' as const,
      render: (value: boolean, _: PostgresField, index: number) => (
        <Checkbox
          checked={value}
          disabled={fields[index].is_primary_key}
          onChange={(e) => updateField(index, { nullable: e.target.checked })}
        />
      ),
    },
    {
      title: 'PK',
      dataIndex: 'is_primary_key',
      key: 'is_primary_key',
      width: 60,
      align: 'center' as const,
      render: (value: boolean, _: PostgresField, index: number) => (
        <Checkbox
          checked={value}
          onChange={(e) => {
            const updates: Partial<PostgresField> = { is_primary_key: e.target.checked };
            // Primary key cannot be nullable
            if (e.target.checked) {
              updates.nullable = false;
            }
            updateField(index, updates);
          }}
        />
      ),
    },
    {
      title: 'Default',
      dataIndex: 'default_value',
      key: 'default_value',
      width: 120,
      render: (value: string | undefined, _: PostgresField, index: number) => (
        <Input
          value={value || ''}
          placeholder="NOW()"
          onChange={(e) => updateField(index, { default_value: e.target.value || undefined })}
          size="small"
        />
      ),
    },
    {
      title: '',
      key: 'actions',
      width: 50,
      render: (_: unknown, __: PostgresField, index: number) => (
        <Button
          type="text"
          danger
          icon={<DeleteOutlined />}
          onClick={() => deleteField(index)}
          size="small"
        />
      ),
    },
  ];

  return (
    <div>
      <Table
        dataSource={fields}
        columns={columns}
        rowKey={(_, index) => `field-${index}`}
        pagination={false}
        size="small"
        bordered
      />
      <Space style={{ marginTop: 8 }}>
        <Button type="dashed" onClick={addField} icon={<PlusOutlined />} size="small">
          Add Field
        </Button>
      </Space>
    </div>
  );
}

export default FieldsTable;
