import { useState, useEffect, useCallback } from 'react';
import { useNavigate, useParams, useSearchParams } from 'react-router';
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
  Modal,
  Spin,
  AutoComplete,
} from 'antd';
import { ArrowLeftOutlined, ExclamationCircleOutlined, WarningOutlined } from '@ant-design/icons';
import { useWorkspace } from '../../../lib/workspace/WorkspaceContext';
import FieldsTable from '../components/FieldsTable';
import client from '../../../api/client';
import type { components } from '../../../api/v1';

type PostgresTable = components['schemas']['storage_postgres_table'];
type PostgresField = components['schemas']['storage_postgres_field'];
type IntrospectField = components['schemas']['StoragePostgresIntrospectField'];
type CreationMode = PostgresTable['creation_mode'];

const { Title, Text } = Typography;

const creationModeOptions: { value: CreationMode; label: string; description: string }[] = [
  {
    value: 'validate_existing',
    label: 'Validate Existing',
    description: 'Import schema from an existing table in the database',
  },
  {
    value: 'auto_create',
    label: 'Create New',
    description: 'Create a new table in the database with the specified schema',
  },
];

// Map introspect field type to our field type
function mapIntrospectFieldType(type: string): PostgresField['type'] {
  return type as PostgresField['type'];
}

function PostgresTableEdit() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const { uuid: rawUuid, index } = useParams<{ uuid: string; index: string }>();
  const { slug } = useWorkspace();
  const [form] = Form.useForm();
  const [validationError, setValidationError] = useState<string | null>(null);

  // For route /storages/new/tables/*, uuid param is undefined (literal 'new' route)
  const uuid = rawUuid || 'new';
  const isNewTable = index === 'new' || index === undefined;
  const isNewStorage = uuid === 'new';
  const tableIndex = isNewTable ? -1 : parseInt(index, 10);
  const [saving, setSaving] = useState(false);

  // Import table name from query parameter
  const importTableName = searchParams.get('import');

  // Database tables for autocomplete
  const [dbTables, setDbTables] = useState<string[]>([]);
  const [loadingDbTables, setLoadingDbTables] = useState(false);

  // Table existence check for create mode
  const [tableExists, setTableExists] = useState(false);
  const [checkingExists, setCheckingExists] = useState(false);
  const [dropConfirmOpen, setDropConfirmOpen] = useState(false);
  const [dropIfExists, setDropIfExists] = useState(false);

  // Loading fields from database
  const [loadingFields, setLoadingFields] = useState(false);

  // Configured tables for duplicate check
  const [configuredTables, setConfiguredTables] = useState<PostgresTable[]>([]);

  // Fields state
  const [fields, setFields] = useState<PostgresField[]>([
    { name: 'id', type: 'TEXT', nullable: false, is_primary_key: true },
  ]);

  // Current creation mode
  const creationMode = Form.useWatch('creation_mode', form) as CreationMode;
  const tableName = Form.useWatch('name', form) as string;

  // Redirect to storage edit page if trying to access tables for a new (unsaved) storage
  useEffect(() => {
    if (isNewStorage) {
      navigate(`/w/${slug}/storages/new`, { replace: true });
    }
  }, [isNewStorage, navigate, slug]);

  // Load database tables for autocomplete
  const loadDbTables = useCallback(async () => {
    if (isNewStorage || !uuid) return;

    setLoadingDbTables(true);
    const { data, error } = await client.GET('/storage/postgres/{uuid}/introspect/tables', {
      params: { path: { uuid } },
    });
    setLoadingDbTables(false);

    if (!error && data?.tables) {
      setDbTables(data.tables.map((t) => t.name));
    }
  }, [uuid, isNewStorage]);

  // Load configured tables
  const loadConfiguredTables = useCallback(async () => {
    if (isNewStorage || !uuid) return;

    const { data, error } = await client.GET('/storage/postgres/{uuid}', {
      params: { path: { uuid } },
    });

    if (!error && data?.tables) {
      setConfiguredTables(data.tables);
    }
  }, [uuid, isNewStorage]);

  // Load table fields from database (for validate mode)
  const loadTableFields = useCallback(
    async (tableName: string) => {
      if (!uuid || !tableName) return;

      setLoadingFields(true);
      const { data, error } = await client.GET('/storage/postgres/{uuid}/introspect/tables/{table_name}', {
        params: { path: { uuid, table_name: tableName } },
      });
      setLoadingFields(false);

      if (error) {
        message.error('Failed to load table schema');
        return;
      }

      if (data?.exists && data.fields) {
        // Map introspect fields to our field format
        const mappedFields: PostgresField[] = data.fields.map((f: IntrospectField) => ({
          name: f.name,
          type: mapIntrospectFieldType(f.type),
          nullable: f.nullable,
          is_primary_key: f.is_primary_key,
          default_value: f.default_value || undefined,
        }));
        setFields(mappedFields);
        setTableExists(true);
      } else {
        setTableExists(false);
        message.warning(`Table "${tableName}" does not exist in the database`);
      }
    },
    [uuid]
  );

  // Check if table exists (for create mode)
  const checkTableExists = useCallback(
    async (tableName: string) => {
      if (!uuid || !tableName) return;

      setCheckingExists(true);
      const { data } = await client.GET('/storage/postgres/{uuid}/introspect/tables/{table_name}', {
        params: { path: { uuid, table_name: tableName } },
      });
      setCheckingExists(false);

      if (data?.exists) {
        setTableExists(true);
      } else {
        setTableExists(false);
        setDropIfExists(false);
      }
    },
    [uuid]
  );

  // Initial data load
  useEffect(() => {
    const load = async () => {
      await loadDbTables();
      await loadConfiguredTables();

      // If editing existing table, load from backend
      if (!isNewTable && uuid) {
        const { data } = await client.GET('/storage/postgres/{uuid}', {
          params: { path: { uuid } },
        });
        const table = data?.tables?.[tableIndex];
        if (!table) {
          navigate(`/w/${slug}/storages/${uuid}/tables`);
          return;
        }
        form.setFieldsValue({
          name: table.name,
          creation_mode: table.creation_mode,
        });
        setFields(table.fields || []);
      }
      // If importing from query param
      else if (importTableName) {
        form.setFieldsValue({
          name: importTableName,
          creation_mode: 'validate_existing',
        });
        await loadTableFields(importTableName);
      }
    };
    load();
  }, [
    isNewTable,
    uuid,
    tableIndex,
    loadDbTables,
    loadConfiguredTables,
    loadTableFields,
    importTableName,
    form,
    navigate,
    slug,
  ]);

  // When creation mode or table name changes
  useEffect(() => {
    if (!tableName) return;

    if (creationMode === 'validate_existing') {
      // Load fields from database
      loadTableFields(tableName);
    } else if (creationMode === 'auto_create') {
      // Check if table exists
      checkTableExists(tableName);
    }
  }, [creationMode, tableName, loadTableFields, checkTableExists]);

  const validate = (): string | null => {
    const values = form.getFieldsValue();

    // Table name validation
    if (!values.name || !values.name.trim()) {
      return 'Table name is required';
    }

    const tableNameVal = values.name.trim().toLowerCase();
    if (!/^[a-z][a-z0-9_]*$/.test(tableNameVal)) {
      return 'Table name must start with a letter and contain only lowercase letters, numbers, and underscores';
    }

    // Check for duplicate table names in configured tables (excluding current table when editing)
    const existingNames = configuredTables
      .filter((_, i) => i !== tableIndex)
      .map((t) => t.name);
    if (existingNames.includes(tableNameVal)) {
      return `A table named "${tableNameVal}" is already configured`;
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
    const tableNameVal = values.name.trim().toLowerCase();

    // For create mode, check if table exists and prompt for drop
    if (creationMode === 'auto_create' && tableExists && !dropIfExists) {
      setDropConfirmOpen(true);
      return;
    }

    setSaving(true);

    try {
      // For create mode, create the table in database first
      if (creationMode === 'auto_create') {
        const { data: createResult, error: createError } = await client.POST(
          '/storage/postgres/{uuid}/tables/create',
          {
            params: { path: { uuid } },
            body: {
              name: tableNameVal,
              fields: fields.map((f) => ({
                name: f.name.trim().toLowerCase(),
                type: f.type,
                nullable: f.nullable ?? true,
                is_primary_key: f.is_primary_key ?? false,
                default_value: f.default_value,
              })),
              drop_if_exists: dropIfExists,
            },
          }
        );

        if (createError || !createResult?.success) {
          message.error(createResult?.error || createError?.detail || 'Failed to create table in database');
          setSaving(false);
          return;
        }

        if (createResult.was_dropped) {
          message.info('Existing table was dropped and recreated');
        }
      }

      // Now save the configuration
      const newTable: PostgresTable = {
        name: tableNameVal,
        creation_mode: values.creation_mode,
        fields: fields.map((f) => ({
          ...f,
          name: f.name.trim().toLowerCase(),
        })),
      };

      // Load current tables and update
      const { data: storageData } = await client.GET('/storage/postgres/{uuid}', {
        params: { path: { uuid } },
      });
      const tables = storageData?.tables || [];

      if (isNewTable) {
        tables.push(newTable);
      } else {
        tables[tableIndex] = newTable;
      }

      const { error: saveError } = await client.PUT('/storage/postgres/{uuid}/tables', {
        params: { path: { uuid } },
        body: tables,
      });

      if (saveError) {
        message.error(saveError.detail || 'Failed to save table configuration');
        setSaving(false);
        return;
      }

      message.success(isNewTable ? 'Table added' : 'Table saved');
      navigate(`/w/${slug}/storages/${uuid}/tables`);
    } catch {
      message.error('Failed to save table');
    }
    setSaving(false);
  };

  const handleDropConfirm = () => {
    setDropIfExists(true);
    setDropConfirmOpen(false);
    // Re-trigger save with drop flag set
    setTimeout(() => handleSave(), 0);
  };

  const handleCancel = () => {
    navigate(`/w/${slug}/storages/${uuid}/tables`);
  };

  // Autocomplete options for table name
  const tableNameOptions = dbTables.map((name) => ({ value: name, label: name }));

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
        initialValues={{ creation_mode: 'validate_existing' }}
        style={{ maxWidth: 800 }}
      >
        <Form.Item
          name="creation_mode"
          label="Mode"
          rules={[{ required: true, message: 'Mode is required' }]}
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

        <Form.Item
          name="name"
          label="Table Name"
          rules={[{ required: true, message: 'Table name is required' }]}
          extra={
            creationMode === 'validate_existing'
              ? 'Select an existing table from the database'
              : 'Enter a name for the new table'
          }
        >
          {creationMode === 'validate_existing' ? (
            <AutoComplete
              options={tableNameOptions}
              placeholder="Select or type table name"
              filterOption={(inputValue, option) =>
                option!.value.toLowerCase().includes(inputValue.toLowerCase())
              }
              notFoundContent={loadingDbTables ? <Spin size="small" /> : 'No tables found'}
            />
          ) : (
            <Input placeholder="my_table_name" />
          )}
        </Form.Item>

        {/* Warning for create mode when table exists */}
        {creationMode === 'auto_create' && tableExists && (
          <Alert
            message="Table already exists"
            description={
              <span>
                A table named <code>{tableName}</code> already exists in the database.
                If you proceed, you can choose to drop and recreate it.
              </span>
            }
            type="warning"
            showIcon
            icon={<WarningOutlined />}
            style={{ marginBottom: 16 }}
          />
        )}

        <Divider />

        <Form.Item
          label="Fields"
          required
          extra={
            creationMode === 'validate_existing'
              ? 'Fields are imported from the existing table'
              : 'Define the columns for your new table'
          }
        >
          {loadingFields ? (
            <div style={{ textAlign: 'center', padding: 24 }}>
              <Spin />
              <br />
              <Text type="secondary">Loading table schema...</Text>
            </div>
          ) : (
            <FieldsTable
              fields={fields}
              onChange={setFields}
              readOnly={creationMode === 'validate_existing'}
            />
          )}
        </Form.Item>

        <Form.Item>
          <Space>
            <Button type="primary" onClick={handleSave} loading={saving || checkingExists}>
              {isNewTable ? 'Add Table' : 'Save'}
            </Button>
            <Button onClick={handleCancel}>Cancel</Button>
          </Space>
        </Form.Item>
      </Form>

      {/* Drop confirmation modal */}
      <Modal
        title={
          <Space>
            <ExclamationCircleOutlined style={{ color: '#faad14' }} />
            Table Already Exists
          </Space>
        }
        open={dropConfirmOpen}
        onOk={handleDropConfirm}
        onCancel={() => setDropConfirmOpen(false)}
        okText="Drop and Recreate"
        okButtonProps={{ danger: true }}
        cancelText="Cancel"
      >
        <p>
          The table <code>{tableName}</code> already exists in the database.
        </p>
        <p>
          <strong>Warning:</strong> Choosing to drop and recreate will delete all existing data in
          this table. This action cannot be undone.
        </p>
        <p>Do you want to drop the existing table and create a new one with your schema?</p>
      </Modal>
    </>
  );
}

export default PostgresTableEdit;
