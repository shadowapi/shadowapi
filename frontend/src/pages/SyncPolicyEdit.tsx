import { useNavigate, useParams } from 'react-router-dom'
import { Breadcrumb, Typography } from 'antd'

import { SyncPolicyForm } from '@/forms/SyncPolicyForm'
import { useTitle } from '@/hooks'
import { FullLayout } from '@/layouts/FullLayout'

const { Title } = Typography

export function SyncPolicyEdit() {
  const navigate = useNavigate()
  const { uuid } = useParams()

  const pageTitle = uuid === 'add' ? 'Add Sync Policy' : 'Edit Sync Policy'
  useTitle(pageTitle)

  return (
    <FullLayout>
      <Breadcrumb
        style={{ marginTop: 16, marginLeft: 24 }}
        items={[
          { title: 'Sync Policies', href: '', onClick: (e) => { e.preventDefault(); navigate('/syncpolicies') } },
          { title: pageTitle },
        ]}
      />
      <Title level={4} style={{ marginLeft: 24, marginTop: 8 }}>{pageTitle}</Title>
      <SyncPolicyForm policyUUID={uuid!} />
    </FullLayout>
  )
}
