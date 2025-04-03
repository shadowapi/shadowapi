import { useNavigate, useParams } from 'react-router-dom'
import { Breadcrumbs, Item } from '@adobe/react-spectrum'

import { SchedulerForm } from '@/forms/SchedulerForm'
import { useTitle } from '@/hooks'
import { FullLayout } from '@/layouts/FullLayout'

export function SchedulerEdit() {
  const navigate = useNavigate()
  const { id } = useParams()
  const pageTitle = id === 'add' ? 'Add Scheduler' : 'Edit Scheduler'
  useTitle(pageTitle)
  console.log('params', { params: useParams() })

  return (
    <FullLayout>
      <Breadcrumbs marginTop="size-200" marginStart="size-300" onAction={(key) => navigate(key.toString())}>
        <Item key="/schedulers">Schedulers</Item>
        <Item key="/schedulers/edit">{pageTitle}</Item>
      </Breadcrumbs>
      <SchedulerForm schedulerUUID={id!} />
    </FullLayout>
  )
}
