import { useState, useEffect } from 'react';
import { Dropdown, Button, Space, Spin, Typography } from 'antd';
import type { MenuProps } from 'antd';
import { SwapOutlined, AppstoreOutlined, DownOutlined } from '@ant-design/icons';
import { useNavigate, useLocation } from 'react-router';
import client from '../api/client';
import { switchWorkspaceAndRedirect } from '../lib/auth/oauth2-client';
import type { components } from '../api/v1';

type Workspace = components['schemas']['workspace'];

interface WorkspaceSwitcherProps {
  currentWorkspaceSlug?: string;
}

export default function WorkspaceSwitcher({ currentWorkspaceSlug }: WorkspaceSwitcherProps) {
  const navigate = useNavigate();
  const location = useLocation();
  const [workspaces, setWorkspaces] = useState<Workspace[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isSwitching, setIsSwitching] = useState(false);

  // Extract current workspace slug from URL if not provided
  const getSlugFromPath = (): string | undefined => {
    const match = location.pathname.match(/^\/w\/([^/]+)/);
    return match ? match[1] : undefined;
  };

  const activeSlug = currentWorkspaceSlug || getSlugFromPath();

  useEffect(() => {
    const fetchWorkspaces = async () => {
      setIsLoading(true);
      const { data, error } = await client.GET('/workspace');
      if (!error && data) {
        setWorkspaces(data);
      }
      setIsLoading(false);
    };

    fetchWorkspaces();
  }, []);

  const handleSwitch = async (slug: string) => {
    if (slug === activeSlug || isSwitching) return;
    setIsSwitching(true);
    try {
      await switchWorkspaceAndRedirect(slug);
    } catch (err) {
      console.error('Failed to switch workspace:', err);
      setIsSwitching(false);
    }
  };

  const currentWorkspace = workspaces.find(w => w.slug === activeSlug);
  const displayName = currentWorkspace?.display_name || activeSlug || 'Select Workspace';

  const menuItems: MenuProps['items'] = [
    ...(workspaces.length > 0
      ? workspaces.map(w => ({
          key: w.slug || w.uuid || '',
          label: (
            <Space>
              {w.display_name || w.slug}
              {w.slug === activeSlug && (
                <Typography.Text type="secondary" style={{ fontSize: 12 }}>
                  (current)
                </Typography.Text>
              )}
            </Space>
          ),
          disabled: !w.is_enabled || w.slug === activeSlug,
          onClick: () => w.slug && handleSwitch(w.slug),
        }))
      : [
          {
            key: 'no-workspaces',
            label: <Typography.Text type="secondary">No workspaces available</Typography.Text>,
            disabled: true,
          },
        ]),
    { type: 'divider' as const },
    {
      key: 'all-workspaces',
      icon: <AppstoreOutlined />,
      label: 'All Workspaces',
      onClick: () => navigate('/workspaces'),
    },
  ];

  if (isLoading) {
    return <Spin size="small" style={{ marginRight: 16 }} />;
  }

  return (
    <Dropdown
      menu={{ items: menuItems }}
      trigger={['click']}
      disabled={isSwitching}
    >
      <Button
        type="text"
        style={{ color: '#fff', marginRight: 8 }}
        loading={isSwitching}
      >
        <Space>
          <SwapOutlined />
          {displayName}
          <DownOutlined />
        </Space>
      </Button>
    </Dropdown>
  );
}
