import {
  Breadcrumbs,
  Item,
} from '@adobe/react-spectrum'
import { FullLayout } from '@/layouts/FullLayout'

import { useNavigate, useParams } from "react-router-dom"

import { OAuth2Credential as OAuth2CredentialForm } from "@/forms"

export function OAuth2CredentialEdit() {
  const navigate = useNavigate()
  const { clientID } = useParams()

  return (
    <FullLayout>
      <Breadcrumbs
        marginTop="size-200"
        marginStart="size-300"
        onAction={(key) => { navigate(key.toString()) }}
      >
        <Item key="/oauth2/credentials">OAuth2 Credentials</Item>
        <Item key="march 2020 assets">{clientID === "add" ? "Add" : "Edit"} OAuth2 Credential</Item>
      </Breadcrumbs>
      <OAuth2CredentialForm clientID={clientID!} />
    </FullLayout>
  )
}
