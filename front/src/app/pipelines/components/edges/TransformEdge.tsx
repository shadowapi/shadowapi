import { memo, useState } from 'react';
import {
  BaseEdge,
  EdgeLabelRenderer,
  getBezierPath,
  type EdgeProps,
} from '@xyflow/react';
import { Popover, Select, Typography, Space, Button, Checkbox } from 'antd';
import { SettingOutlined, DeleteOutlined } from '@ant-design/icons';
import { colors } from '../../../../theme';
import type { components } from '../../../../api/v1';

const { Text } = Typography;

type MapperTransform = components['schemas']['mapper_transform'];
type TransformDefinition = components['schemas']['transform_definition'];

interface TransformEdgeData {
  transform: MapperTransform;
  isEnabled: boolean;
  transforms: TransformDefinition[];
  onTransformChange?: (transform: MapperTransform) => void;
  onEnabledChange?: (enabled: boolean) => void;
  onDelete?: () => void;
}

function TransformEdge({
  id,
  sourceX,
  sourceY,
  targetX,
  targetY,
  sourcePosition,
  targetPosition,
  data,
  selected,
}: EdgeProps) {
  const edgeData = data as unknown as TransformEdgeData;
  const [popoverOpen, setPopoverOpen] = useState(false);

  const [edgePath, labelX, labelY] = getBezierPath({
    sourceX,
    sourceY,
    sourcePosition,
    targetX,
    targetY,
    targetPosition,
  });

  const transformType = edgeData?.transform?.type || 'set';
  const transformDef = edgeData?.transforms?.find((t) => t.type === transformType);
  const displayName = transformDef?.display_name || transformType;

  const transformOptions = (edgeData?.transforms || []).map((t) => ({
    value: t.type,
    label: t.display_name,
  }));

  const popoverContent = (
    <Space direction="vertical" size="small" style={{ width: 200 }}>
      <div>
        <Text type="secondary" style={{ fontSize: 12 }}>
          Transform
        </Text>
        <Select
          value={transformType}
          options={transformOptions}
          onChange={(type) => {
            edgeData?.onTransformChange?.({
              type: type as NonNullable<MapperTransform>['type'],
            });
          }}
          style={{ width: '100%' }}
          size="small"
        />
      </div>
      <div>
        <Checkbox
          checked={edgeData?.isEnabled ?? true}
          onChange={(e) => edgeData?.onEnabledChange?.(e.target.checked)}
        >
          Enabled
        </Checkbox>
      </div>
      <Button
        type="text"
        danger
        size="small"
        icon={<DeleteOutlined />}
        onClick={() => {
          setPopoverOpen(false);
          edgeData?.onDelete?.();
        }}
        style={{ width: '100%' }}
      >
        Remove Mapping
      </Button>
    </Space>
  );

  return (
    <>
      <BaseEdge
        id={id}
        path={edgePath}
        style={{
          stroke: edgeData?.isEnabled === false ? '#d9d9d9' : colors.orange,
          strokeWidth: selected ? 3 : 2,
          strokeDasharray: edgeData?.isEnabled === false ? '5,5' : undefined,
        }}
      />
      <EdgeLabelRenderer>
        <Popover
          content={popoverContent}
          title="Configure Mapping"
          trigger="click"
          open={popoverOpen}
          onOpenChange={setPopoverOpen}
        >
          <div
            style={{
              position: 'absolute',
              transform: `translate(-50%, -50%) translate(${labelX}px,${labelY}px)`,
              pointerEvents: 'all',
              cursor: 'pointer',
              background: edgeData?.isEnabled === false ? '#f5f5f5' : colors.white,
              border: `1px solid ${edgeData?.isEnabled === false ? '#d9d9d9' : colors.orange}`,
              borderRadius: 4,
              padding: '2px 8px',
              fontSize: 11,
              display: 'flex',
              alignItems: 'center',
              gap: 4,
            }}
            className="nodrag nopan"
          >
            <SettingOutlined style={{ fontSize: 10, color: colors.orange }} />
            <Text style={{ color: edgeData?.isEnabled === false ? '#8c8c8c' : undefined }}>
              {displayName}
            </Text>
          </div>
        </Popover>
      </EdgeLabelRenderer>
    </>
  );
}

export default memo(TransformEdge);
