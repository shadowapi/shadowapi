import { useNavigate, useParams } from 'react-router-dom'
import { Breadcrumb, Typography } from 'antd'

import { OAuth2CredentialForm } from '@/forms'
import { FullLayout } from '@/layouts/FullLayout'

const { Title } = Typography

export function OAuth2CredentialEdit() {
  const navigate = useNavigate()
  const { clientID } = useParams()

  const pageTitle = clientID === 'add' ? 'Add OAuth2 Credential' : 'Edit OAuth2 Credential'

  return (
    <FullLayout>
      <Breadcrumb
        style={{ marginTop: 16, marginLeft: 24 }}
        items={[
          { title: 'OAuth2 Credentials', href: '', onClick: (e) => { e.preventDefault(); navigate('/oauth2/credentials') } },
          { title: pageTitle },
        ]}
      />
      <Title level={4} style={{ marginLeft: 24, marginTop: 8 }}>{pageTitle}</Title>
      <OAuth2CredentialForm clientID={clientID!} />
    </FullLayout>
  )
}
