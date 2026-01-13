import { useEffect, useState, useCallback } from 'react';
import { useNavigate, useParams } from 'react-router';
import { Button, Space, Typography, message, Spin, Empty } from 'antd';
import { ArrowLeftOutlined, SaveOutlined } from '@ant-design/icons';
import client from '../../api/client';
import { useWorkspace } from '../../lib/workspace/WorkspaceContext';
import type { components } from '../../api/v1';
import PipelineCanvas from './components/PipelineCanvas';

const { Title, Text } = Typography;

type Datasource = components['schemas']['datasource'];
type Storage = components['schemas']['storage'];
type MapperFieldMapping = components['schemas']['mapper_field_mapping'];
type MapperConfig = components['schemas']['mapper_config'];
type Pipeline = components['schemas']['pipeline'];
type SourceFieldDefinition = components['schemas']['source_field_definition'];
type StoragePostgresTable = components['schemas']['storage_postgres_table'];
type TransformDefinition = components['schemas']['transform_definition'];

function PipelineFlowEdit() {
  const navigate = useNavigate();
  const { uuid } = useParams<{ uuid: string }>();
  const { slug } = useWorkspace();
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [pipeline, setPipeline] = useState<Pipeline | null>(null);
  const [datasource, setDatasource] = useState<Datasource | undefined>();
  const [storage, setStorage] = useState<Storage | undefined>();
  const [mappings, setMappings] = useState<MapperFieldMapping[]>([]);
  const [sourceFields, setSourceFields] = useState<SourceFieldDefinition[]>([]);
  const [targetTables, setTargetTables] = useState<StoragePostgresTable[]>([]);
  const [transforms, setTransforms] = useState<TransformDefinition[]>([]);

  const loadPipeline = useCallback(async () => {
    if (!uuid) {
      message.error('Pipeline UUID is required');
      navigate(`/w/${slug}/pipelines`);
      return;
    }

    setLoading(true);

    // Load pipeline, datasources, storages, and transforms in parallel
    const [pipelineRes, dsRes, stRes, transformRes] = await Promise.all([
      client.GET('/pipeline/{uuid}', {
        params: { path: { uuid } },
      }),
      client.GET('/datasource'),
      client.GET('/storage'),
      client.GET('/mapper/transforms'),
    ]);

    if (pipelineRes.error || !pipelineRes.data) {
      message.error('Failed to load pipeline');
      navigate(`/w/${slug}/pipelines`);
      return;
    }

    const pipelineData = pipelineRes.data;
    setPipeline(pipelineData);

    // Extract mapper config from flow nodes
    const flow = pipelineData.flow as Pipeline['flow'];
    if (flow?.nodes) {
      const mapperNode = flow.nodes.find((n) => n.type === 'mapper');
      if (mapperNode?.data?.config) {
        const config = mapperNode.data.config as MapperConfig;
        setMappings(config.mappings || []);
      }
    }

    // Find datasource and storage by UUID
    let foundDatasource: Datasource | undefined;
    let foundStorage: Storage | undefined;

    if (dsRes.data) {
      foundDatasource = dsRes.data.find((d) => d.uuid === pipelineData.datasource_uuid);
      if (foundDatasource) {
        setDatasource(foundDatasource);
      }
    }
    if (stRes.data) {
      foundStorage = stRes.data.find((s) => s.uuid === pipelineData.storage_uuid);
      if (foundStorage) {
        setStorage(foundStorage);
      }
    }

    // Load transforms
    if (transformRes.data?.transforms) {
      setTransforms(transformRes.data.transforms);
    }

    // Load source fields based on datasource type
    if (foundDatasource) {
      const sourceRes = await client.GET('/mapper/source-fields', {
        params: {
          query: {
            datasource_type: foundDatasource.type as
              | 'email'
              | 'email_oauth'
              | 'telegram'
              | 'whatsapp'
              | 'linkedin',
          },
        },
      });
      if (sourceRes.data?.fields) {
        setSourceFields(sourceRes.data.fields);
      }
    }

    // Load target tables for postgres storage
    if (foundStorage?.type === 'postgres') {
      const tablesRes = await client.GET('/storage/postgres/{uuid}', {
        params: { path: { uuid: foundStorage.uuid } },
      });
      if (tablesRes.data?.tables) {
        setTargetTables(tablesRes.data.tables);
      }
    }

    setLoading(false);
  }, [uuid, slug, navigate]);

  useEffect(() => {
    loadPipeline();
  }, [loadPipeline]);

  const handleSave = async () => {
    if (!pipeline || !uuid) return;

    setSaving(true);

    // Build mapper config
    const mapperConfig: MapperConfig = {
      version: '1.0',
      mappings,
    };

    // Build flow with mapper node
    const flow: Pipeline['flow'] = {
      nodes: [
        {
          id: 'mapper-1',
          type: 'mapper',
          position: { x: 0, y: 0 },
          data: {
            label: 'Mapper',
            config: mapperConfig,
          },
        },
      ],
      edges: [],
    };

    try {
      const { error } = await client.PUT('/pipeline/{uuid}', {
        params: { path: { uuid } },
        body: {
          name: pipeline.name,
          datasource_uuid: pipeline.datasource_uuid,
          storage_uuid: pipeline.storage_uuid,
          worker_uuid: pipeline.worker_uuid,
          is_enabled: pipeline.is_enabled,
          flow,
        },
      });

      if (error) throw new Error((error as { detail?: string }).detail);
      message.success('Flow saved');
      navigate(`/w/${slug}/pipelines/${uuid}`);
    } catch (err) {
      message.error(err instanceof Error ? err.message : 'Failed to save flow');
    } finally {
      setSaving(false);
    }
  };

  const handleBack = () => {
    navigate(`/w/${slug}/pipelines/${uuid}`);
  };

  if (loading) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', padding: 48 }}>
        <Spin size="large" />
      </div>
    );
  }

  // Show message if storage is not postgres (only postgres supports field mapping)
  if (storage && storage.type !== 'postgres') {
    return (
      <div style={{ display: 'flex', flexDirection: 'column', height: 'calc(100vh - 140px)' }}>
        <div
          style={{
            marginBottom: 16,
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
          }}
        >
          <Space>
            <Button icon={<ArrowLeftOutlined />} onClick={handleBack}>
              Back
            </Button>
            <Title level={4} style={{ margin: 0 }}>
              {pipeline?.name} - Flow Editor
            </Title>
          </Space>
        </div>
        <div style={{ flex: 1, display: 'flex', justifyContent: 'center', alignItems: 'center' }}>
          <Empty
            description={
              <Text type="secondary">
                Field mapping is only available for PostgreSQL storage. This pipeline uses{' '}
                {storage.type} storage.
              </Text>
            }
          />
        </div>
      </div>
    );
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', height: 'calc(100vh - 140px)' }}>
      <div
        style={{
          marginBottom: 16,
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
        }}
      >
        <Space>
          <Button icon={<ArrowLeftOutlined />} onClick={handleBack}>
            Back
          </Button>
          <Title level={4} style={{ margin: 0 }}>
            {pipeline?.name} - Flow Editor
          </Title>
        </Space>
        <Button type="primary" icon={<SaveOutlined />} onClick={handleSave} loading={saving}>
          Save
        </Button>
      </div>

      <div style={{ flex: 1, minHeight: 0 }}>
        <PipelineCanvas
          mappings={mappings}
          onMappingsChange={setMappings}
          sourceFields={sourceFields}
          targetTables={targetTables}
          transforms={transforms}
          datasourceName={datasource?.name}
          storageName={storage?.name}
          fullScreen
        />
      </div>
    </div>
  );
}

export default PipelineFlowEdit;
