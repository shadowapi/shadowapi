import { useNavigate } from 'react-router-dom'
import { Button, Empty, Flex } from 'antd'
import { PlusOutlined } from '@ant-design/icons'

import { FullLayout } from '@/layouts/FullLayout'

export function Dashboard() {
  const navigate = useNavigate()
  return (
    <FullLayout>
      <Flex vertical justify="center" align="center" style={{ flex: 1, padding: 40 }} gap={16}>
        <Empty description="No data sources are set up. Add a new data source to start processing data." />
        <Button type="primary" icon={<PlusOutlined />} onClick={() => navigate('/datasources/add')}>
          Add Data Source
        </Button>
      </Flex>
    </FullLayout>
  )
}
