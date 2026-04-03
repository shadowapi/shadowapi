import React, { useMemo } from 'react'
import { List, Typography } from 'antd'

interface FlowItem {
  id: number
  uuid?: string
  type?: string
  title: string
  parent?: number
}

export interface FlowEntriesProp {
  items: FlowItem[]
}

type RenderList = {
  [key: number]: FlowItem & { children: FlowItem[] }
}

// TODO @reactima research on React.memo use here
export const FlowEntries = React.memo(function FlowEntries(props: FlowEntriesProp) {
  const { items } = props

  const renderList = useMemo<RenderList>(() => {
    const tmp: RenderList = {}
    if (!items) return tmp

    for (const item of items) {
      if (item.parent) {
        if (!tmp[item.parent]) {
          tmp[item.parent] = { id: item.parent, title: '', children: [] }
        }
        tmp[item.parent].children.push(item)
      } else {
        tmp[item.id] = { ...item, children: [] }
      }
    }
    return tmp
  }, [items])

  if (!items || items.length === 0) {
    return <>No FlowEntries or empty</>
  }

  return (
    <>
      {Object.values(renderList).map((node) => (
        <div key={node.id}>
          <Typography.Title level={5}>{node.title}</Typography.Title>
          <List
            size="small"
            dataSource={node.children}
            renderItem={(child: FlowItem) => (
              <List.Item key={child.id}>{child.title}</List.Item>
            )}
          />
        </div>
      ))}
    </>
  )
})
