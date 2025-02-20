import {
  Breadcrumbs,
  Item,
} from '@adobe/react-spectrum'
import { FullLayout } from '@/layouts/FullLayout'

import { useNavigate, useParams } from "react-router-dom"

import { DataSource as StorageForm } from "@/forms"

export function StorageEdit() {
  const navigate = useNavigate()
  const { uuid } = useParams()

  return (
    <FullLayout>
      <Breadcrumbs
        marginTop="size-200"
        marginStart="size-300"
        onAction={(key) => { navigate(key.toString()) }}
      >
        <Item key="/storages">Storages</Item>
        <Item key="march 2020 assets">{uuid === "add" ? "Add" : "Edit"} Storage</Item>
      </Breadcrumbs>
      <StorageForm datasourceUUID={uuid!} />
    </FullLayout>
  )
}
