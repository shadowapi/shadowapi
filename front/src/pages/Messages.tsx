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

export function Messages() {
  const [offset, setOffset] = useState(0)
  const [limit, setLimit] = useState(20)
  const [selected, setSelected] = useState<components['schemas']['message'] | null>(null)

  const messageQuery = useQuery({
    queryKey: ['messages', offset, limit],
    queryFn: async () => {
      const { data } = await client.POST('/message/query', {
        body: {
          source: 'unified',
          offset,
          limit,
        },
      })
      return (data?.messages || []) as components['schemas']['message'][]
    },
  })

  const rows = messageQuery.data || []

  return (
    <FullLayout>
      <Flex direction="row" gap="size-200" margin="size-200" flex>
        <View flex>
          <Heading level={3} marginBottom="size-100">
            Messages
          </Heading>
          <TableView aria-label="Messages Table" overflowMode="wrap" maxHeight="size-6000">
            <TableHeader>
              <Column key="sender">Sender</Column>
              <Column key="subject">Subject</Column>
              <Column key="createdAt">Created</Column>
              <Column key="actions">Actions</Column>
            </TableHeader>
            <TableBody items={rows}>
              {(item) => (
                <Row key={item.uuid}>
                  <Cell>{item.sender}</Cell>
                  <Cell>{item.subject || 'No subject'}</Cell>
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

          <Flex marginTop="size-200" gap="size-100">
            {offset > 0 && (
              <ActionButton onPress={() => setOffset(Math.max(0, offset - limit))} isDisabled={offset === 0}>
                Previous
              </ActionButton>
            )}
            <ActionButton onPress={() => setOffset(offset + limit)}>Next</ActionButton>
          </Flex>
        </View>

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
            <Flex direction="row" justifyContent="space-between" alignItems="center" marginBottom="size-100">
              <Heading level={4}>Preview Message</Heading>
              <ActionButton
                onPress={() => setSelected(null)}
                aria-label="Close Preview"
                UNSAFE_style={{ fontSize: '1.2rem' }}
              >
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
