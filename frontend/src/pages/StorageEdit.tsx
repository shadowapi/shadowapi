import { useNavigate, useParams } from 'react-router-dom'
import { Breadcrumb, Typography } from 'antd'

import { StorageForm } from '@/forms'
import { StorageKind } from '@/forms/StorageForm'
import { useTitle } from '@/hooks'
import { FullLayout } from '@/layouts/FullLayout'

const { Title } = Typography

export function StorageEdit() {
  const navigate = useNavigate()
  const { uuid, storageKind } = useParams()

  const pageTitle = uuid === 'add' ? 'Add Storage' : 'Edit Storage'
  useTitle(pageTitle)

  return (
    <FullLayout>
      <Breadcrumb
        style={{ marginTop: 16, marginLeft: 24 }}
        items={[
          { title: 'Storages', href: '', onClick: (e) => { e.preventDefault(); navigate('/storages') } },
          { title: pageTitle },
        ]}
      />
      <Title level={4} style={{ marginLeft: 24, marginTop: 8 }}>{pageTitle}</Title>
      <StorageForm storageUUID={uuid!} storageKind={storageKind as StorageKind} />
    </FullLayout>
  )
}
