import { useNavigate, useParams } from 'react-router-dom'
import { Breadcrumb, Typography } from 'antd'
import { ReactFlowProvider } from '@xyflow/react'

import { PipelineForm } from '@/forms'
import { useTitle } from '@/hooks'
import { FullLayout } from '@/layouts/FullLayout'

const { Title } = Typography

export function PipelineEdit() {
  const navigate = useNavigate()
  const { uuid } = useParams()

  const pageTitle = uuid === 'add' ? 'Add Pipeline' : 'Edit Pipeline'
  useTitle(pageTitle)

  return (
    <FullLayout>
      <Breadcrumb
        style={{ marginTop: 16, marginLeft: 24 }}
        items={[
          { title: 'Pipelines', href: '', onClick: (e) => { e.preventDefault(); navigate('/pipelines') } },
          { title: pageTitle },
        ]}
      />
      <Title level={4} style={{ marginLeft: 24, marginTop: 8 }}>{pageTitle}</Title>
      <ReactFlowProvider>
        <PipelineForm pipelineUUID={uuid!} userUUID="" />
      </ReactFlowProvider>
    </FullLayout>
  )
}
