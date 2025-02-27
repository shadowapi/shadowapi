import '@xyflow/react/dist/style.css'
import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { Controller, useForm } from 'react-hook-form'
import { useNavigate } from 'react-router-dom'
import { ActionButton, DropZone, Flex, Form, TextField, useDragAndDrop, useListData, View } from '@adobe/react-spectrum'
import { DropEvent } from '@react-types/shared'
import { DropOperation } from '@react-types/shared'
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

  // Track current nodes/edges in React Flow
  const [nodes, setNodes] = useState<Node[]>([])
  const [edges, setEdges] = useState<Edge[]>([])

  // ======== Queries & Mutations ========
  const entryTypesQuery = useQuery({
    queryKey: ['/pipeline/entry/types'],
    queryFn: async ({ signal }) => {
      const { data } = await client.GET('/pipeline/entry/types', { signal })
      return data
    },
    throwOnError: false,
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
    throwOnError: false,
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
        // Refresh just this pipeline
        queryClient.setQueryData(['/pipeline/{uuid}', { uuid: variable.uuid }], data)
      }
    },
  })

  const deleteMutation = useMutation({
    mutationFn: async (uuid: string) => {
      const resp = await client.DELETE(`/pipeline/{uuid}`, {
        params: { path: { uuid } },
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

  // --------------------- FLOWENTRIES DATA (ListView) ---------------------
  const flowEntries = useListData<FlowEntryItem>({})

  // Only rebuild the flowEntries data when entryTypesQuery.data changes
  useEffect(() => {
    // If no data, clear and return
    if (!entryTypesQuery.data) {
      if (flowEntries.items.length > 0) {
        flowEntries.remove(...flowEntries.items.map((item) => item.id))
      }
      return
    }
    // If we have data, remove existing items, then rebuild
    if (flowEntries.items.length > 0) {
      flowEntries.remove(...flowEntries.items.map((item) => item.id))
    }

    // Now rebuild
    let idx = 0
    const addLeaf = (categoryName: string, label: string) => {
      idx++
      const parentId = idx
      const hasEntries = entryTypesQuery.data.entries.some((e) => e.category === categoryName)
      const displayTitle = hasEntries ? label : `${label} (No entries)`

      // Append "parent" category
      flowEntries.append({ id: parentId, title: displayTitle })

      // If none, stop
      if (!hasEntries) return

      // Child items for this category
      entryTypesQuery.data.entries
        .filter((entry) => entry.category === categoryName)
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
    // flowEntries removed from deps
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [entryTypesQuery.data])

  // --------------------- PIPELINE DATA & REACT FLOW ---------------------
  // Populate form and React Flow with pipeline data
  useEffect(() => {
    if (pipelineQuery.data) {
      form.reset(pipelineQuery.data)
      rf.addNodes(pipelineQuery.data.flow.nodes as Node[])
      rf.addEdges(pipelineQuery.data.flow.edges as Edge[])
    }
    // Only run this effect when pipeline data first arrives or changes
  }, [pipelineQuery.data, rf, form])

  // React Flow event handlers
  const onNodesChange = useCallback((changes: NodeChange<Node>[]) => {
    setNodes((nds) => applyNodeChanges(changes, nds))
  }, [])

  const onEdgesChange = useCallback((changes: EdgeChange<Edge>[]) => {
    setEdges((eds) => applyEdgeChanges(changes, eds))
  }, [])

  const onConnect = useCallback((connection: Connection) => {
    setEdges((oldEdges) => addEdge(connection, oldEdges))
  }, [])

  // --------------------- SUBMISSION HANDLERS ---------------------
  const onSubmit = async (data: components['schemas']['pipeline']) => {
    // If flow is missing or undefined, set it to an empty object
    if (!data.flow) {
      data.flow = {}
    }

    data.flow.nodes = nodes
    data.flow.edges = edges

    pipelineMutation.mutate(data, {
      onError: (error) => {
        console.error(error)
      },
      onSuccess: () => {
        navigate(`/pipelines`)
      },
    })
  }

  const onDelete = () => {
    deleteMutation.mutate(pipelineUUID, {
      onSuccess: () => {
        navigate(`/pipelines`)
      },
    })
  }

  // --------------------- DRAG & DROP (FROM ListView) ---------------------
  // Memoize the config object so it's stable each render
  const getItemsCallback = useCallback(
    (keys: Iterable<string | number>) => {
      return [...keys].map((key) => {
        const item = flowEntries.getItem(key)
        // We attach a custom drag format with JSON data
        return {
          'custom-app-type': JSON.stringify(item),
        }
      })
    },
    [flowEntries]
  )

  const stableDnDOptions = useMemo(
    () => ({
      getAllowedDropOperations: (): DropOperation[] => ['copy' as DropOperation],
      getItems: getItemsCallback,
    }),
    [getItemsCallback]
  )

  // Provide stable options to useDragAndDrop
  const { dragAndDropHooks } = useDragAndDrop(stableDnDOptions)

  // --------------------- DROP ZONE (ReactFlow area) ---------------------
  // We use a ref to help us calculate the offset for dropping
  const dropZoneRef = useRef<HTMLDivElement>(null)

  const onDrop = async (event: DropEvent) => {
    if (event.type !== 'drop' || event.dropOperation !== 'copy') return
    if (event.items[0].kind !== 'text') return

    // Parse data from the drag item
    const data = JSON.parse(await event.items[0].getText('custom-app-type'))

    // Adjust for the DropZone offset
    if (dropZoneRef.current) {
      // Adjust event.x / event.y to account for the DropZone offset
      const rect = dropZoneRef.current.getBoundingClientRect()
      event.x = event.x + rect.x
      event.y = event.y + rect.y
    }

    // Create a new node at the dropped position
    const newNode: Node = {
      id: `${nodes.length + 1}`,
      position: rf.screenToFlowPosition({ x: event.x, y: event.y }),
      data: { label: data.title },
      type: data.type,
    }
    setNodes((prev) => [...prev, newNode])
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
                  labelPosition="side"
                  isRequired
                  // Ensure we always have a string instead of undefined
                  value={field.value ?? ''}
                  // Wire up React Hook Form's onChange/onBlur
                  onChange={field.onChange}
                  onBlur={field.onBlur}
                  // Mark invalid in Spectrum if there's an error
                  validationState={fieldState.invalid ? 'invalid' : undefined}
                  errorMessage={fieldState.error?.message}
                  aria-label={`Enter ${field.name} value`}
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
          isFilled
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
