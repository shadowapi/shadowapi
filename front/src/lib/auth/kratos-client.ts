const KRATOS_BASE_URL = '/auth/kratos';

export interface KratosIdentity {
  id: string;
  traits: {
    email: string;
    name?: {
      first?: string;
      last?: string;
    };
  };
}

export interface KratosSession {
  id: string;
  active: boolean;
  expires_at: string;
  authenticated_at: string;
  identity: KratosIdentity;
}

interface KratosUiNode {
  type: string;
  group: string;
  attributes: {
    name?: string;
    type?: string;
    value?: string;
    required?: boolean;
    disabled?: boolean;
    node_type: string;
  };
  messages: Array<{ id: number; text: string; type: string }>;
}

export interface KratosLoginFlow {
  id: string;
  type: string;
  expires_at: string;
  ui: {
    action: string;
    method: string;
    nodes: KratosUiNode[];
  };
}

interface KratosLogoutFlow {
  logout_url: string;
  logout_token: string;
}

interface KratosError {
  error?: {
    id?: string;
    code?: number;
    status?: string;
    reason?: string;
    message?: string;
  };
  ui?: {
    messages?: Array<{ id: number; text: string; type: string }>;
  };
}

export class KratosAuthError extends Error {
  constructor(
    message: string,
    public code?: string,
    public details?: KratosError
  ) {
    super(message);
    this.name = 'KratosAuthError';
  }
}

export async function createLoginFlow(): Promise<KratosLoginFlow> {
  const response = await fetch(
    `${KRATOS_BASE_URL}/self-service/login/browser`,
    {
      method: 'GET',
      credentials: 'include',
      headers: {
        Accept: 'application/json',
      },
    }
  );

  if (!response.ok) {
    const error = await response.json().catch(() => ({}));
    throw new KratosAuthError(
      'Failed to create login flow',
      error.error?.id,
      error
    );
  }

  return response.json();
}

export async function submitLogin(
  flowId: string,
  email: string,
  password: string,
  csrfToken?: string
): Promise<KratosSession> {
  const body: Record<string, string> = {
    method: 'password',
    identifier: email,
    password: password,
  };

  if (csrfToken) {
    body.csrf_token = csrfToken;
  }

  const response = await fetch(
    `${KRATOS_BASE_URL}/self-service/login?flow=${flowId}`,
    {
      method: 'POST',
      credentials: 'include',
      headers: {
        'Content-Type': 'application/json',
        Accept: 'application/json',
      },
      body: JSON.stringify(body),
    }
  );

  if (!response.ok) {
    const error: KratosError = await response.json().catch(() => ({}));

    // Extract error message from Kratos response
    let message = 'Login failed';
    if (error.ui?.messages?.[0]?.text) {
      message = error.ui.messages[0].text;
    } else if (error.error?.message) {
      message = error.error.message;
    } else if (error.error?.reason) {
      message = error.error.reason;
    }

    throw new KratosAuthError(message, error.error?.id, error);
  }

  const data = await response.json();

  // Kratos returns session info on successful login
  return data.session || data;
}

export async function getSession(): Promise<KratosSession | null> {
  try {
    const response = await fetch(`${KRATOS_BASE_URL}/sessions/whoami`, {
      method: 'GET',
      credentials: 'include',
      headers: {
        Accept: 'application/json',
      },
    });

    if (response.status === 401) {
      return null;
    }

    if (!response.ok) {
      return null;
    }

    return response.json();
  } catch {
    return null;
  }
}

export async function createLogoutFlow(): Promise<KratosLogoutFlow> {
  const response = await fetch(
    `${KRATOS_BASE_URL}/self-service/logout/browser`,
    {
      method: 'GET',
      credentials: 'include',
      headers: {
        Accept: 'application/json',
      },
    }
  );

  if (!response.ok) {
    const error = await response.json().catch(() => ({}));
    throw new KratosAuthError(
      'Failed to create logout flow',
      error.error?.id,
      error
    );
  }

  return response.json();
}

export async function executeLogout(): Promise<void> {
  const logoutFlow = await createLogoutFlow();

  // Execute logout by visiting the logout URL
  const response = await fetch(logoutFlow.logout_url, {
    method: 'GET',
    credentials: 'include',
  });

  if (!response.ok && response.status !== 303) {
    throw new KratosAuthError('Logout failed');
  }
}
