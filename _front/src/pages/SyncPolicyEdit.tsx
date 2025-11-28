import { useNavigate, useParams } from 'react-router-dom'
import { Breadcrumbs, Item } from '@adobe/react-spectrum'

import { SyncPolicyForm } from '@/forms/SyncPolicyForm'
import { useTitle } from '@/hooks'
import { FullLayout } from '@/layouts/FullLayout'

export function SyncPolicyEdit() {
  const navigate = useNavigate()
  const { uuid } = useParams()
  const pageTitle = uuid === 'add' ? 'Add Sync Policy' : 'Edit Sync Policy'
  useTitle(pageTitle)

  return (
    <FullLayout>
      <Breadcrumbs marginTop="size-200" marginStart="size-300" onAction={(key) => navigate(key.toString())}>
        <Item key="/syncpolicies">Sync Policies</Item>
        <Item key="syncpolicy">{uuid === 'add' ? 'Add Sync Policy' : 'Edit Sync Policy'}</Item>
      </Breadcrumbs>
      <SyncPolicyForm policyUUID={uuid!} />
    </FullLayout>
  )
}
