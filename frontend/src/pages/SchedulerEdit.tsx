import { useNavigate, useParams } from 'react-router-dom'
import { Breadcrumb, Typography } from 'antd'

import { SchedulerForm } from '@/forms/SchedulerForm'
import { useTitle } from '@/hooks'
import { FullLayout } from '@/layouts/FullLayout'

const { Title } = Typography

export function SchedulerEdit() {
  const navigate = useNavigate()
  const { uuid } = useParams()

  const pageTitle = uuid === 'add' ? 'Add Scheduler' : 'Edit Scheduler'
  useTitle(pageTitle)

  return (
    <FullLayout>
      <Breadcrumb
        style={{ marginTop: 16, marginLeft: 24 }}
        items={[
          { title: 'Schedulers', href: '', onClick: (e) => { e.preventDefault(); navigate('/schedulers') } },
          { title: pageTitle },
        ]}
      />
      <Title level={4} style={{ marginLeft: 24, marginTop: 8 }}>{pageTitle}</Title>
      <SchedulerForm schedulerUUID={uuid!} />
    </FullLayout>
  )
}
