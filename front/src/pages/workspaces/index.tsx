import { useState, useEffect, useCallback } from 'react';
import { useNavigate, useSearchParams } from 'react-router';
import { Typography, Card, Button, Alert, Spin, Empty, Flex } from 'antd';
import { FolderOutlined, PlusOutlined, LogoutOutlined } from '@ant-design/icons';
import client from '../../api/client';
import { uiColors } from '../../theme';
import { useAuth } from '../../lib/auth';

const { Title, Text, Paragraph } = Typography;

interface Workspace {
  uuid?: string;
  slug: string;
  display_name: string;
  is_enabled?: boolean;
}

function WorkspaceSelectionPage() {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const { logout, user } = useAuth();
  const [workspaces, setWorkspaces] = useState<Workspace[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Check if we just returned from OAuth2 callback
  const isOAuth2Callback = searchParams.get('oauth2_success') === 'true';

  const fetchWorkspaces = useCallback(async (retry = 0): Promise<boolean> => {
    try {
      const response = await client.GET('/workspace');

      if (response.error) {
        console.error('Failed to fetch workspaces:', response.error);
        // If auth error and haven't retried much, wait and retry (cookies might not be ready)
        const isAuthError = response.response?.status === 401;
        if (isAuthError && retry < 3) {
          await new Promise(resolve => setTimeout(resolve, 500));
          return fetchWorkspaces(retry + 1);
        }
        const errorDetail = typeof response.error === 'object' && 'detail' in response.error
          ? (response.error as { detail?: string }).detail
          : 'Unknown error';
        setError(`Failed to load workspaces: ${errorDetail}`);
        return false;
      } else if (response.data) {
        setWorkspaces(response.data);

        // If user has exactly one workspace, auto-redirect
        if (response.data.length === 1) {
          navigate(`/w/${response.data[0].slug}`);
        }
        return true;
      }
      return false;
    } catch (err) {
      console.error('Error fetching workspaces:', err);
      setError('Failed to load workspaces');
      return false;
    }
  }, [navigate]);

  // Fetch user's workspaces on mount
  useEffect(() => {
    const doFetch = async () => {
      // If returning from OAuth2 callback, wait a bit for cookies to settle
      if (isOAuth2Callback) {
        // Clear the oauth2_success param from URL
        searchParams.delete('oauth2_success');
        setSearchParams(searchParams, { replace: true });
        // Wait for cookies to be fully set
        await new Promise(resolve => setTimeout(resolve, 300));
      }
      await fetchWorkspaces();
      setLoading(false);
    };
    doFetch();
  }, [fetchWorkspaces, isOAuth2Callback, searchParams, setSearchParams]);

  const handleLogout = async () => {
    await logout();
    navigate('/');
  };

  const handleRetry = async () => {
    setError(null);
    setLoading(true);
    await fetchWorkspaces();
    setLoading(false);
  };

  // Navigate to a workspace
  const handleWorkspaceSelect = (slug: string) => {
    navigate(`/w/${slug}`);
  };

  if (loading) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '60vh' }}>
        <Spin size="large" />
      </div>
    );
  }

  return (
    <Card style={{ width: '100%', maxWidth: 500, position: 'relative' }}>
      <div style={{ position: 'absolute', top: 16, right: 16 }}>
        <Button
          type="text"
          icon={<LogoutOutlined />}
          onClick={handleLogout}
        >
          Logout
        </Button>
      </div>

      <div style={{ textAlign: 'center', marginBottom: 24 }}>
        <FolderOutlined style={{ fontSize: 48, color: uiColors.primary, marginBottom: 16 }} />
        <Title level={2} style={{ margin: 0 }}>Select Workspace</Title>
        <Paragraph type="secondary">
          {user?.email ? `Welcome, ${user.email}` : 'Choose a workspace to continue.'}
        </Paragraph>
      </div>

      {error && (
        <Alert
          message={error}
          type="error"
          showIcon
          closable
          onClose={() => setError(null)}
          style={{ marginBottom: 24 }}
          action={
            <Button size="small" onClick={handleRetry}>
              Retry
            </Button>
          }
        />
      )}

      {/* User's workspaces list */}
      {workspaces.length > 0 && (
        <Flex vertical gap={8}>
          {workspaces.map((workspace) => (
            <Button
              key={workspace.uuid || workspace.slug}
              type="default"
              block
              size="large"
              onClick={() => handleWorkspaceSelect(workspace.slug)}
              style={{ textAlign: 'left', height: 'auto', padding: '16px 20px' }}
              disabled={!workspace.is_enabled}
            >
              <div>
                <Text strong style={{ fontSize: 16 }}>{workspace.display_name}</Text>
                <br />
                <Text type="secondary" style={{ fontSize: 12 }}>
                  /w/{workspace.slug}
                  {!workspace.is_enabled && ' (disabled)'}
                </Text>
              </div>
            </Button>
          ))}
        </Flex>
      )}

      {workspaces.length === 0 && (
        <div style={{ textAlign: 'center', marginTop: 24 }}>
          <Empty
            description="You don't have access to any workspaces yet"
            image={Empty.PRESENTED_IMAGE_SIMPLE}
          >
            <Button type="primary" icon={<PlusOutlined />}>
              Create Workspace
            </Button>
          </Empty>
        </div>
      )}
    </Card>
  );
}

export default WorkspaceSelectionPage;
