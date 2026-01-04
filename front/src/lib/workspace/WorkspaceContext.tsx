import { createContext, useContext, useState, useEffect, useRef, type ReactNode } from 'react';
import { useParams, useNavigate, useSearchParams } from 'react-router';
import { Spin, Result, Button } from 'antd';
import client from '../../api/client';
import { switchWorkspaceAndRedirect } from '../auth/oauth2-client';
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
  const [searchParams, setSearchParams] = useSearchParams();
  const [workspace, setWorkspace] = useState<Workspace | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  // Track if we're currently switching workspaces to prevent loops
  const isSwitchingRef = useRef(false);
  // Track if we already processed OAuth2 callback
  const processedOAuth2CallbackRef = useRef(false);

  useEffect(() => {
    if (!slug) {
      setError('No workspace specified');
      setIsLoading(false);
      return;
    }

    const checkWorkspace = async () => {
      // Don't check again if we're in the middle of switching
      if (isSwitchingRef.current) {
        return;
      }

      setIsLoading(true);
      setError(null);

      // Check if we just returned from OAuth2 callback
      const isOAuth2Callback = searchParams.get('oauth2_success') === 'true';
      if (isOAuth2Callback && !processedOAuth2CallbackRef.current) {
        processedOAuth2CallbackRef.current = true;
        // Clear the oauth2_success param from URL
        const newParams = new URLSearchParams(searchParams);
        newParams.delete('oauth2_success');
        setSearchParams(newParams, { replace: true });
        // Wait for cookies to be fully set
        await new Promise(resolve => setTimeout(resolve, 300));
      }

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

      // Fetch profile to check if JWT has correct workspace claims
      const { data: profile } = await client.GET('/profile');
      const jwtWorkspaceSlug = profile?.current_workspace?.slug;

      // If JWT doesn't have this workspace, trigger workspace switch
      // But don't switch if we just came from OAuth2 callback (would cause loop)
      if (jwtWorkspaceSlug !== slug) {
        if (processedOAuth2CallbackRef.current) {
          // We just came from OAuth2 callback but JWT still doesn't have workspace
          // This means something went wrong - show error instead of looping
          console.error('OAuth2 callback completed but JWT still missing workspace claims');
          setError('Failed to switch workspace. Please try again.');
          setIsLoading(false);
          return;
        }

        isSwitchingRef.current = true;
        try {
          await switchWorkspaceAndRedirect(slug);
          // Page will redirect, don't continue
          return;
        } catch (err) {
          isSwitchingRef.current = false;
          console.error('Failed to switch workspace:', err);
          setError('Failed to switch workspace');
          setIsLoading(false);
          return;
        }
      }

      // JWT matches URL workspace - proceed
      setWorkspace({
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
