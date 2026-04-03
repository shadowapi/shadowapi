import { useNavigate, useParams } from 'react-router-dom'
import { Breadcrumb, Typography } from 'antd'

import { DataSourceForm } from '@/forms'
import { useTitle } from '@/hooks'
import { FullLayout } from '@/layouts/FullLayout'

const { Title } = Typography

export function DataSourceEdit() {
  const navigate = useNavigate()
  const { uuid } = useParams()

  const pageTitle = uuid === 'add' ? 'Add Data Source' : 'Edit Data Source'
  useTitle(pageTitle)

  return (
    <FullLayout>
      <Breadcrumb
        style={{ marginTop: 16, marginLeft: 24 }}
        items={[
          { title: 'Data Sources', href: '', onClick: (e) => { e.preventDefault(); navigate('/datasources') } },
          { title: pageTitle },
        ]}
      />
      <Title level={4} style={{ marginLeft: 24, marginTop: 8 }}>{pageTitle}</Title>
      <DataSourceForm datasourceUUID={uuid!} />
    </FullLayout>
  )
}
