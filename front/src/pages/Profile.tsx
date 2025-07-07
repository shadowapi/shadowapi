import { Breadcrumbs, Item } from '@adobe/react-spectrum'
import { useNavigate } from 'react-router-dom'

import { ProfileForm } from '@/forms/ProfileForm'
import { useTitle } from '@/hooks'
import { FullLayout } from '@/layouts/FullLayout'

export function Profile() {
  const navigate = useNavigate()
  useTitle('Edit Profile')
  return (
    <FullLayout>
      <Breadcrumbs marginTop="size-200" marginStart="size-300" onAction={(key) => navigate(key.toString())}>
        <Item key="/">Home</Item>
        <Item key="/profile">Edit Profile</Item>
      </Breadcrumbs>
      <ProfileForm />
    </FullLayout>
  )
}
