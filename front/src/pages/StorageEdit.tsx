import { useNavigate, useParams } from 'react-router-dom'
import { Breadcrumbs, Item } from '@adobe/react-spectrum'

import { StorageForm } from '@/forms'
import { useTitle } from '@/hooks'
import { FullLayout } from '@/layouts/FullLayout'

export function StorageEdit() {
  const navigate = useNavigate()
  const { uuid } = useParams()

  const pageTitle = uuid === 'add' ? 'Add Storage!' : 'Edit Storage'
  useTitle(pageTitle)

  return (
    <FullLayout>
      <Breadcrumbs
        marginTop="size-200"
        marginStart="size-300"
        onAction={(key) => {
          navigate(key.toString())
        }}
      >
        <Item key="/storages">Storages</Item>
        <Item key="march 2020 assets">{uuid === 'add' ? 'Add' : 'Edit'} Storage</Item>
      </Breadcrumbs>
      <StorageForm storageUUID={uuid!} />
    </FullLayout>
  )
}
