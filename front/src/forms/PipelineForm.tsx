import '@xyflow/react/dist/style.css'
import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { Controller, useForm } from 'react-hook-form'
import { useNavigate } from 'react-router-dom'
import {
  ActionButton,
  Badge,
  Button,
  Cell,
  Column,
  Content,
  Dialog,
  DialogTrigger,
  DropZone,
  Flex,
  Form,
  Heading,
  Item,
  Picker,
  Row,
  TableBody,
  TableHeader,
  TableView,
  Text,
  TextField,
  useDragAndDrop,
  useListData,
  View,
} from '@adobe/react-spectrum'
import { DropEvent } from '@react-types/shared'
import { DropOperation } from '@react-types/shared'
import CalendarAdd from '@spectrum-icons/workflow/CalendarAdd'
import Delete from '@spectrum-icons/workflow/Delete'
import Trash from '@spectrum-icons/workflow/Delete'
import Edit from '@spectrum-icons/workflow/Edit'
import Fast from '@spectrum-icons/workflow/Fast'
import Play from '@spectrum-icons/workflow/Play'
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
import { CustomNode } from '@/components/CustomeNode'
import { SchedulerForm } from '@/forms/SchedulerForm'

interface PipelineProps {
  pipelineUUID: string
  userUUID: string // or accountUUID, whichever your system uses
}

interface FlowEntryItem {
  id: number
  uuid?: string
  parent?: number
  type?: string
  title: string
}

const initialNodes = [
  {
    id: '1',
    type: 'customNode',
    data: { label: 'Scheduler' },
    position: { x: 250, y: 25 },
  },

  {
    id: '2',
    type: 'customNode',
    data: { label: 'Data Source' },
    position: { x: 250, y: 125 },
  },
  {
    id: '3',
    type: 'customNode',
    data: { label: 'Contact Extractor' },
    position: { x: 250, y: 225 },
  },
  {
    id: '4',
    type: 'customNode',
    data: { label: 'Storage S3' },
    position: { x: 250, y: 350 },
  },
]

const initialEdges = [
  { id: 'e1-2', source: '1', target: '2', animated: true },
  { id: 'e2-3', source: '2', target: '3' },
  { id: 'e2-4', source: '3', target: '4' },
]

export const PipelineForm = ({ pipelineUUID, userUUID }: PipelineProps) => {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const form = useForm<components['schemas']['pipeline']>({})
  const rf = useReactFlow()

  // Track current nodes/edges in React Flow
  const [nodes, setNodes] = useState<Node[]>(initialNodes)
  const [edges, setEdges] = useState<Edge[]>(initialEdges)

  // ----------------------- Scheduler local state -----------------------
  const [showSchedulerForm, setShowSchedulerForm] = useState(false)
  const [editingSchedulerUUID, setEditingSchedulerUUID] = useState<'add' | string>('add')

  const openScheduler = (uuid: 'add' | string) => {
    setEditingSchedulerUUID(uuid)
    setShowSchedulerForm(true)
  }

  const closeScheduler = () => {
    setShowSchedulerForm(false)
  }

  const isAdd = pipelineUUID === 'add'

  // ======== Queries & Mutations ========

  const datasourceQuery = useQuery({
    queryKey: ['datasources'],
    queryFn: async ({ signal }) => {
      const { data } = await client.GET('/datasource', { signal })
      return data ?? ([] as components['schemas']['datasource'][])
    },
  })

  const storageQuery = useQuery({
    queryKey: ['storages'],
    queryFn: async ({ signal }) => {
      const { data } = await client.GET('/storage', { signal })
      return data ?? ([] as components['schemas']['storage'][])
    },
  })

  // --- Schedulers for this pipeline ---
  const schedulersQuery = useQuery({
    queryKey: ['schedulers', pipelineUUID],
    enabled: !isAdd,
    queryFn: async ({ signal }) => {
      if (isAdd) return []
      const { data } = await client.GET('/scheduler', {
        params: { query: { pipeline_uuid: pipelineUUID } },
        signal,
      })
      return data as components['schemas']['scheduler'][]
    },
  })

  const deleteSchedulerMutation = useMutation({
    mutationFn: async (uuid: string) => {
      const resp = await client.DELETE('/scheduler/' + uuid)
      if (resp.error) throw new Error(resp.error.detail)
      return resp
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['schedulers', pipelineUUID] })
    },
  })

  const pipelineQuery = useQuery({
    queryKey: ['/pipeline/{uuid}', { uuid: pipelineUUID, user_uuid: userUUID }],
    queryFn: async ({ signal }) => {
      const { data } = await client.GET('/pipeline/{uuid}', {
        params: {
          query: {
            user_uuid: userUUID,
          },
          path: { uuid: pipelineUUID },
        },
        signal,
      })
      return data ?? ({} as components['schemas']['pipeline'])
    },
    // throwOnError: false,
    enabled: pipelineUUID !== 'add',
  })

  const pipelineMutation = useMutation({
    mutationFn: async (data: components['schemas']['pipeline']) => {
      let resp
      if (pipelineUUID === 'add') {
        resp = await client.POST('/pipeline', {
          body: {
            name: data.name || '',
            datasource_uuid: data.datasource_uuid || '',
            storage_uuid: data.storage_uuid || '',
            flow: data.flow || {},
          },
        })
      } else {
        resp = await client.PUT(`/pipeline/{uuid}`, {
          params: { path: { uuid: pipelineUUID } }, // TODO @reactima , query: { user_uuid: userUUID }
          body: {
            name: data.name,
            datasource_uuid: data.datasource_uuid,
            storage_uuid: data.storage_uuid,
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

  useEffect(() => {
    console.log('useEffect pipelineQuery', { pipelineQuery })
    // If no data, clear and return
    // if (!entryTypesQuery.data) {
    //   if (flowEntries.items.length > 0) {
    //     flowEntries.remove(...flowEntries.items.map((item) => item.id))
    //   }
    //   return
    // }
    // If we have data, remove existing items, then rebuild
    if (flowEntries.items.length > 0) {
      flowEntries.remove(...flowEntries.items.map((item) => item.id))
    }

    // Now rebuild
    let idx = 0
    const addLeaf = (categoryName: string, label: string) => {
      idx++
      const parentId = idx
      // const hasEntries = entryTypesQuery.data.entries.some((e) => e.category === categoryName)
      // const displayTitle = hasEntries ? label : `${label} (No entries)`

      // Append "parent" category
      // flowEntries.append({ id: parentId, title: displayTitle })
      //
      // // If none, stop
      // if (!hasEntries) return

      // Child items for this category
      // entryTypesQuery.data.entries
      //   .filter((entry) => entry.category === categoryName)
      //   .forEach((entryType) => {
      //     idx++
      //     flowEntries.append({
      //       id: idx,
      //       parent: parentId,
      //       uuid: entryType.uuid,
      //       title: entryType.name,
      //       type: entryType.flow_type,
      //     })
      //   })
    }

    ;['datasource', 'extractor', 'filter', 'mapper', 'storage'].forEach((category) => {
      addLeaf(category, category.charAt(0).toUpperCase() + category.slice(1) + 's')
    })
    // flowEntries removed from deps
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [pipelineQuery.data])

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

  const nodeTypes = useMemo(() => ({ customNode: CustomNode }), [])

  if (datasourceQuery.isPending && storageQuery.isPending && !isAdd) return <></>

  console.log('PipelineForm!', { datasourceQuery, initialEdges, initialNodes, nodes, edges })

  return (
    <Flex direction="row" width="100%" height="100%">
      {/* Left: Form section with fixed width */}
      <View width="50%" padding="size-500" borderEndWidth="thin" borderColor="dark">
        <Form onSubmit={form.handleSubmit(onSubmit)} aria-label="Update flow">
          <Flex direction="column" rowGap="size-200">
            <Flex direction="row" gap="size-200">
              <Controller
                name="name"
                control={form.control}
                rules={{ required: 'Name is required' }}
                render={({ field, fieldState }) => (
                  <TextField
                    labelPosition="side"
                    isRequired
                    value={field.value ?? ''}
                    onChange={field.onChange}
                    onBlur={field.onBlur}
                    validationState={fieldState.invalid ? 'invalid' : undefined}
                    errorMessage={fieldState.error?.message}
                    aria-label={`Enter ${field.name} value`}
                  />
                )}
              />

              <ActionButton aria-label="Save or add pipeline" onPress={() => form.handleSubmit(onSubmit)()}>
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

            <Controller
              name="datasource_uuid"
              control={form.control}
              rules={{ required: 'Datasource is required' }}
              render={({ field, fieldState }) => (
                <Picker
                  label="Data Source"
                  isRequired
                  selectedKey={field.value}
                  onSelectionChange={(key) => field.onChange(key ? key.toString() : '')}
                  errorMessage={fieldState.error?.message}
                  width="100%"
                >
                  {datasourceQuery?.data?.map((datasource: components['schemas']['datasource']) => 
                    <Item key={datasource.uuid} textValue={`${datasource.name} ${datasource.type}`}>
                      <span
                        style={{
                          whiteSpace: 'nowrap',
                          margin: '0 10px',
                          lineHeight: '24px',
                        }}
                      >
                        {datasource.name} {datasource.type}
                      </span>
                    </Item>
                  )}
                </Picker>
              )}
            />

            <Controller
              name="contact_extractor"
              control={form.control}
              rules={{ required: 'Datasource is required' }}
              render={({ field, fieldState }) => (
                <Picker
                  label="Contact Extractor"
                  isRequired
                  selectedKey={field.value}
                  onSelectionChange={(key) => field.onChange(key.toString())}
                  errorMessage={fieldState.error?.message}
                  width="100%"
                >
                  <Item key="extractor1">
                    <span style={{ whiteSpace: 'nowrap', margin: '0 10px', lineHeight: '24px' }}>
                      Standard Extractor
                    </span>
                  </Item>
                </Picker>
              )}
            />

            <Controller
              name="storage_uuid"
              control={form.control}
              rules={{ required: 'Storage is required' }}
              render={({ field, fieldState }) => (
                <Picker
                  label="Storage"
                  isRequired
                  selectedKey={field.value}
                  onSelectionChange={(key) => field.onChange(key.toString())}
                  errorMessage={fieldState.error?.message}
                  width="100%"
                >
                  {storageQuery?.data?.map((storage: components['schemas']['storage']) => (
                    <Item key={storage.uuid}>
                      <span style={{ whiteSpace: 'nowrap', margin: '0 10px', lineHeight: '24px' }}>
                        {storage.name} {storage.type}
                      </span>
                    </Item>
                  ))}
                </Picker>
              )}
            />

            <Flex direction="row" gap="size-200">
              <ActionButton
                onPress={() => {}}
                aria-label="Run pipeline"
                isDisabled={deleteMutation.isPending || pipelineQuery.isFetching || pipelineUUID === 'add'}
              >
                <span style={{ display: 'flex', alignItems: 'center', gap: 2 }}>
                  Run
                  <Play />
                </span>
              </ActionButton>
              <ActionButton
                onPress={() => {}}
                aria-label="Run pipeline"
                isDisabled={deleteMutation.isPending || pipelineQuery.isFetching || pipelineUUID === 'add'}
              >
                <span style={{ display: 'flex', alignItems: 'center', gap: 2 }}>
                  Test <Fast />
                </span>
              </ActionButton>
              <ActionButton
                onPress={() => openScheduler('add')}
                aria-label="Schedule pipeline"
                isDisabled={deleteMutation.isPending || pipelineQuery.isFetching || pipelineUUID === 'add'}
              >
                <span style={{ display: 'flex', alignItems: 'center', gap: 2 }}>
                  Schedule <CalendarAdd />
                </span>
              </ActionButton>
            </Flex>

            {/* --- Schedulers Table --- */}
            {!isAdd && (
              <View marginTop="size-300">
                <Heading level={3}>Schedulers</Heading>
                <TableView aria-label="Schedulers list" overflowMode="wrap">
                  <TableHeader>
                    <Column key="type">Type</Column>
                    <Column key="expr">Cron / RunAt</Column>
                    <Column key="next">Next Run</Column>
                    <Column key="enabled">Enabled</Column>
                    <Column key="actions" width={60}>
                      Actions
                    </Column>
                  </TableHeader>
                  <TableBody items={schedulersQuery.data ?? []}>
                    {(item: components['schemas']['scheduler']) => (
                      <Row key={item.uuid!}>
                        <Cell>{item.schedule_type}</Cell>
                        <Cell>
                          {item.schedule_type === 'cron'
                            ? item.cron_expression
                            : new Date(item.run_at ?? '').toLocaleString()}
                        </Cell>
                        <Cell>{item.next_run ? new Date(item.next_run).toLocaleString() : '—'}</Cell>
                        <Cell>
                          <Badge variant={item.is_enabled ? 'positive' : 'negative'}>
                            {item.is_enabled ? 'On' : 'Off'}
                          </Badge>
                        </Cell>
                        <Cell>
                          <Flex direction="row" gap="size-75">
                            <ActionButton aria-label="Edit" isQuiet onPress={() => openScheduler(item.uuid!)}>
                              <Edit />
                            </ActionButton>
                            <ActionButton
                              aria-label="Remove"
                              isQuiet
                              onPress={() => deleteSchedulerMutation.mutate(item.uuid!)}
                            >
                              <Trash />
                            </ActionButton>
                          </Flex>
                        </Cell>
                      </Row>
                    )}
                  </TableBody>
                </TableView>
              </View>
            )}

            {/* Inline Scheduler Form in a modal */}
            <DialogTrigger isOpen={showSchedulerForm} type="modal" onOpenChange={setShowSchedulerForm}>
              <></>
              <Dialog isDismissable onDismiss={closeScheduler} width="size-6000">
                <Content>
                  {/* Pass schedulerUUID, SchedulerForm will handle add/edit */}
                  <SchedulerForm schedulerUUID={editingSchedulerUUID} />
                  <Flex direction="row" gap="size-200" marginTop="size-200" justifyContent="end">
                    <ActionButton variant="secondary" onPress={closeScheduler}>
                      Close
                    </ActionButton>
                  </Flex>
                </Content>
              </Dialog>
            </DialogTrigger>
          </Flex>
        </Form>
      </View>

      {/* Right: Flow section taking remaining space */}
      <Flex direction="row" flexGrow={1} overflow="auto">
        {/* Left column: FlowEntries */}
        {/*<View padding="size-200" width="240px" overflow="auto">*/}
        {/*  <FlowEntries list={flowEntries} dragAndDropOptions={dragAndDropHooks} />*/}
        {/*</View>*/}

        {/* Right column: DropZone + ReactFlow */}
        <DropZone
          isFilled
          onDrop={onDrop}
          flexGrow={1}
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
            nodeTypes={nodeTypes}
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
