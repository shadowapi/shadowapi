import { useNavigate, useParams } from 'react-router-dom'
import { Breadcrumb, Typography } from 'antd'

import { FullLayout } from '@/layouts/FullLayout'

const { Title } = Typography

export function PipelineFlow() {
  const navigate = useNavigate()
  const { uuid } = useParams()

  const pageTitle = uuid === 'add' ? 'Add Pipeline' : 'Edit Pipeline'

  return (
    <FullLayout>
      <Breadcrumb
        style={{ marginTop: 16, marginLeft: 24 }}
        items={[
          { title: 'Pipelines', href: '', onClick: (e) => { e.preventDefault(); navigate('/pipelines') } },
          { title: `${pageTitle} Flow` },
        ]}
      />
      <Title level={4} style={{ marginLeft: 24, marginTop: 8 }}>{pageTitle} Flow</Title>
    </FullLayout>
  )
}
