import { useNavigate, useParams } from 'react-router-dom'
import { Breadcrumbs, Item } from '@adobe/react-spectrum'

import { DataSourceForm } from '@/forms'
import { useTitle } from '@/hooks'
import { FullLayout } from '@/layouts/FullLayout'

export function DataSourceEdit() {
  const navigate = useNavigate()
  const { uuid } = useParams()

  const pageTitle = uuid === 'add' ? 'Add Data Source' : 'Edit Data Source'
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
        <Item key="/datasources">Data Sources</Item>
        <Item key="march 2020 assets">{uuid === 'add' ? 'Add' : 'Edit'} Data Source</Item>
      </Breadcrumbs>
      <DataSourceForm datasourceUUID={uuid!} />
    </FullLayout>
  )
}
