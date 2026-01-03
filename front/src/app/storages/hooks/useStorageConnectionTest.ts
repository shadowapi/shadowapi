import { useState, useRef, useCallback } from 'react';
import client from '../../../api/client';
import type { components } from '../../../api/v1';

const POLL_INTERVAL = 1000; // 1 second
const MAX_POLL_ATTEMPTS = 5; // 5 second timeout

type TestConnectionJob = components['schemas']['test_connection_job'];
type TestConnectionResult = components['schemas']['test_connection_result'];

export interface PostgresTestParams {
  is_same_database?: boolean;
  user?: string;
  password?: string;
  host?: string;
  port?: string;
  database?: string;
  options?: string;
}

export interface StorageTestResult {
  success: boolean;
  errorCode?: string;
  errorMessage?: string;
  errorDetails?: string;
  durationMs?: number;
  skipped?: boolean;
  skipReason?: string;
}

export interface StorageTestState {
  status: 'idle' | 'testing' | 'completed';
  result?: StorageTestResult;
  error?: string;
  remainingSeconds?: number;
}

export function useStorageConnectionTest() {
  const [state, setState] = useState<StorageTestState>({ status: 'idle' });
  const abortRef = useRef(false);

  const pollJob = async (jobUuid: string): Promise<TestConnectionResult | null> => {
    let attempts = 0;

    while (attempts < MAX_POLL_ATTEMPTS && !abortRef.current) {
      const remaining = MAX_POLL_ATTEMPTS - attempts;
      setState((prev) => ({ ...prev, remainingSeconds: remaining }));

      const { data, error } = await client.GET('/test-connection-job/{uuid}', {
        params: { path: { uuid: jobUuid } },
      });

      if (error) {
        return null;
      }

      const job = data as TestConnectionJob;
      if (job.status === 'completed' || job.status === 'failed') {
        return {
          success: job.result?.success ?? false,
          error_code: job.result?.error_code,
          error_message: job.result?.error_message,
          error_details: job.result?.error_details,
          duration_ms: job.result?.duration_ms,
          tested_at: job.result?.tested_at,
        };
      }

      await new Promise((resolve) => setTimeout(resolve, POLL_INTERVAL));
      attempts++;
    }

    // Timeout
    return {
      success: false,
      error_code: 'connection_timeout',
      error_message: 'Test timed out after 5 seconds',
    };
  };

  const testPostgresConnection = useCallback(async (params: PostgresTestParams) => {
    abortRef.current = false;
    setState({ status: 'testing', remainingSeconds: MAX_POLL_ATTEMPTS });

    // is_same_database = true means we don't need an external test
    if (params.is_same_database) {
      setState({
        status: 'completed',
        result: {
          success: true,
          skipped: true,
          skipReason: 'Using application database - no external test needed',
        },
      });
      return;
    }

    // Call inline test endpoint
    const { data, error, response } = await client.POST('/storage/postgres/test', {
      body: {
        is_same_database: params.is_same_database ?? false,
        user: params.user,
        password: params.password,
        host: params.host,
        port: params.port,
        database: params.database,
        options: params.options,
      },
    });

    if (error) {
      setState({
        status: 'completed',
        result: {
          success: false,
          errorMessage: (error as { detail?: string }).detail || 'Failed to start test',
        },
      });
      return;
    }

    // Check for immediate success (200)
    if (response.status === 200 && data && 'success' in data) {
      const result = data as TestConnectionResult;
      setState({
        status: 'completed',
        result: {
          success: result.success,
          durationMs: result.duration_ms ?? 0,
        },
      });
      return;
    }

    // 202 - need to poll
    const job = data as TestConnectionJob;
    if (!job?.uuid) {
      setState({
        status: 'completed',
        result: {
          success: false,
          errorMessage: 'No job UUID returned',
        },
      });
      return;
    }

    const result = await pollJob(job.uuid);
    if (abortRef.current) {
      setState({ status: 'idle' });
      return;
    }

    if (!result) {
      setState({
        status: 'completed',
        result: {
          success: false,
          errorMessage: 'Failed to poll job status',
        },
      });
      return;
    }

    setState({
      status: 'completed',
      result: {
        success: result.success,
        errorCode: result.error_code,
        errorMessage: result.error_message,
        errorDetails: result.error_details,
        durationMs: result.duration_ms,
      },
    });
  }, []);

  const cancel = useCallback(() => {
    abortRef.current = true;
    setState({ status: 'idle' });
  }, []);

  const reset = useCallback(() => {
    abortRef.current = false;
    setState({ status: 'idle' });
  }, []);

  return { state, testPostgresConnection, cancel, reset };
}
