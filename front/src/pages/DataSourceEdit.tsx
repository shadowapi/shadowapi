import {
  Breadcrumbs,
  Item,
} from '@adobe/react-spectrum'
import { FullLayout } from '@/layouts/FullLayout'

import { useNavigate, useParams } from "react-router-dom"

import { DataSource as DataSourceForm } from "@/forms"

export function DataSourceEdit() {
  const navigate = useNavigate()
  const { uuid } = useParams()

  return (
    <FullLayout>
      <Breadcrumbs
        marginTop="size-200"
        marginStart="size-300"
        onAction={(key) => { navigate(key.toString()) }}
      >
        <Item key="/datasources">Data Sources</Item>
        <Item key="march 2020 assets">{uuid === "add" ? "Add" : "Edit"} Data Source</Item>
      </Breadcrumbs>
      <DataSourceForm datasourceUUID={uuid!} />
    </FullLayout>
  )
}
