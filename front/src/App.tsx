import { DatePicker, Typography } from 'antd';

const { Title } = Typography;

function App() {
  return (
    <>
      <Title level={2}>Dashboard</Title>
      <p>Welcome to ShadowAPI</p>
      <DatePicker />
    </>
  )
}

export default App
