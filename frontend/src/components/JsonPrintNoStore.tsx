import React, { useState } from 'react'

interface JsonPrintNoStoreProps {
  beta?: boolean
  off?: boolean
  icon?: React.ReactElement | null
  iconMode?: boolean
  className?: string
  title?: string
  data?: any
  viewSetting?: any
  style?: React.CSSProperties
}

function JsonPrintNoStore({
  beta = false,
  off = false,
  icon = null,
  iconMode = false,
  className = '',
  title = '',
  data = {},
  viewSetting = {},
  style = {},
}: JsonPrintNoStoreProps) {
  const [offIn, setOff] = useState(off)

  return iconMode && icon ? (
    <div
      style={{
        textAlign: 'left',
        fontSize: '12px',
        lineHeight: '12px',
        border: '1px dotted #333',
        display: beta ? 'block' : 'block',
        ...style,
      }}
      className={className}
    >
      <div
        onClick={(e) => {
          e.preventDefault()
          setOff(!offIn)
        }}
        style={{
          padding: 2,
          fontSize: '12px',
          cursor: 'pointer',
          backgroundColor: '#ebebeb',
          fontWeight: 400,
        }}
      >
        {icon}
      </div>
      {!offIn && (
        <pre>
          <small>{JSON.stringify(data, null, 2)}</small>
        </pre>
      )}
    </div>
  ) : (
    <div
      style={{
        textAlign: 'left',
        fontSize: '12px',
        lineHeight: '12px',
        border: '1px dotted #333',
        display: beta ? 'block' : 'block',
        ...style,
      }}
      className={className}
    >
      <div
        onClick={(e) => {
          e.preventDefault()
          setOff(!offIn)
        }}
        style={{
          padding: 10,
          fontSize: '14px',
          cursor: 'pointer',
          backgroundColor: '#ebebeb',
          fontWeight: 400,
        }}
      >
        {title || (data ? `${JSON.stringify(data).substring(0, 60)}...` : 'no data')}{' '}
      </div>
      {!offIn && (
        <pre>
          <small>{JSON.stringify(data, null, 2)}</small>
        </pre>
      )}
    </div>
  )
}

export default JsonPrintNoStore
