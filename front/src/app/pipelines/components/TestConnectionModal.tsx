import { Modal, Result, Space, Button, Typography, Alert, Progress } from 'antd';
import {
  CheckCircleOutlined,
  CloseCircleOutlined,
  LoadingOutlined,
  InfoCircleOutlined,
} from '@ant-design/icons';
import type { TestState, TestResult } from './useConnectionTest';

const { Text, Paragraph } = Typography;
const TOTAL_SECONDS = 5; // Must match MAX_POLL_ATTEMPTS in useConnectionTest

interface TestConnectionModalProps {
  open: boolean;
  state: TestState;
  onProceed: () => void;
  onCancel: () => void;
  onRetry: () => void;
}

function TestConnectionModal({
  open,
  state,
  onProceed,
  onCancel,
  onRetry,
}: TestConnectionModalProps) {
  const { status, datasourceResult, storageResult } = state;

  const isTestComplete = (result?: TestResult) =>
    result && (result.success || result.skipped || result.errorCode || result.errorMessage);

  const allSuccess =
    status === 'completed' &&
    (datasourceResult?.success || datasourceResult?.skipped) &&
    (storageResult?.success || storageResult?.skipped);

  const anyFailed =
    status === 'completed' &&
    ((datasourceResult && !datasourceResult.success && !datasourceResult.skipped) ||
      (storageResult && !storageResult.success && !storageResult.skipped));

  const getTitle = () => {
    if (status === 'testing') return 'Testing Connections...';
    if (allSuccess) return 'Connections Verified';
    if (anyFailed) return 'Connection Issue';
    return 'Test Results';
  };

  const getResultStatus = (): 'success' | 'error' | 'warning' => {
    if (allSuccess) return 'success';
    if (anyFailed) return 'error';
    return 'warning';
  };

  const renderTestItem = (label: string, result?: TestResult, showLoading?: boolean) => {
    // Show loading spinner during testing phase when result not yet complete
    if (showLoading && !isTestComplete(result)) {
      return (
        <Space>
          <LoadingOutlined spin style={{ color: '#1890ff' }} />
          <Text>{label}</Text>
        </Space>
      );
    }

    if (!result) return null;

    if (result.skipped) {
      return (
        <Space>
          <InfoCircleOutlined style={{ color: '#1890ff' }} />
          <Text type="secondary">
            {label}: {result.skipReason}
          </Text>
        </Space>
      );
    }

    if (result.success) {
      return (
        <Space>
          <CheckCircleOutlined style={{ color: '#52c41a' }} />
          <Text>{label}: Connected successfully</Text>
          {result.durationMs !== undefined && <Text type="secondary">({result.durationMs}ms)</Text>}
        </Space>
      );
    }

    return (
      <div>
        <Space>
          <CloseCircleOutlined style={{ color: '#ff4d4f' }} />
          <Text type="danger">{label}: Failed</Text>
        </Space>
        {result.errorMessage && (
          <Paragraph type="secondary" style={{ marginLeft: 22, marginTop: 4, marginBottom: 0 }}>
            {result.errorMessage}
          </Paragraph>
        )}
        {result.errorDetails && (
          <Alert
            type="error"
            message={result.errorDetails}
            style={{ marginLeft: 22, marginTop: 8 }}
          />
        )}
      </div>
    );
  };

  const footer = () => {
    if (status === 'testing') {
      return [
        <Button key="cancel" onClick={onCancel}>
          Cancel
        </Button>,
      ];
    }

    if (allSuccess) {
      return [
        <Button key="cancel" onClick={onCancel}>
          Cancel
        </Button>,
        <Button key="proceed" type="primary" onClick={onProceed}>
          Save Pipeline
        </Button>,
      ];
    }

    // Some tests failed
    return [
      <Button key="cancel" onClick={onCancel}>
        Cancel
      </Button>,
      <Button key="retry" onClick={onRetry}>
        Retry Tests
      </Button>,
      <Button key="proceed" type="primary" danger onClick={onProceed}>
        Save Anyway
      </Button>,
    ];
  };

  return (
    <Modal
      title={getTitle()}
      open={open}
      closable={status !== 'testing'}
      maskClosable={false}
      footer={footer()}
      onCancel={onCancel}
    >
      {status === 'testing' && (
        <Space direction="vertical" style={{ width: '100%' }} size="middle">
          {renderTestItem(
            `Data Source: ${datasourceResult?.target?.name || 'Loading...'}`,
            datasourceResult,
            true
          )}
          {renderTestItem(
            `Storage: ${storageResult?.target?.name || 'Loading...'}`,
            storageResult,
            true
          )}
          <div style={{ marginTop: 16 }}>
            <Progress
              percent={((TOTAL_SECONDS - (state.remainingSeconds ?? TOTAL_SECONDS)) / TOTAL_SECONDS) * 100}
              status="active"
              showInfo={false}
              strokeColor="#1890ff"
            />
            <Text type="secondary" style={{ display: 'block', textAlign: 'center', marginTop: 4 }}>
              Timeout in {state.remainingSeconds ?? TOTAL_SECONDS}s
            </Text>
          </div>
        </Space>
      )}

      {status === 'completed' && (
        <Result
          status={getResultStatus()}
          title={allSuccess ? 'All tests passed!' : 'Some tests failed'}
          subTitle={
            anyFailed
              ? 'You can still save the pipeline, but the connection may not work properly.'
              : undefined
          }
        >
          <Space direction="vertical" style={{ width: '100%' }} size="middle">
            {renderTestItem(`Data Source: ${datasourceResult?.target?.name}`, datasourceResult)}
            {renderTestItem(`Storage: ${storageResult?.target?.name}`, storageResult)}
          </Space>
        </Result>
      )}
    </Modal>
  );
}

export default TestConnectionModal;
