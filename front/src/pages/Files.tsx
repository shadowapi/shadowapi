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

export function Files() {
  // For pagination
  const [offset, setOffset] = useState(0)
  const [limit, setLimit] = useState(20)

  // For preview panel: store currently selected file to preview
  const [selected, setSelected] = useState<components['schemas']['FileObject'] | null>(null)

  // Query file objects (GET /file) using the offset and limit parameters.
  const fileQuery = useQuery({
    queryKey: ['files', offset, limit],
    queryFn: async () => {
      const { data } = await client.GET('/file', {
        params: {
          query: { offset, limit },
        },
      })
      // We assume the response is an array of file objects.
      return data as components['schemas']['FileObject'][]
    },
  })

  const rows = fileQuery.data || []

  console.log({ rows })

  return (
    <FullLayout>
      <Flex direction="row" gap="size-200" margin="size-200" flex>
        {/* LEFT SIDE: Files Table */}
        <View flex>
          <Heading level={3} marginBottom="size-100">
            Files
          </Heading>
          <TableView aria-label="Files Table" overflowMode="wrap" maxHeight="size-6000">
            <TableHeader>
              <Column key="name">Name</Column>
              <Column key="mime_type">MIME Type</Column>
              <Column key="size">Size (bytes)</Column>
              <Column key="createdAt">Created</Column>
              <Column key="actions" align="end">
                Actions
              </Column>
            </TableHeader>
            <TableBody items={rows}>
              {(item) => (
                <Row key={item.uuid}>
                  <Cell>{item.name}</Cell>
                  <Cell>{item.mime_type || 'N/A'}</Cell>
                  <Cell>{item.size || 0}</Cell>
                  <Cell>{item.created_at ? item.created_at.slice(0, 19) : ''}</Cell>
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
            width="size-4600"
            borderStartWidth="thin"
            borderColor="dark"
            overflow="auto"
            UNSAFE_style={{ boxSizing: 'border-box' }}
          >
            <Flex justifyContent="space-between" alignItems="center" marginBottom="size-100">
              <Heading level={4}>Preview File</Heading>
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
