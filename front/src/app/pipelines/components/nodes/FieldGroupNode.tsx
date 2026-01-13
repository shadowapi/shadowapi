import { memo } from 'react';
import { Handle, Position, type NodeProps } from '@xyflow/react';
import { Card, Typography } from 'antd';
import { colors } from '../../../../theme';

const { Text } = Typography;

interface FieldInfo {
  name: string;
  type: string;
}

interface FieldGroupNodeData {
  title: string;
  icon?: React.ReactNode;
  fields: FieldInfo[];
  handleType: 'source' | 'target';
  handleIdPrefix: string;
}

function FieldGroupNode({ data }: NodeProps) {
  const nodeData = data as unknown as FieldGroupNodeData;
  const { title, icon, fields, handleType, handleIdPrefix } = nodeData;

  const isSource = handleType === 'source';

  return (
    <Card
      size="small"
      title={
        <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
          {icon}
          <Text strong style={{ color: colors.oxfordBlue }}>
            {title}
          </Text>
        </div>
      }
      style={{
        minWidth: 200,
        borderColor: colors.oxfordBlue,
        borderWidth: 2,
      }}
      styles={{
        body: { padding: 8 },
        header: { background: '#f5f5f5', minHeight: 36, padding: '0 12px' },
      }}
    >
      {fields.map((field) => (
        <div
          key={field.name}
          style={{
            position: 'relative',
            padding: '4px 8px',
            marginBottom: 4,
            background: '#fafafa',
            borderRadius: 4,
            fontSize: 12,
          }}
        >
          {!isSource && (
            <Handle
              type="target"
              position={Position.Left}
              id={`${handleIdPrefix}-${field.name}`}
              style={{
                background: colors.orange,
                width: 8,
                height: 8,
                left: -4,
              }}
            />
          )}
          <Text>{field.name}</Text>
          <Text type="secondary" style={{ marginLeft: 4, fontSize: 10 }}>
            ({field.type})
          </Text>
          {isSource && (
            <Handle
              type="source"
              position={Position.Right}
              id={`${handleIdPrefix}-${field.name}`}
              style={{
                background: colors.orange,
                width: 8,
                height: 8,
                right: -4,
              }}
            />
          )}
        </div>
      ))}
    </Card>
  );
}

export default memo(FieldGroupNode);
