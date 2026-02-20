import { createContext, useContext, useState, useEffect, useRef, type ReactNode } from 'react';
import { useParams, useNavigate } from 'react-router';
import { Spin, Result, Button } from 'antd';
import client from '../../api/client';
import { switchWorkspace } from '../auth/oauth2-client';

interface Workspace {
  uuid?: string;
  slug: string;
  display_name?: string;
}

interface WorkspaceContextType {
  workspace: Workspace | null;
  slug: string;
  isLoading: boolean;
  error: string | null;
}

const WorkspaceContext = createContext<WorkspaceContextType | undefined>(undefined);

interface WorkspaceProviderProps {
  children: ReactNode;
}

export function WorkspaceProvider({ children }: WorkspaceProviderProps) {
  const { slug } = useParams<{ slug: string }>();
  const navigate = useNavigate();
  const [workspace, setWorkspace] = useState<Workspace | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const isSwitchingRef = useRef(false);

  useEffect(() => {
    if (!slug) {
      setError('No workspace specified');
      setIsLoading(false);
      return;
    }

    const checkWorkspace = async () => {
      if (isSwitchingRef.current) {
        return;
      }

      setIsLoading(true);
      setError(null);

      // Check if workspace exists
      const { data: checkData, error: checkError } = await client.GET('/workspace/check', {
        params: { query: { slug } },
      });

      if (checkError || !checkData?.exists) {
        setError('Workspace not found');
        setWorkspace(null);
        setIsLoading(false);
        return;
      }

      // Fetch profile to check if workspace cookie matches
      const { data: profile } = await client.GET('/profile');
      const currentWorkspaceSlug = profile?.current_workspace?.slug;

      // If current workspace doesn't match URL, switch via cookie
      if (currentWorkspaceSlug !== slug) {
        isSwitchingRef.current = true;
        try {
          const result = await switchWorkspace(slug);
          // Cookie is now set - re-fetch profile to confirm
          await client.GET('/profile');
          setWorkspace({
            uuid: result.workspace_uuid,
            slug,
            display_name: checkData.display_name || slug,
          });
          isSwitchingRef.current = false;
          setIsLoading(false);
          return;
        } catch (err) {
          isSwitchingRef.current = false;
          console.error('Failed to switch workspace:', err);
          setError('Failed to switch workspace');
          setIsLoading(false);
          return;
        }
      }

      // Workspace cookie matches URL - proceed
      setWorkspace({
        uuid: profile?.current_workspace?.uuid,
        slug,
        display_name: checkData.display_name || slug,
      });
      setIsLoading(false);
    };

    checkWorkspace();
  }, [slug]); // eslint-disable-line react-hooks/exhaustive-deps

  if (isLoading) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: 200 }}>
        <Spin size="large" />
      </div>
    );
  }

  if (error) {
    return (
      <Result
        status="404"
        title="Workspace Not Found"
        subTitle={`The workspace "${slug}" does not exist or you don't have access.`}
        extra={
          <Button type="primary" onClick={() => navigate('/workspaces')}>
            Back to Workspaces
          </Button>
        }
      />
    );
  }

  const value: WorkspaceContextType = {
    workspace,
    slug: slug || '',
    isLoading,
    error,
  };

  return <WorkspaceContext.Provider value={value}>{children}</WorkspaceContext.Provider>;
}

export function useWorkspace(): WorkspaceContextType {
  const context = useContext(WorkspaceContext);
  if (context === undefined) {
    throw new Error('useWorkspace must be used within a WorkspaceProvider');
  }
  return context;
}

export { WorkspaceContext };
