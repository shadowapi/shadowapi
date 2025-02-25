import '@xyflow/react/dist/style.css'
import { useCallback, useEffect, useRef, useState } from 'react'
import { Controller, useForm } from 'react-hook-form'
import { useNavigate } from 'react-router-dom'
import { ActionButton, DropZone, Flex, Form, TextField, useDragAndDrop, useListData, View } from '@adobe/react-spectrum'
import { DropEvent } from '@react-types/shared'
import Delete from '@spectrum-icons/workflow/Delete'
import SaveAsFloppy from '@spectrum-icons/workflow/SaveAsFloppy'
import SaveFloppy from '@spectrum-icons/workflow/SaveFloppy'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  addEdge,
  applyEdgeChanges,
  applyNodeChanges,
  Background,
  BackgroundVariant,
  Connection,
  Controls,
  Edge,
  EdgeChange,
  Node,
  NodeChange,
  ReactFlow,
  useReactFlow,
} from '@xyflow/react'

import client from '@/api/client'
import type { components } from '@/api/v1'
import { FlowEntries } from '@/components/FlowEntries'

interface PipelineProps {
  pipelineUUID: string
}

interface FlowEntryItem {
  id: number
  uuid?: string
  parent?: number
  type?: string
  title: string
}

export const PipelineForm = ({ pipelineUUID }: PipelineProps) => {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const form = useForm<components['schemas']['pipeline']>({})
  const rf = useReactFlow()
  const [nodes, setNodes] = useState<Node[]>([])
  const [edges, setEdges] = useState<Edge[]>([])
  const entryTypesQuery = useQuery({
    queryKey: ['/pipeline/entry/types'],
    queryFn: async ({ signal }) => {
      const { data } = await client.GET('/pipeline/entry/types', { signal })
      return data
    },
  })
  const pipelineQuery = useQuery({
    queryKey: ['/pipeline/{uuid}', { id: pipelineUUID }],
    queryFn: async ({ signal }) => {
      const { data } = await client.GET('/pipeline/{uuid}', {
        params: { path: { uuid: pipelineUUID } },
        signal,
      })
      return data
    },
    enabled: pipelineUUID !== 'add',
  })
  const pipelineMutation = useMutation({
    mutationFn: async (data: components['schemas']['pipeline']) => {
      let resp
      if (pipelineUUID === 'add') {
        resp = await client.POST('/pipeline', {
          body: {
            name: data.name || '',
            flow: data.flow || {},
          },
        })
      } else {
        resp = await client.PUT(`/pipeline/{uuid}`, {
          params: { path: { uuid: pipelineUUID } },
          body: {
            name: data.name,
            flow: data.flow,
          },
        })
      }
      if (resp.error) {
        form.setError('name', { message: resp.error.detail })
        throw new Error(resp.error.detail)
      }
      return resp
    },
    onSuccess: (data, variable) => {
      if (pipelineUUID === 'add') {
        queryClient.invalidateQueries({ queryKey: '/pipeline' })
      } else {
        queryClient.setQueryData(['/pipeline/{uuid}', { uuid: variable.uuid }], data)
      }
    },
  })
  const deleteMutation = useMutation({
    mutationFn: async (uuid: string) => {
      const resp = await client.DELETE(`/pipeline/{uuid}`, {
        params: { path: { uuid: uuid } },
      })
      if (resp.error) {
        form.setError('name', { message: resp.error.detail })
        throw new Error(resp.error.detail)
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: '/pipeline' })
    },
  })

  const flowEntries = useListData<FlowEntryItem>({})

  const processEntryTypes = useCallback(() => {
    if (!entryTypesQuery.data) return

    flowEntries.items = []

    let idx = 0

    const addLeaf = (name: string, title: string) => {
      idx++
      const parentId = idx
      const hasEntries = entryTypesQuery.data?.entries.some((e) => e.category === name)
      const displayTitle = hasEntries ? title : `${title} (No entries)`

      flowEntries.append({ id: parentId, title: displayTitle })

      if (!hasEntries) return

      entryTypesQuery.data?.entries
        .filter((entry) => entry.category === name)
        .forEach((entryType) => {
          idx++
          flowEntries.append({
            id: idx,
            parent: parentId,
            uuid: entryType.uuid,
            title: entryType.name,
            type: entryType.flow_type,
          })
        })
    }

    ;['datasource', 'extractor', 'filter', 'mapper', 'storage'].forEach((category) => {
      addLeaf(category, category.charAt(0).toUpperCase() + category.slice(1) + 's')
    })
  }, [entryTypesQuery.data, flowEntries])

  useEffect(() => {
    processEntryTypes()
  }, [processEntryTypes])

  useEffect(() => {
    if (pipelineQuery.data) {
      form.reset(pipelineQuery.data)
      rf.addNodes(pipelineQuery.data.flow.nodes as Node[])
      rf.addEdges(pipelineQuery.data.flow.edges as Edge[])
    }
  }, [pipelineQuery.data, rf, form])

  const onSubmit = async (data: components['schemas']['pipeline']) => {
    data.flow['nodes'] = nodes
    data.flow['edges'] = edges
    pipelineMutation.mutate(data, {
      onError: (error) => {
        console.error(error)
      },
      onSuccess: () => {
        navigate(`/pipelines`)
      },
    })
  }

  const onDelete = async () => {
    deleteMutation.mutate(pipelineUUID, {
      onSuccess: () => {
        navigate(`/pipelines`)
      },
    })
  }

  const { dragAndDropHooks } = useDragAndDrop({
    getAllowedDropOperations: () => ['copy'],
    getItems: (keys) =>
      [...keys].map((key) => {
        const item = flowEntries.getItem(key)
        // Setup the drag types and associated info for each dragged item.
        return {
          'custom-app-type': JSON.stringify(item),
        }
      }),
  })

  const onNodesChange = useCallback((changes: NodeChange<Node>[]) => {
    setNodes((nds) => applyNodeChanges(changes, nds))
  }, [])
  const onEdgesChange = useCallback((changes: EdgeChange<Edge>[]) => {
    setEdges((eds) => applyEdgeChanges(changes, eds))
  }, [])
  const onConnect = useCallback(
    (connection: Connection) => {
      setEdges((oldEdges) => addEdge(connection, oldEdges))
    },
    [setEdges]
  )

  const dropZoneRef = useRef<HTMLDivElement>(null)

  const onDrop = async (event: DropEvent) => {
    if (event.type !== 'drop') return
    if (event.dropOperation !== 'copy') return
    if (event.items[0].kind !== 'text') return

    const data = JSON.parse(await event.items[0].getText('custom-app-type'))
    if (dropZoneRef.current) {
      const rect = dropZoneRef.current.getBoundingClientRect()
      event.x = event.x + rect.x
      event.y = event.y + rect.y
    }
    const newNode = {
      id: `${nodes.length + 1}`,
      position: rf.screenToFlowPosition({ x: event.x, y: event.y }),
      data: { label: data.title },
      type: data.type,
    }
    setNodes([...nodes, newNode])
  }

  return (
    <Flex direction="column" flexBasis="100%" rowGap="size-200">
      <View padding="size-75">
        <Form onSubmit={form.handleSubmit(onSubmit)} aria-label="Update flow">
          <Flex direction="row" width="size-4600" marginStart="size-400" gap="size-200">
            <Controller
              name="name"
              control={form.control}
              rules={{ required: 'Name is required' }}
              render={({ field, fieldState }) => (
                <TextField
                  type="text"
                  width="100%"
                  isRequired
                  labelPosition="side"
                  validationState={fieldState.invalid ? 'invalid' : undefined}
                  aria-label={`Enter ${field.name} value`}
                  errorMessage={fieldState.error?.message}
                  {...field}
                />
              )}
            />
            <ActionButton
              aria-label="Save or add pipeline"
              onPress={() => {
                form.handleSubmit(onSubmit)()
              }}
            >
              {pipelineUUID !== 'add' ? <SaveFloppy /> : <SaveAsFloppy />}
            </ActionButton>

            <ActionButton
              onPress={onDelete}
              aria-label="Delete pipeline"
              isDisabled={deleteMutation.isPending || pipelineQuery.isFetching || pipelineUUID === 'add'}
            >
              <Delete />
            </ActionButton>
          </Flex>
        </Form>
      </View>
      <Flex direction="row" flexGrow={1}>
        <View padding="size-200" minWidth="size-4600" marginStart="size-300">
          <FlowEntries list={flowEntries} dragAndDropOptions={dragAndDropHooks} />
        </View>
        <DropZone
          isFilled={true}
          onDrop={onDrop}
          flexBasis="100%"
          aria-label="Drop pipeline step here"
          UNSAFE_style={{ padding: 0 }}
          replaceMessage="Add pipeline step here"
        >
          <ReactFlow
            nodes={nodes}
            edges={edges}
            onNodesChange={onNodesChange}
            onEdgesChange={onEdgesChange}
            onConnect={onConnect}
            ref={dropZoneRef}
            fitView
          >
            <Controls />
            <Background id="1" gap={10} variant={BackgroundVariant.Lines} />
          </ReactFlow>
        </DropZone>
      </Flex>
    </Flex>
  )
}
