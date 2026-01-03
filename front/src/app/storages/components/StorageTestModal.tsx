import { Modal, Result, Space, Button, Typography, Alert, Progress } from 'antd';
import {
  CheckCircleOutlined,
  CloseCircleOutlined,
  LoadingOutlined,
  InfoCircleOutlined,
} from '@ant-design/icons';
import type { StorageTestState } from '../hooks/useStorageConnectionTest';

const { Text, Paragraph } = Typography;
const TOTAL_SECONDS = 5;

interface StorageTestModalProps {
  open: boolean;
  state: StorageTestState;
  storageName: string;
  onProceed: () => void;
  onCancel: () => void;
  onRetry: () => void;
}

function StorageTestModal({
  open,
  state,
  storageName,
  onProceed,
  onCancel,
  onRetry,
}: StorageTestModalProps) {
  const { status, result } = state;

  const isSuccess = result?.success || result?.skipped;
  const isFailed = result && !result.success && !result.skipped;

  const getTitle = () => {
    if (status === 'testing') return 'Testing Connection...';
    if (isSuccess) return 'Connection Verified';
    if (isFailed) return 'Connection Failed';
    return 'Test Results';
  };

  const renderResult = () => {
    if (!result) return null;

    if (result.skipped) {
      return (
        <Space>
          <InfoCircleOutlined style={{ color: '#1890ff' }} />
          <Text type="secondary">{result.skipReason}</Text>
        </Space>
      );
    }

    if (result.success) {
      return (
        <Space>
          <CheckCircleOutlined style={{ color: '#52c41a' }} />
          <Text>Connected successfully</Text>
          {result.durationMs !== undefined && (
            <Text type="secondary">({result.durationMs}ms)</Text>
          )}
        </Space>
      );
    }

    return (
      <div>
        <Space>
          <CloseCircleOutlined style={{ color: '#ff4d4f' }} />
          <Text type="danger">Connection failed</Text>
        </Space>
        {result.errorMessage && (
          <Paragraph type="secondary" style={{ marginTop: 8 }}>
            {result.errorMessage}
          </Paragraph>
        )}
        {result.errorDetails && (
          <Alert type="error" message={result.errorDetails} style={{ marginTop: 8 }} />
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

    if (isSuccess) {
      return [
        <Button key="cancel" onClick={onCancel}>
          Cancel
        </Button>,
        <Button key="proceed" type="primary" onClick={onProceed}>
          Save Storage
        </Button>,
      ];
    }

    // Failed
    return [
      <Button key="cancel" onClick={onCancel}>
        Cancel
      </Button>,
      <Button key="retry" onClick={onRetry}>
        Retry Test
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
          <Space>
            <LoadingOutlined spin style={{ color: '#1890ff' }} />
            <Text>Testing PostgreSQL connection to {storageName || 'database'}...</Text>
          </Space>
          <Progress
            percent={
              ((TOTAL_SECONDS - (state.remainingSeconds ?? TOTAL_SECONDS)) / TOTAL_SECONDS) * 100
            }
            status="active"
            showInfo={false}
            strokeColor="#1890ff"
          />
          <Text type="secondary" style={{ display: 'block', textAlign: 'center' }}>
            Timeout in {state.remainingSeconds ?? TOTAL_SECONDS}s
          </Text>
        </Space>
      )}

      {status === 'completed' && (
        <Result
          status={isSuccess ? 'success' : 'error'}
          title={isSuccess ? 'Connection test passed!' : 'Connection test failed'}
          subTitle={
            isFailed
              ? 'You can still save the storage, but the connection may not work properly.'
              : undefined
          }
        >
          {renderResult()}
        </Result>
      )}
    </Modal>
  );
}

export default StorageTestModal;
