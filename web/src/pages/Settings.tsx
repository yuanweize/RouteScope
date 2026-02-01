import React, { useEffect, useState } from 'react';
import { Card, Form, Input, Button, Typography, Tabs, Row, Col, Descriptions, Tag, Space, Spin, Progress, Modal, message } from 'antd';
import { ReloadOutlined, InfoCircleOutlined, LockOutlined, CloudDownloadOutlined } from '@ant-design/icons';
import { useRequest } from 'ahooks';
import { updatePassword, getSystemInfo, checkUpdate, performUpdate, type SystemInfo, type UpdateCheckResult } from '../api';

const Settings: React.FC = () => {
  const [form] = Form.useForm();
  
  // System info & update state
  const [systemInfo, setSystemInfo] = useState<SystemInfo | null>(null);
  const [updateInfo, setUpdateInfo] = useState<UpdateCheckResult | null>(null);
  const [checkingUpdate, setCheckingUpdate] = useState(false);
  const [updating, setUpdating] = useState(false);
  const [updateProgress, setUpdateProgress] = useState(0);

  const { run, loading } = useRequest(
    async (values: { newPassword: string }) => updatePassword(values.newPassword),
    {
      manual: true,
      onSuccess: () => {
        form.resetFields();
        message.success('Password updated successfully');
      },
    }
  );

  useEffect(() => {
    fetchSystemInfo();
    // Auto-check for updates on mount (only once)
    handleCheckUpdateSilent();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // Silent update check (no messages) - used only on initial mount
  const handleCheckUpdateSilent = async () => {
    try {
      const result = await checkUpdate();
      setUpdateInfo(result);
    } catch (e) {
      console.error('Failed to check for updates:', e);
    }
  };

  const fetchSystemInfo = async () => {
    try {
      const info = await getSystemInfo();
      setSystemInfo(info);
    } catch (e) {
      console.error('Failed to fetch system info:', e);
    }
  };

  const handleCheckUpdate = async () => {
    setCheckingUpdate(true);
    try {
      const result = await checkUpdate();
      setUpdateInfo(result);
      if (result.has_update) {
        message.info(`New version ${result.latest_version} available!`);
      } else {
        message.success('You are on the latest version.');
      }
    } catch (e) {
      message.error('Failed to check for updates');
    } finally {
      setCheckingUpdate(false);
    }
  };

  const handlePerformUpdate = async () => {
    setUpdating(true);
    setUpdateProgress(0);
    
    const progressInterval = setInterval(() => {
      setUpdateProgress(prev => {
        if (prev >= 90) return prev;
        return prev + Math.random() * 15;
      });
    }, 500);

    try {
      const result = await performUpdate();
      clearInterval(progressInterval);
      setUpdateProgress(100);
      
      if (result.updated) {
        message.success('Update successful! Restarting service...');
        setTimeout(() => window.location.reload(), 3000);
      } else {
        message.info(result.message || 'No update available');
        setUpdating(false);
        setUpdateProgress(0);
      }
    } catch (e: any) {
      clearInterval(progressInterval);
      message.error(e?.message || 'Update failed');
      setUpdating(false);
      setUpdateProgress(0);
    }
  };

  const tabItems = [
    {
      key: '1',
      label: <span><LockOutlined style={{ marginRight: 6 }} />Security</span>,
      children: (
        <Card title="Change Admin Password" style={{ maxWidth: 500 }}>
          <Form layout="vertical" form={form} onFinish={run}>
            <Form.Item name="newPassword" label="New Password" rules={[{ required: true, min: 6 }]}>
              <Input.Password placeholder="Enter new password" />
            </Form.Item>
            <Form.Item
              name="confirm"
              label="Confirm Password"
              dependencies={['newPassword']}
              rules={[
                { required: true },
                ({ getFieldValue }) => ({
                  validator: (_, value) =>
                    value && value !== getFieldValue('newPassword')
                      ? Promise.reject(new Error('Passwords do not match'))
                      : Promise.resolve(),
                }),
              ]}
            >
              <Input.Password placeholder="Confirm password" />
            </Form.Item>
            <Button type="primary" htmlType="submit" loading={loading}>
              Update Password
            </Button>
          </Form>
        </Card>
      ),
    },
    {
      key: '2',
      label: <span><InfoCircleOutlined style={{ marginRight: 6 }} />About & Updates</span>,
      children: (
        <Row gutter={24}>
          <Col xs={24} lg={12}>
            <Card title="System Information">
              {systemInfo ? (
                <Descriptions column={1} size="small">
                  <Descriptions.Item label="Version">
                    <Tag color="blue">{systemInfo.version}</Tag>
                  </Descriptions.Item>
                  <Descriptions.Item label="Commit">
                    <Typography.Text code>{systemInfo.commit?.slice(0, 8)}</Typography.Text>
                  </Descriptions.Item>
                  <Descriptions.Item label="Build Date">
                    {systemInfo.build_date}
                  </Descriptions.Item>
                  <Descriptions.Item label="Go Version">
                    {systemInfo.go_version}
                  </Descriptions.Item>
                  <Descriptions.Item label="OS / Arch">
                    {systemInfo.os} / {systemInfo.arch}
                  </Descriptions.Item>
                </Descriptions>
              ) : (
                <Spin />
              )}
            </Card>
          </Col>
          <Col xs={24} lg={12}>
            <Card title="Software Update">
              <Space direction="vertical" style={{ width: '100%' }}>
                {updateInfo && (
                  <Descriptions column={1} size="small">
                    <Descriptions.Item label="Current">
                      {updateInfo.current_version}
                    </Descriptions.Item>
                    <Descriptions.Item label="Latest">
                      {updateInfo.latest_version || 'Unknown'}
                    </Descriptions.Item>
                    <Descriptions.Item label="Status">
                      {updateInfo.has_update 
                        ? <Tag color="orange">Update Available</Tag> 
                        : <Tag color="green">Up to Date</Tag>
                      }
                    </Descriptions.Item>
                  </Descriptions>
                )}
                
                <Space style={{ marginTop: 16 }}>
                  <Button 
                    icon={<ReloadOutlined />}
                    onClick={handleCheckUpdate}
                    loading={checkingUpdate}
                  >
                    Check for Updates
                  </Button>
                  
                  {updateInfo?.has_update && (
                    <Button 
                      type="primary" 
                      icon={<CloudDownloadOutlined />}
                      onClick={handlePerformUpdate}
                      loading={updating}
                    >
                      Install Update
                    </Button>
                  )}
                </Space>
                
                {updateInfo?.release_notes && (
                  <Card title="Release Notes" size="small" style={{ marginTop: 16 }}>
                    <Typography.Paragraph 
                      style={{ 
                        maxHeight: 200, 
                        overflow: 'auto',
                        whiteSpace: 'pre-wrap',
                        fontSize: 12,
                        margin: 0
                      }}
                    >
                      {updateInfo.release_notes}
                    </Typography.Paragraph>
                  </Card>
                )}
              </Space>
            </Card>
          </Col>
        </Row>
      ),
    },
  ];

  return (
    <>
      <Card className="page-card" title="Settings">
        <Tabs items={tabItems} type="card" />
      </Card>

      {/* Update Progress Modal */}
      <Modal
        title="Updating RouteLens"
        open={updating}
        footer={null}
        closable={false}
        maskClosable={false}
      >
        <div style={{ textAlign: 'center', padding: '20px 0' }}>
          <Progress 
            percent={Math.round(updateProgress)} 
            status={updateProgress >= 100 ? 'success' : 'active'}
          />
          <Typography.Text style={{ marginTop: 16, display: 'block' }}>
            {updateProgress >= 100 
              ? 'Update complete! Restarting service...' 
              : 'Downloading and installing update...'}
          </Typography.Text>
        </div>
      </Modal>
    </>
  );
};

export default Settings;

