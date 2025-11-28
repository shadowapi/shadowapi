import { memo } from 'react'
import { Handle, NodeProps, Position } from '@xyflow/react'
export const CustomNode = memo(({ data }: NodeProps) => {
  return (
    <div
      style={{
        padding: '4px 8px',
        fontSize: '12px',
        background: '#f0f0f0',
        border: '1px solid #999',
        borderRadius: '4px',
        width: '100px',
        textAlign: 'center',
      }}
    >
      <Handle type="target" position={Position.Top} style={{ background: '#555' }} />
      <div style={{ fontSize: 12 }}>{data.label}</div>
      <Handle type="source" position={Position.Bottom} style={{ background: '#555' }} />
    </div>
  )
})
