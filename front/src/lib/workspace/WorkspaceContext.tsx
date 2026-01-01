import { createContext, useContext, useState, useEffect, type ReactNode } from 'react';
import { useParams, useNavigate } from 'react-router';
import { Spin, Result, Button } from 'antd';
import client from '../../api/client';
import type { components } from '../../api/v1';

type Workspace = components['schemas']['workspace'];

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

  useEffect(() => {
    if (!slug) {
      setError('No workspace specified');
      setIsLoading(false);
      return;
    }

    const checkWorkspace = async () => {
      setIsLoading(true);
      setError(null);

      const { data, error: apiError } = await client.GET('/workspace/check', {
        params: { query: { slug } },
      });

      if (apiError || !data?.exists) {
        setError('Workspace not found');
        setWorkspace(null);
        setIsLoading(false);
        return;
      }

      // Set workspace with available info from check endpoint
      setWorkspace({
        slug,
        display_name: data.display_name || slug,
      });
      setIsLoading(false);
    };

    checkWorkspace();
  }, [slug]);

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
