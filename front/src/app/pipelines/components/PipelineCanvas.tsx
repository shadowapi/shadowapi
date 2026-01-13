import { useMemo, useCallback } from 'react';
import {
  ReactFlow,
  Background,
  Controls,
  type Node,
  type Edge,
  type Connection,
  useNodesState,
  useEdgesState,
} from '@xyflow/react';
import '@xyflow/react/dist/style.css';
import { message } from 'antd';
import { InboxOutlined, UserOutlined, TableOutlined } from '@ant-design/icons';
import { FieldGroupNode } from './nodes';
import { TransformEdge } from './edges';
import type { components } from '../../../api/v1';

type MapperFieldMapping = components['schemas']['mapper_field_mapping'];
type SourceFieldDefinition = components['schemas']['source_field_definition'];
type StoragePostgresTable = components['schemas']['storage_postgres_table'];
type TransformDefinition = components['schemas']['transform_definition'];

interface PipelineCanvasProps {
  mappings: MapperFieldMapping[];
  onMappingsChange: (mappings: MapperFieldMapping[]) => void;
  sourceFields: SourceFieldDefinition[];
  targetTables: StoragePostgresTable[];
  transforms: TransformDefinition[];
  datasourceName?: string;
  storageName?: string;
  fullScreen?: boolean;
}

const nodeTypes = {
  fieldGroup: FieldGroupNode,
};

const edgeTypes = {
  transform: TransformEdge,
};

function generateId(): string {
  return `mapping-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
}

function PipelineCanvas({
  mappings,
  onMappingsChange,
  sourceFields,
  targetTables,
  transforms,
  datasourceName,
  storageName,
  fullScreen = false,
}: PipelineCanvasProps) {
  // Group source fields by entity
  const messageFields = sourceFields.filter((f) => f.entity === 'message');
  const contactFields = sourceFields.filter((f) => f.entity === 'contact');

  // Create nodes for source field groups and target tables
  const initialNodes: Node[] = useMemo(() => {
    const nodes: Node[] = [];
    let sourceY = 50;
    let targetY = 50;
    const nodeSpacing = 40;

    // Source nodes (left side)
    if (messageFields.length > 0) {
      nodes.push({
        id: 'source-message',
        type: 'fieldGroup',
        position: { x: 50, y: sourceY },
        data: {
          title: datasourceName ? `${datasourceName} - Message` : 'Message',
          icon: <InboxOutlined style={{ color: '#fca311' }} />,
          fields: messageFields.map((f) => ({ name: f.name, type: f.type })),
          handleType: 'source',
          handleIdPrefix: 'source-message',
        },
        draggable: fullScreen,
      });
      // Estimate height: header (~50px) + fields (~32px each) + padding
      sourceY += 70 + messageFields.length * 32 + nodeSpacing;
    }

    if (contactFields.length > 0) {
      nodes.push({
        id: 'source-contact',
        type: 'fieldGroup',
        position: { x: 50, y: sourceY },
        data: {
          title: datasourceName ? `${datasourceName} - Contact` : 'Contact',
          icon: <UserOutlined style={{ color: '#fca311' }} />,
          fields: contactFields.map((f) => ({ name: f.name, type: f.type })),
          handleType: 'source',
          handleIdPrefix: 'source-contact',
        },
        draggable: fullScreen,
      });
    }

    // Target nodes (right side) - one per table
    targetTables.forEach((table) => {
      const tableFields = table.fields || [];
      nodes.push({
        id: `target-${table.name}`,
        type: 'fieldGroup',
        position: { x: 450, y: targetY },
        data: {
          title: storageName ? `${storageName} - ${table.name}` : table.name,
          icon: <TableOutlined style={{ color: '#fca311' }} />,
          fields: tableFields.map((f) => ({ name: f.name, type: f.type })),
          handleType: 'target',
          handleIdPrefix: `target-${table.name}`,
        },
        draggable: fullScreen,
      });
      // Estimate height for next node position
      targetY += 70 + tableFields.length * 32 + nodeSpacing;
    });

    return nodes;
  }, [messageFields, contactFields, targetTables, datasourceName, storageName, fullScreen]);

  // Parse handle IDs to extract entity/field info
  const parseSourceHandle = (handleId: string) => {
    const match = handleId.match(/^source-(message|contact)-(.+)$/);
    if (match) {
      return { entity: match[1] as 'message' | 'contact', field: match[2] };
    }
    return null;
  };

  const parseTargetHandle = (handleId: string) => {
    const match = handleId.match(/^target-([^-]+)-(.+)$/);
    if (match) {
      return { table: match[1], field: match[2] };
    }
    return null;
  };

  // Handle transform change for a mapping
  const handleTransformChange = useCallback(
    (mappingId: string, transform: MapperFieldMapping['transform']) => {
      const updated = mappings.map((m) =>
        m.id === mappingId ? { ...m, transform } : m
      );
      onMappingsChange(updated);
    },
    [mappings, onMappingsChange]
  );

  // Handle enabled change for a mapping
  const handleEnabledChange = useCallback(
    (mappingId: string, isEnabled: boolean) => {
      const updated = mappings.map((m) =>
        m.id === mappingId ? { ...m, is_enabled: isEnabled } : m
      );
      onMappingsChange(updated);
    },
    [mappings, onMappingsChange]
  );

  // Handle delete mapping
  const handleDeleteMapping = useCallback(
    (mappingId: string) => {
      const updated = mappings.filter((m) => m.id !== mappingId);
      onMappingsChange(updated);
    },
    [mappings, onMappingsChange]
  );

  // Convert mappings to edges
  const initialEdges: Edge[] = useMemo(
    () =>
      mappings.map((mapping) => ({
        id: mapping.id || generateId(),
        source: `source-${mapping.source_entity}`,
        sourceHandle: `source-${mapping.source_entity}-${mapping.source_field}`,
        target: `target-${mapping.target_table}`,
        targetHandle: `target-${mapping.target_table}-${mapping.target_field}`,
        type: 'transform',
        data: {
          transform: mapping.transform || { type: 'set' },
          isEnabled: mapping.is_enabled ?? true,
          transforms,
          onTransformChange: (transform: MapperFieldMapping['transform']) =>
            handleTransformChange(mapping.id!, transform),
          onEnabledChange: (enabled: boolean) =>
            handleEnabledChange(mapping.id!, enabled),
          onDelete: () => handleDeleteMapping(mapping.id!),
        },
      })),
    [mappings, transforms, handleTransformChange, handleEnabledChange, handleDeleteMapping]
  );

  const [nodes, , onNodesChange] = useNodesState(initialNodes);
  const [edges, setEdges, onEdgesChange] = useEdgesState(initialEdges);

  // Handle new connection
  const onConnect = useCallback(
    (connection: Connection) => {
      if (!connection.sourceHandle || !connection.targetHandle) return;

      const source = parseSourceHandle(connection.sourceHandle);
      const target = parseTargetHandle(connection.targetHandle);

      if (!source || !target) return;

      // Check if mapping already exists
      const exists = mappings.some(
        (m) =>
          m.source_entity === source.entity &&
          m.source_field === source.field &&
          m.target_table === target.table &&
          m.target_field === target.field
      );

      if (exists) {
        message.warning('This mapping already exists');
        return;
      }

      const newMapping: MapperFieldMapping = {
        id: generateId(),
        source_entity: source.entity,
        source_field: source.field,
        transform: { type: 'set' },
        target_table: target.table,
        target_field: target.field,
        is_enabled: true,
      };

      onMappingsChange([...mappings, newMapping]);
    },
    [mappings, onMappingsChange]
  );

  // Sync edges when mappings change externally
  useMemo(() => {
    setEdges(
      mappings.map((mapping) => ({
        id: mapping.id || generateId(),
        source: `source-${mapping.source_entity}`,
        sourceHandle: `source-${mapping.source_entity}-${mapping.source_field}`,
        target: `target-${mapping.target_table}`,
        targetHandle: `target-${mapping.target_table}-${mapping.target_field}`,
        type: 'transform',
        data: {
          transform: mapping.transform || { type: 'set' },
          isEnabled: mapping.is_enabled ?? true,
          transforms,
          onTransformChange: (transform: MapperFieldMapping['transform']) =>
            handleTransformChange(mapping.id!, transform),
          onEnabledChange: (enabled: boolean) =>
            handleEnabledChange(mapping.id!, enabled),
          onDelete: () => handleDeleteMapping(mapping.id!),
        },
      }))
    );
  }, [mappings, transforms, setEdges, handleTransformChange, handleEnabledChange, handleDeleteMapping]);

  return (
    <div
      style={{
        height: fullScreen ? '100%' : 280,
        border: '1px solid #d9d9d9',
        borderRadius: 8,
        overflow: 'hidden',
      }}
    >
      <ReactFlow
        nodes={nodes}
        edges={edges}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        onConnect={onConnect}
        nodeTypes={nodeTypes}
        edgeTypes={edgeTypes}
        fitView
        fitViewOptions={{ padding: 0.2 }}
        nodesDraggable={fullScreen}
        nodesConnectable={fullScreen}
        elementsSelectable={fullScreen}
        panOnDrag={fullScreen}
        zoomOnScroll={fullScreen}
        zoomOnPinch={fullScreen}
        zoomOnDoubleClick={fullScreen}
        preventScrolling={!fullScreen}
        proOptions={{ hideAttribution: true }}
        connectionLineStyle={{ stroke: '#fca311', strokeWidth: 2 }}
      >
        <Background color="#e5e5e5" gap={16} />
        <Controls showInteractive={fullScreen} />
      </ReactFlow>
    </div>
  );
}

export default PipelineCanvas;
