import { useState, useRef, useCallback } from 'react';
import client from '../../../api/client';
import type { components } from '../../../api/v1';

const POLL_INTERVAL = 1000; // 1 second
const MAX_POLL_ATTEMPTS = 5; // 5 second timeout

type TestConnectionJob = components['schemas']['test_connection_job'];
type TestConnectionResult = components['schemas']['test_connection_result'];

export interface TestTarget {
  type: 'datasource' | 'storage';
  uuid: string;
  resourceType: string; // e.g., 'email_oauth', 'postgres'
  name: string;
}

export interface TestResult {
  target: TestTarget;
  success: boolean;
  errorCode?: string;
  errorMessage?: string;
  errorDetails?: string;
  durationMs?: number;
  skipped?: boolean;
  skipReason?: string;
}

export interface TestState {
  status: 'idle' | 'testing' | 'completed';
  datasourceResult?: TestResult;
  storageResult?: TestResult;
  error?: string;
  remainingSeconds?: number;
}

interface DatasourceInput {
  uuid: string;
  type: string;
  name: string;
}

interface StorageInput {
  uuid: string;
  type: string;
  name: string;
  isSameDatabase?: boolean;
}

export function useConnectionTest() {
  const [state, setState] = useState<TestState>({ status: 'idle' });
  const abortRef = useRef(false);

  const pollJob = async (jobUuid: string): Promise<TestConnectionResult | null> => {
    let attempts = 0;

    while (attempts < MAX_POLL_ATTEMPTS && !abortRef.current) {
      // Update remaining seconds
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

  const testDatasource = async (datasource: DatasourceInput): Promise<TestResult> => {
    const target: TestTarget = {
      type: 'datasource',
      uuid: datasource.uuid,
      resourceType: datasource.type,
      name: datasource.name,
    };

    // Only email_oauth has test endpoint
    if (datasource.type !== 'email_oauth') {
      return {
        target,
        success: true,
        skipped: true,
        skipReason: `Test not available for ${datasource.type} datasources`,
      };
    }

    const { data, error } = await client.POST('/datasource/email_oauth/{uuid}/test', {
      params: { path: { uuid: datasource.uuid } },
    });

    if (error) {
      return {
        target,
        success: false,
        errorMessage: (error as { detail?: string }).detail || 'Failed to start test',
      };
    }

    const job = data as TestConnectionJob;
    if (!job?.uuid) {
      return {
        target,
        success: false,
        errorMessage: 'No job UUID returned',
      };
    }

    // Update state to show datasource test started
    setState((prev) => ({
      ...prev,
      datasourceResult: { target, success: false }, // In progress indicator
    }));

    const result = await pollJob(job.uuid);
    if (!result) {
      return {
        target,
        success: false,
        errorMessage: 'Failed to poll job status',
      };
    }

    return {
      target,
      success: result.success,
      errorCode: result.error_code,
      errorMessage: result.error_message,
      errorDetails: result.error_details,
      durationMs: result.duration_ms,
    };
  };

  const testStorage = async (storage: StorageInput): Promise<TestResult> => {
    const target: TestTarget = {
      type: 'storage',
      uuid: storage.uuid,
      resourceType: storage.type,
      name: storage.name,
    };

    // Only postgres has test endpoint
    if (storage.type !== 'postgres') {
      return {
        target,
        success: true,
        skipped: true,
        skipReason: `Test not available for ${storage.type} storages`,
      };
    }

    const { data, error, response } = await client.POST('/storage/postgres/{uuid}/test', {
      params: { path: { uuid: storage.uuid } },
    });

    if (error) {
      return {
        target,
        success: false,
        errorMessage: (error as { detail?: string }).detail || 'Failed to start test',
      };
    }

    // Check if immediate success (200 for is_same_database=true)
    // The response will be TestConnectionResult directly
    if (response.status === 200 && data && 'success' in data) {
      const result = data as TestConnectionResult;
      return {
        target,
        success: result.success,
        durationMs: result.duration_ms ?? 0,
      };
    }

    // 202 - need to poll
    const job = data as TestConnectionJob;
    if (!job?.uuid) {
      return {
        target,
        success: false,
        errorMessage: 'No job UUID returned',
      };
    }

    // Update state to show storage test started
    setState((prev) => ({
      ...prev,
      storageResult: { target, success: false }, // In progress indicator
    }));

    const result = await pollJob(job.uuid);
    if (!result) {
      return {
        target,
        success: false,
        errorMessage: 'Failed to poll job status',
      };
    }

    return {
      target,
      success: result.success,
      errorCode: result.error_code,
      errorMessage: result.error_message,
      errorDetails: result.error_details,
      durationMs: result.duration_ms,
    };
  };

  const startTest = useCallback(
    async (datasource: DatasourceInput, storage: StorageInput) => {
      abortRef.current = false;
      setState({ status: 'testing', remainingSeconds: MAX_POLL_ATTEMPTS });

      // Run tests concurrently
      const [dsResult, stResult] = await Promise.all([
        testDatasource(datasource),
        testStorage(storage),
      ]);

      if (abortRef.current) {
        setState({ status: 'idle' });
        return;
      }

      setState({
        status: 'completed',
        datasourceResult: dsResult,
        storageResult: stResult,
      });
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps -- testDatasource/testStorage use only stable refs
    []
  );

  const cancelTest = useCallback(() => {
    abortRef.current = true;
  }, []);

  const reset = useCallback(() => {
    abortRef.current = false;
    setState({ status: 'idle' });
  }, []);

  return { state, startTest, cancelTest, reset };
}
