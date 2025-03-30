import { useNavigate, useParams } from 'react-router-dom'
import { Breadcrumbs, Item } from '@adobe/react-spectrum'

import { UserForm } from '@/forms/UserForm'
import { useTitle } from '@/hooks'
import { FullLayout } from '@/layouts/FullLayout'

export function UserEdit() {
  const navigate = useNavigate()
  const { uuid } = useParams()

  const pageTitle = uuid === 'add' ? 'Add User' : 'Edit User'
  useTitle(pageTitle)
  console.log('params', { params: useParams() })

  return (
    <FullLayout>
      <Breadcrumbs marginTop="size-200" marginStart="size-300" onAction={(key) => navigate(key.toString())}>
        <Item key="/users">Users</Item>
        <Item key="/users/edit">{pageTitle}</Item>
      </Breadcrumbs>
      <UserForm userUUID={uuid!} />
    </FullLayout>
  )
}
