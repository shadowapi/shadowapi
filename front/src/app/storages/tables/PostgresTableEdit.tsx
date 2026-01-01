import { useState, useEffect } from 'react';
import { useNavigate, useParams } from 'react-router';
import {
  Form,
  Input,
  Select,
  Button,
  Space,
  Typography,
  Alert,
  Divider,
  message,
} from 'antd';
import { ArrowLeftOutlined } from '@ant-design/icons';
import { useWorkspace } from '../../../lib/workspace/WorkspaceContext';
import { getStorageKey, getTables, setTables } from '../../../lib/storage/storageTablesStore';
import FieldsTable from '../components/FieldsTable';
import client from '../../../api/client';
import type { components } from '../../../api/v1';

type PostgresTable = components['schemas']['storage_postgres_table'];
type PostgresField = components['schemas']['storage_postgres_field'];
type CreationMode = PostgresTable['creation_mode'];

const { Title, Text } = Typography;

const creationModeOptions: { value: CreationMode; label: string; description: string }[] = [
  {
    value: 'auto_create',
    label: 'Auto Create',
    description: 'Creates the table if it does not exist in the database',
  },
  {
    value: 'validate_existing',
    label: 'Validate Existing',
    description: 'Verifies the existing table schema matches this definition',
  },
];

function PostgresTableEdit() {
  const navigate = useNavigate();
  const { uuid: rawUuid, index } = useParams<{ uuid: string; index: string }>();
  const { slug } = useWorkspace();
  const [form] = Form.useForm();
  const [validationError, setValidationError] = useState<string | null>(null);

  // For route /storages/new/tables/*, uuid param is undefined (literal 'new' route)
  // Normalize to 'new' for consistency
  const uuid = rawUuid || 'new';
  const storageKey = getStorageKey(uuid);
  const isNewTable = index === 'new' || index === undefined;
  const isNewStorage = uuid === 'new';
  const tableIndex = isNewTable ? -1 : parseInt(index, 10);
  const [saving, setSaving] = useState(false);

  // Redirect to storage edit page if trying to access tables for a new (unsaved) storage
  useEffect(() => {
    if (isNewStorage) {
      navigate(`/w/${slug}/storages/new`, { replace: true });
    }
  }, [isNewStorage, navigate, slug]);

  // Initialize fields from sessionStorage or defaults
  const [fields, setFields] = useState<PostgresField[]>(() => {
    if (isNewTable) {
      return [{ name: 'id', type: 'TEXT', nullable: false, is_primary_key: true }];
    }
    const tables = getTables(storageKey);
    const table = tables[tableIndex];
    return table?.fields || [];
  });

  // Set form values on mount
  useEffect(() => {
    if (isNewTable) {
      form.setFieldsValue({
        name: '',
        creation_mode: 'auto_create',
      });
    } else {
      const tables = getTables(storageKey);
      const table = tables[tableIndex];
      if (!table) {
        navigate(`/w/${slug}/storages/${uuid}/tables`);
        return;
      }
      form.setFieldsValue({
        name: table.name,
        creation_mode: table.creation_mode,
      });
    }
  }, [storageKey, isNewTable, tableIndex, form, navigate, slug, uuid]);

  const validate = (): string | null => {
    const values = form.getFieldsValue();
    const tables = getTables(storageKey);

    // Table name validation
    if (!values.name || !values.name.trim()) {
      return 'Table name is required';
    }

    const tableName = values.name.trim().toLowerCase();
    if (!/^[a-z][a-z0-9_]*$/.test(tableName)) {
      return 'Table name must start with a letter and contain only lowercase letters, numbers, and underscores';
    }

    // Check for duplicate table names (excluding current table when editing)
    const existingNames = tables
      .filter((_, i) => i !== tableIndex)
      .map((t) => t.name);
    if (existingNames.includes(tableName)) {
      return `A table named "${tableName}" already exists`;
    }

    // Fields validation
    if (fields.length === 0) {
      return 'At least one field is required';
    }

    const fieldNames = new Set<string>();
    let pkCount = 0;

    for (const field of fields) {
      if (!field.name || !field.name.trim()) {
        return 'All fields must have a name';
      }

      const fieldName = field.name.trim().toLowerCase();
      if (!/^[a-z][a-z0-9_]*$/.test(fieldName)) {
        return `Field name "${field.name}" must start with a letter and contain only lowercase letters, numbers, and underscores`;
      }

      if (fieldNames.has(fieldName)) {
        return `Duplicate field name: ${fieldName}`;
      }
      fieldNames.add(fieldName);

      if (field.is_primary_key) {
        pkCount++;
        if (field.nullable) {
          return `Primary key field "${field.name}" cannot be nullable`;
        }
      }
    }

    if (pkCount > 1) {
      return 'Only one field can be a primary key';
    }

    return null;
  };

  const handleSave = async () => {
    const error = validate();
    if (error) {
      setValidationError(error);
      return;
    }

    const values = form.getFieldsValue();
    const newTable: PostgresTable = {
      name: values.name.trim().toLowerCase(),
      creation_mode: values.creation_mode,
      fields: fields.map((f) => ({
        ...f,
        name: f.name.trim().toLowerCase(),
      })),
    };

    const tables = getTables(storageKey);

    if (isNewTable) {
      tables.push(newTable);
    } else {
      tables[tableIndex] = newTable;
    }

    // Save to sessionStorage
    setTables(storageKey, tables);

    // For existing storages, also save to backend
    if (!isNewStorage && uuid) {
      setSaving(true);
      try {
        const { error: apiError } = await client.PUT('/storage/postgres/{uuid}/tables', {
          params: { path: { uuid } },
          body: tables,
        });

        if (apiError) {
          message.error(apiError.detail || 'Failed to save tables');
          setSaving(false);
          return;
        }

        message.success('Tables saved');
      } catch (err) {
        message.error('Failed to save tables');
        setSaving(false);
        return;
      }
      setSaving(false);
    }

    navigate(`/w/${slug}/storages/${uuid}/tables`);
  };

  const handleCancel = () => {
    navigate(`/w/${slug}/storages/${uuid}/tables`);
  };

  return (
    <>
      <Space style={{ marginBottom: 16 }}>
        <Button icon={<ArrowLeftOutlined />} onClick={handleCancel}>
          Back to Tables
        </Button>
        <Title level={4} style={{ margin: 0 }}>
          {isNewTable ? 'Add Table' : 'Edit Table'}
        </Title>
      </Space>

      {validationError && (
        <Alert
          message={validationError}
          type="error"
          showIcon
          style={{ marginBottom: 16 }}
          closable
          onClose={() => setValidationError(null)}
        />
      )}

      <Form
        form={form}
        layout="vertical"
        initialValues={{ creation_mode: 'auto_create' }}
        style={{ maxWidth: 800 }}
      >
        <Form.Item
          name="name"
          label="Table Name"
          rules={[{ required: true, message: 'Table name is required' }]}
          extra="Lowercase letters, numbers, and underscores only"
        >
          <Input placeholder="my_table_name" />
        </Form.Item>

        <Form.Item
          name="creation_mode"
          label="Creation Mode"
          rules={[{ required: true, message: 'Creation mode is required' }]}
        >
          <Select
            options={creationModeOptions.map((opt) => ({
              value: opt.value,
              label: (
                <div>
                  <Text strong>{opt.label}</Text>
                  <br />
                  <Text type="secondary" style={{ fontSize: 12 }}>
                    {opt.description}
                  </Text>
                </div>
              ),
            }))}
          />
        </Form.Item>

        <Divider />

        <Form.Item label="Fields" required>
          <FieldsTable fields={fields} onChange={setFields} />
        </Form.Item>

        <Form.Item>
          <Space>
            <Button type="primary" onClick={handleSave} loading={saving}>
              {isNewTable ? 'Add Table' : 'Save'}
            </Button>
            <Button onClick={handleCancel}>Cancel</Button>
          </Space>
        </Form.Item>
      </Form>
    </>
  );
}

export default PostgresTableEdit;
