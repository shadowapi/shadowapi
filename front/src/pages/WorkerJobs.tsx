import React, { useState } from 'react'
import {
  ActionButton,
  Cell,
  Column,
  Divider,
  Flex,
  Heading,
  Row,
  TableBody,
  TableHeader,
  TableView,
  Text,
  View,
} from '@adobe/react-spectrum'
import Edit from '@spectrum-icons/workflow/Edit'
import { useQuery } from '@tanstack/react-query'

import client from '@/api/client'
import type { components } from '@/api/v1'
import { FullLayout } from '@/layouts/FullLayout'

export function WorkerJobs() {
  // Store paging values
  const [offset, setOffset] = useState(0)
  const [limit, setLimit] = useState(20)

  // For preview panel
  const [selected, setSelected] = useState<components['schemas']['worker_jobs'] | null>(null)

  // Query worker jobs with offset and limit; assuming the GET /workerjobs endpoint returns a payload like: { jobs: [...] }
  const workerQuery = useQuery({
    queryKey: ['workerjobs', offset, limit],
    queryFn: async () => {
      const { data } = await client.GET('/workerjobs', {
        params: {
          query: {
            offset,
            limit,
          },
        },
      })
      return data?.jobs as components['schemas']['worker_jobs'][]
    },
  })

  const jobs = workerQuery.data || []

  return (
    <FullLayout>
      <Flex direction="row" gap="size-200" margin="size-200" flex>
        {/* LEFT SIDE: Worker Jobs Table */}
        <View flex={1} minWidth={0}>
          <Heading level={3} marginBottom="size-100">
            Worker Jobs
          </Heading>
          <TableView aria-label="Worker Jobs Table" overflowMode="wrap" maxHeight="size-6000">
            <TableHeader>
              <Column key="subject">Subject</Column>
              <Column key="status">Status</Column>
              <Column key="startedAt">Started</Column>
              <Column key="finishedAt">Finished</Column>
              <Column key="actions" align="end">
                Actions
              </Column>
            </TableHeader>
            <TableBody items={jobs}>
              {(item) => (
                <Row key={item.uuid}>
                  <Cell>{item.subject}</Cell>
                  <Cell>{item.status}</Cell>
                  <Cell>{item.started_at ? item.started_at.slice(0, 19) : ''}</Cell>
                  <Cell>{item.finished_at ? item.finished_at.slice(0, 19) : ''}</Cell>
                  <Cell>
                    <ActionButton onPress={() => setSelected(item)}>
                      <Edit />
                      <Text>Preview</Text>
                    </ActionButton>
                  </Cell>
                </Row>
              )}
            </TableBody>
          </TableView>

          {/* Paging controls */}
          <Flex marginTop="size-200" gap="size-100">
            {offset > 0 && (
              <ActionButton onPress={() => setOffset(Math.max(0, offset - limit))} isDisabled={offset === 0}>
                Previous
              </ActionButton>
            )}
            <ActionButton onPress={() => setOffset(offset + limit)}>Next</ActionButton>
          </Flex>
        </View>

        {/* RIGHT SIDE: Preview Panel */}
        {selected && (
          <View
            backgroundColor="gray-100"
            padding="size-200"
            borderStartWidth="thin"
            borderColor="dark"
            overflow="auto"
            width="460px"
            flexShrink={0}
            flexGrow={0}
            flexBasis="460px"
            UNSAFE_style={{
              boxSizing: 'border-box',
              minWidth: '460px',
            }}
          >
            <Flex justifyContent="space-between" alignItems="center" marginBottom="size-100">
              <Heading level={4}>Preview Worker Job</Heading>
              <ActionButton onPress={() => setSelected(null)} aria-label="Close Preview">
                ×
              </ActionButton>
            </Flex>
            <Divider size="S" marginBottom="size-200" />
            <pre style={{ whiteSpace: 'pre-wrap', fontSize: '0.85rem' }}>{JSON.stringify(selected, null, 2)}</pre>
          </View>
        )}
      </Flex>
    </FullLayout>
  )
}
