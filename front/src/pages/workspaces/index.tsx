import { useState, useEffect, useCallback } from 'react';
import { useNavigate, useSearchParams } from 'react-router';
import { Typography, Card, Button, Spin, Empty, Flex, message } from 'antd';
import { FolderOutlined, PlusOutlined, LogoutOutlined } from '@ant-design/icons';
import client from '../../api/client';
import { uiColors } from '../../theme';
import { useAuth } from '../../lib/auth';
import { switchWorkspaceAndRedirect, OAuth2Error } from '../../lib/auth/oauth2-client';

const { Title, Text, Paragraph } = Typography;

interface Workspace {
  uuid?: string;
  slug: string;
  display_name: string;
  is_enabled?: boolean;
}

const workspaceButtonStyles = `
  .workspace-btn {
    background-color: rgba(252, 163, 17, 0.1) !important;
  }
  .workspace-btn:not(:disabled):hover {
    background-color: rgba(252, 163, 17, 0.25) !important;
  }
`;

function WorkspaceSelectionPage() {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const { logout, user } = useAuth();
  const [workspaces, setWorkspaces] = useState<Workspace[]>([]);
  const [loading, setLoading] = useState(true);
  const [switchingWorkspace, setSwitchingWorkspace] = useState<string | null>(null);

  // Check if we just returned from OAuth2 callback
  const isOAuth2Callback = searchParams.get('oauth2_success') === 'true';

  const fetchWorkspaces = useCallback(async (): Promise<boolean> => {
    try {
      const response = await client.GET('/workspace');

      if (response.error) {
        // If auth error (401), logout and let ProtectedRoute redirect to login
        if (response.response?.status === 401) {
          await logout();
          return false;
        }
        console.error('Failed to fetch workspaces:', response.error);
        return false;
      } else if (response.data) {
        setWorkspaces(response.data);

        // If user has exactly one workspace, auto-redirect using workspace switch
        if (response.data.length === 1) {
          try {
            await switchWorkspaceAndRedirect(response.data[0].slug);
          } catch {
            // If switch fails, still show the workspace list
            console.error('Auto-switch failed, showing workspace list');
          }
        }
        return true;
      }
      return false;
    } catch (err) {
      console.error('Error fetching workspaces:', err);
      return false;
    }
  }, [navigate, logout]);

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

  // Switch to a workspace (initiates OAuth2 flow to get workspace-scoped JWT)
  const handleWorkspaceSelect = async (slug: string) => {
    setSwitchingWorkspace(slug);
    try {
      // This will redirect through the OAuth2 flow
      // The user will be redirected back to /w/{slug} after getting new tokens
      await switchWorkspaceAndRedirect(slug);
      // Note: Page will redirect, so this code won't execute
    } catch (err) {
      setSwitchingWorkspace(null);
      if (err instanceof OAuth2Error) {
        if (err.status === 403) {
          message.error('You are not a member of this workspace');
        } else if (err.status === 404) {
          message.error('Workspace not found');
        } else {
          message.error(err.message || 'Failed to switch workspace');
        }
      } else {
        message.error('Failed to switch workspace');
      }
    }
  };

  if (loading) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '60vh' }}>
        <Spin size="large" />
      </div>
    );
  }

  return (
    <>
      <style>{workspaceButtonStyles}</style>
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
              className="workspace-btn"
              style={{
                textAlign: 'left',
                height: 'auto',
                padding: '16px 20px',
              }}
              disabled={!workspace.is_enabled || switchingWorkspace !== null}
              loading={switchingWorkspace === workspace.slug}
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
    </>
  );
}

export default WorkspaceSelectionPage;
