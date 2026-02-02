import React, { useEffect, useState } from 'react';
import { Card, Form, Input, Button, Typography, Tabs, Row, Col, Descriptions, Tag, Space, Spin, Progress, Modal, message, Statistic, InputNumber, Popconfirm, Alert } from 'antd';
import { ReloadOutlined, InfoCircleOutlined, LockOutlined, CloudDownloadOutlined, DatabaseOutlined, DeleteOutlined, ClearOutlined, SettingOutlined } from '@ant-design/icons';
import { useRequest } from 'ahooks';
import { useTranslation } from 'react-i18next';
import { 
  updatePassword, getSystemInfo, checkUpdate, performUpdate, 
  getDatabaseStats, cleanDatabase, vacuumDatabase, getSettings, saveSettings,
  type SystemInfo, type UpdateCheckResult, type DatabaseStats, type SystemSettings 
} from '../api';

const Settings: React.FC = () => {
  const { t } = useTranslation();
  const [form] = Form.useForm();
  const [settingsForm] = Form.useForm();
  
  // System info & update state
  const [systemInfo, setSystemInfo] = useState<SystemInfo | null>(null);
  const [updateInfo, setUpdateInfo] = useState<UpdateCheckResult | null>(null);
  const [checkingUpdate, setCheckingUpdate] = useState(false);
  const [updating, setUpdating] = useState(false);
  const [updateProgress, setUpdateProgress] = useState(0);

  // Database state
  const [dbStats, setDbStats] = useState<DatabaseStats | null>(null);
  const [loadingStats, setLoadingStats] = useState(false);
  const [cleaning, setCleaning] = useState(false);
  const [vacuuming, setVacuuming] = useState(false);

  // Settings state
  const [settings, setSettings] = useState<SystemSettings | null>(null);
  const [savingSettings, setSavingSettings] = useState(false);

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
    fetchDatabaseStats();
    fetchSettings();
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

  const fetchDatabaseStats = async () => {
    setLoadingStats(true);
    try {
      const stats = await getDatabaseStats();
      setDbStats(stats);
    } catch (e) {
      console.error('Failed to fetch database stats:', e);
    } finally {
      setLoadingStats(false);
    }
  };

  const fetchSettings = async () => {
    try {
      const s = await getSettings();
      setSettings(s);
      settingsForm.setFieldsValue(s);
    } catch (e) {
      console.error('Failed to fetch settings:', e);
    }
  };

  const handleCheckUpdate = async () => {
    setCheckingUpdate(true);
    try {
      const result = await checkUpdate();
      setUpdateInfo(result);
      if (result.has_update) {
        message.info(t('settings.newVersionAvailable', { version: result.latest_version }));
      } else {
        message.success(t('settings.onLatestVersion'));
      }
    } catch (e) {
      message.error(t('settings.updateFailed'));
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
        message.success(t('settings.updateSuccessRestart'));
        setTimeout(() => window.location.reload(), 3000);
      } else {
        message.info(result.message || t('settings.onLatestVersion'));
        setUpdating(false);
        setUpdateProgress(0);
      }
    } catch (e: any) {
      clearInterval(progressInterval);
      message.error(e?.message || t('settings.updateFailed'));
      setUpdating(false);
      setUpdateProgress(0);
    }
  };

  const handleCleanDatabase = async () => {
    setCleaning(true);
    try {
      const result = await cleanDatabase(dbStats?.retention_days || 30);
      message.success(t('settings.dbCleanSuccess', { count: result.deleted }) || `Deleted ${result.deleted} old records`);
      fetchDatabaseStats();
    } catch (e) {
      message.error(t('settings.dbCleanFailed') || 'Failed to clean database');
    } finally {
      setCleaning(false);
    }
  };

  const handleVacuum = async () => {
    setVacuuming(true);
    try {
      await vacuumDatabase();
      message.success(t('settings.dbVacuumSuccess') || 'Database optimized successfully');
      fetchDatabaseStats();
    } catch (e) {
      message.error(t('settings.dbVacuumFailed') || 'Failed to optimize database');
    } finally {
      setVacuuming(false);
    }
  };

  const handleSaveSettings = async () => {
    setSavingSettings(true);
    try {
      const values = await settingsForm.validateFields();
      const result = await saveSettings(values);
      setSettings(result);
      message.success(t('settings.settingsSaved') || 'Settings saved');
    } catch (e) {
      message.error(t('settings.settingsSaveFailed') || 'Failed to save settings');
    } finally {
      setSavingSettings(false);
    }
  };

  const tabItems = [
    {
      key: '1',
      label: <span><DatabaseOutlined style={{ marginRight: 6 }} />{t('settings.tabs.database') || 'Database'}</span>,
      children: (
        <Row gutter={[24, 24]}>
          <Col xs={24} lg={12}>
            <Card title={t('settings.databaseStats') || 'Database Statistics'} loading={loadingStats}>
              {dbStats && (
                <Row gutter={[16, 16]}>
                  <Col span={12}>
                    <Statistic 
                      title={t('settings.dbSize') || 'Database Size'} 
                      value={dbStats.size_human} 
                    />
                  </Col>
                  <Col span={12}>
                    <Statistic 
                      title={t('settings.recordCount') || 'Records'} 
                      value={dbStats.record_count} 
                    />
                  </Col>
                  <Col span={12}>
                    <Statistic 
                      title={t('settings.targetCount') || 'Targets'} 
                      value={dbStats.target_count} 
                    />
                  </Col>
                  <Col span={12}>
                    <Statistic 
                      title={t('settings.retentionDays') || 'Retention'} 
                      value={dbStats.retention_days} 
                      suffix={t('common.days') || 'days'}
                    />
                  </Col>
                  {dbStats.oldest_record && (
                    <Col span={24}>
                      <Typography.Text type="secondary">
                        {t('settings.oldestRecord') || 'Oldest'}: {new Date(dbStats.oldest_record).toLocaleString()}
                      </Typography.Text>
                    </Col>
                  )}
                </Row>
              )}
            </Card>
          </Col>
          <Col xs={24} lg={12}>
            <Card title={t('settings.databaseActions') || 'Database Actions'}>
              <Space direction="vertical" style={{ width: '100%' }} size="middle">
                <Alert 
                  message={t('settings.dbCleanWarning') || 'Cleaning will delete old monitoring records'} 
                  type="warning" 
                  showIcon 
                />
                <Space>
                  <Popconfirm
                    title={t('settings.confirmClean') || 'Clean old records?'}
                    description={t('settings.confirmCleanDesc', { days: dbStats?.retention_days || 30 }) || `Delete records older than ${dbStats?.retention_days || 30} days`}
                    onConfirm={handleCleanDatabase}
                    okText={t('common.yes') || 'Yes'}
                    cancelText={t('common.no') || 'No'}
                  >
                    <Button icon={<DeleteOutlined />} loading={cleaning} danger>
                      {t('settings.cleanOldRecords') || 'Clean Old Records'}
                    </Button>
                  </Popconfirm>
                  <Popconfirm
                    title={t('settings.confirmVacuum') || 'Optimize database?'}
                    description={t('settings.confirmVacuumDesc') || 'This may take a moment for large databases'}
                    onConfirm={handleVacuum}
                    okText={t('common.yes') || 'Yes'}
                    cancelText={t('common.no') || 'No'}
                  >
                    <Button icon={<ClearOutlined />} loading={vacuuming}>
                      {t('settings.optimizeDatabase') || 'Optimize (Vacuum)'}
                    </Button>
                  </Popconfirm>
                </Space>
                <Button icon={<ReloadOutlined />} onClick={fetchDatabaseStats} loading={loadingStats}>
                  {t('common.refresh') || 'Refresh'}
                </Button>
              </Space>
            </Card>
          </Col>
        </Row>
      ),
    },
    {
      key: '2',
      label: <span><SettingOutlined style={{ marginRight: 6 }} />{t('settings.tabs.monitoring') || 'Monitoring'}</span>,
      children: (
        <Card title={t('settings.monitoringSettings') || 'Monitoring Settings'} style={{ maxWidth: 600 }}>
          <Form
            form={settingsForm}
            layout="vertical"
            initialValues={settings || { retention_days: 30, speed_test_interval_minutes: 5, ping_interval_seconds: 30 }}
          >
            <Form.Item
              name="retention_days"
              label={t('settings.retentionDays') || 'Data Retention (days)'}
              rules={[{ required: true }]}
              tooltip={t('settings.retentionDaysTooltip') || 'How long to keep monitoring data'}
            >
              <InputNumber min={1} max={365} style={{ width: '100%' }} />
            </Form.Item>
            <Form.Item
              name="speed_test_interval_minutes"
              label={t('settings.speedTestInterval') || 'Speed Test Interval (minutes)'}
              rules={[{ required: true }]}
              tooltip={t('settings.speedTestIntervalTooltip') || 'How often to run speed tests'}
            >
              <InputNumber min={1} max={60} style={{ width: '100%' }} />
            </Form.Item>
            <Form.Item
              name="ping_interval_seconds"
              label={t('settings.pingInterval') || 'Ping Interval (seconds)'}
              rules={[{ required: true }]}
              tooltip={t('settings.pingIntervalTooltip') || 'How often to run ping/trace'}
            >
              <InputNumber min={10} max={300} style={{ width: '100%' }} />
            </Form.Item>
            <Button type="primary" onClick={handleSaveSettings} loading={savingSettings}>
              {t('common.save') || 'Save'}
            </Button>
          </Form>
        </Card>
      ),
    },
    {
      key: '3',
      label: <span><LockOutlined style={{ marginRight: 6 }} />{t('settings.tabs.security')}</span>,
      children: (
        <Card title={t('settings.changePassword')} style={{ maxWidth: 500 }}>
          <Form layout="vertical" form={form} onFinish={run}>
            <Form.Item name="newPassword" label={t('settings.newPassword')} rules={[{ required: true, min: 6 }]}>
              <Input.Password placeholder={t('settings.newPassword')} />
            </Form.Item>
            <Form.Item
              name="confirm"
              label={t('settings.confirmPassword')}
              dependencies={['newPassword']}
              rules={[
                { required: true },
                ({ getFieldValue }) => ({
                  validator: (_, value) =>
                    value && value !== getFieldValue('newPassword')
                      ? Promise.reject(new Error(t('settings.passwordMismatch') || 'Passwords do not match'))
                      : Promise.resolve(),
                }),
              ]}
            >
              <Input.Password placeholder={t('settings.confirmPassword')} />
            </Form.Item>
            <Button type="primary" htmlType="submit" loading={loading}>
              {t('settings.changePassword')}
            </Button>
          </Form>
        </Card>
      ),
    },
    {
      key: '4',
      label: <span><InfoCircleOutlined style={{ marginRight: 6 }} />{t('settings.tabs.about')}</span>,
      children: (
        <Row gutter={24}>
          <Col xs={24} lg={12}>
            <Card title={t('settings.systemInfo')}>
              {systemInfo ? (
                <Descriptions column={1} size="small">
                  <Descriptions.Item label={t('common.version') || 'Version'}>
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
            <Card title={t('settings.softwareUpdate')}>
              <Space direction="vertical" style={{ width: '100%' }}>
                {updateInfo && (
                  <Descriptions column={1} size="small">
                    <Descriptions.Item label={t('settings.current')}>
                      {updateInfo.current_version}
                    </Descriptions.Item>
                    <Descriptions.Item label={t('settings.latest')}>
                      {updateInfo.latest_version || t('common.na')}
                    </Descriptions.Item>
                    <Descriptions.Item label={t('settings.status')}>
                      {updateInfo.has_update 
                        ? <Tag color="orange">{t('settings.updateAvailable')}</Tag> 
                        : <Tag color="green">{t('settings.upToDate')}</Tag>
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
                    {t('settings.checkForUpdates')}
                  </Button>
                  
                  {updateInfo?.has_update && (
                    <Button 
                      type="primary" 
                      icon={<CloudDownloadOutlined />}
                      onClick={handlePerformUpdate}
                      loading={updating}
                    >
                      {t('settings.installUpdate')}
                    </Button>
                  )}
                </Space>
                
                {updateInfo?.release_notes && (
                  <Card title={t('settings.releaseNotes')} size="small" style={{ marginTop: 16 }}>
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
      <Card className="page-card" title={t('settings.title')}>
        <Tabs items={tabItems} type="card" />
      </Card>

      {/* Update Progress Modal */}
      <Modal
        title={t('settings.updating')}
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
              ? t('settings.updateComplete')
              : t('settings.downloading')}
          </Typography.Text>
        </div>
      </Modal>
    </>
  );
};

export default Settings;

