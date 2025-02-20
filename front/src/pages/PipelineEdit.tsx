import {
  Breadcrumbs,
  Item,
} from '@adobe/react-spectrum'
import { FullLayout } from '@/layouts/FullLayout'
import { ReactFlowProvider } from '@xyflow/react'

import { useNavigate, useParams } from "react-router-dom"

import { Pipeline as PipelineForm } from "@/forms"

export function PipelineEdit() {
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
      <ReactFlowProvider>
        <PipelineForm pipelineUUID={uuid!} />
      </ReactFlowProvider>
    </FullLayout>
  )
}
