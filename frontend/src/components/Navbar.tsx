import { useLocation, useNavigate } from 'react-router-dom'
import { Menu } from 'antd'
import type { MenuProps } from 'antd'

export interface NavbarProps {
  Label: string
  AriaLabel: string
  Icon: React.JSX.Element
  URL: string
  Childrens?: NavbarProps[]
}

function toMenuItems(elements: NavbarProps[]): MenuProps['items'] {
  return elements.map((item) => ({
    key: item.URL,
    icon: item.Icon,
    label: item.Label,
    children: item.Childrens?.length ? toMenuItems(item.Childrens) : undefined,
  }))
}

export function Navbar(props: { elements: NavbarProps[] }) {
  const location = useLocation()
  const navigate = useNavigate()

  const onClick: MenuProps['onClick'] = ({ key }) => {
    navigate(key)
  }

  return (
    <Menu
      mode="inline"
      selectedKeys={[location.pathname]}
      items={toMenuItems(props.elements)}
      onClick={onClick}
      style={{ width: 256 }}
    />
  )
}
