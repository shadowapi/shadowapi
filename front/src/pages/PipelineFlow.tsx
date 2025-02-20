import {
  Breadcrumbs,
  Item,
} from '@adobe/react-spectrum'
import { FullLayout } from '@/layouts/FullLayout'

import { useNavigate, useParams } from "react-router-dom"

export function PipelineFlow() {
  const navigate = useNavigate()
  const { uuid } = useParams()

  return (
    <FullLayout>
      <Breadcrumbs
        marginTop="size-200"
        marginStart="size-300"
        onAction={(key) => { navigate(key.toString()) }}
      >
        <Item key="/pipelines">Pipelines</Item>
        <Item key="march 2020 assets">{uuid === "add" ? "Add" : "Edit"} Pipeline</Item>
      </Breadcrumbs>
    </FullLayout>
  )
}
