import { useNavigate } from 'react-router-dom'
import { Breadcrumb, Typography } from 'antd'

import { ProfileForm } from '@/forms/ProfileForm'
import { useTitle } from '@/hooks'
import { FullLayout } from '@/layouts/FullLayout'

const { Title } = Typography

export function Profile() {
  const navigate = useNavigate()
  useTitle('Edit Profile')

  return (
    <FullLayout>
      <Breadcrumb
        style={{ marginTop: 16, marginLeft: 24 }}
        items={[
          { title: 'Home', href: '', onClick: (e) => { e.preventDefault(); navigate('/') } },
          { title: 'Edit Profile' },
        ]}
      />
      <Title level={4} style={{ marginLeft: 24, marginTop: 8 }}>Edit Profile</Title>
      <ProfileForm />
    </FullLayout>
  )
}
