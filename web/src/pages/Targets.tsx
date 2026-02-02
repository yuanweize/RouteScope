import React, { useMemo, useState } from 'react';
import { Button, Card, Form, Input, Modal, Select, Space, Switch, Table, Tag, Upload, message } from 'antd';
import { PlusOutlined, UploadOutlined } from '@ant-design/icons';
import { useRequest } from 'ahooks';
import { useTranslation } from 'react-i18next';
import type { Target } from '../api';
import { deleteTarget, getTargets, saveTarget } from '../api';

const probeOptions = [
  { label: 'ICMP', value: 'MODE_ICMP' },
  { label: 'HTTP', value: 'MODE_HTTP' },
  { label: 'SSH', value: 'MODE_SSH' },
  { label: 'IPERF', value: 'MODE_IPERF' },
];

const Targets: React.FC = () => {
  const { t } = useTranslation();
  const [form] = Form.useForm();
  const [open, setOpen] = useState(false);
  const [editing, setEditing] = useState<Target | null>(null);

  const { data = [], refresh, loading } = useRequest(getTargets);

  // Toggle target enabled/disabled status
  const handleToggleEnabled = async (record: Target, checked: boolean) => {
    const payload: Target = {
      ...record,
      enabled: checked,
    };
    try {
      await saveTarget(payload);
      refresh();
      message.success(checked ? t('common.enabled') : t('common.disabled'));
    } catch (e) {
      message.error(t('common.error'));
    }
  };

  const columns = useMemo(() => [
    { title: t('targets.name'), dataIndex: 'name' },
    { title: t('targets.hostIp'), dataIndex: 'address' },
    { title: t('targets.probeType'), dataIndex: 'probe_type', render: (val: string) => <Tag color="blue">{val}</Tag> },
    { 
      title: t('common.status'), 
      dataIndex: 'enabled', 
      render: (val: boolean, record: Target) => (
        <Switch 
          checked={val} 
          onChange={(checked) => handleToggleEnabled(record, checked)}
          checkedChildren={t('common.enabled')}
          unCheckedChildren={t('common.disabled')}
        />
      )
    },
    {
      title: t('common.actions'),
      render: (_: any, record: Target) => (
        <Space>
          <Button type="link" onClick={() => onEdit(record)}>{t('common.edit')}</Button>
          <Button type="link" danger onClick={() => onDelete(record.id)}>{t('common.delete')}</Button>
        </Space>
      ),
    },
  // eslint-disable-next-line react-hooks/exhaustive-deps
  ], [t]);

  const onEdit = (record: Target) => {
    setEditing(record);
    
    // Parse probe_config JSON to populate individual fields
    let parsedConfig: Record<string, unknown> = {};
    if (record.probe_config) {
      try {
        parsedConfig = JSON.parse(record.probe_config);
      } catch {
        // ignore parse error
      }
    }
    
    form.setFieldsValue({
      name: record.name,
      address: record.address,
      desc: record.desc,
      enabled: record.enabled,
      probe_type: record.probe_type,
      // HTTP fields
      http_url: parsedConfig.url || '',
      // SSH fields
      ssh_user: parsedConfig.user || 'root',
      ssh_port: parsedConfig.port || 22,
      ssh_key_path: parsedConfig.key_path || '',
      ssh_key_text: parsedConfig.key_text || '',
      // iPerf fields
      iperf_port: parsedConfig.port || 5201,
    });
    setOpen(true);
  };

  const onDelete = async (id?: number) => {
    if (!id) return;
    await deleteTarget(id);
    refresh();
  };

  const onCreate = () => {
    setEditing(null);
    form.resetFields();
    form.setFieldsValue({ enabled: true, probe_type: 'MODE_ICMP' });
    setOpen(true);
  };

  const handleUpload = (file: File) => {
    const reader = new FileReader();
    reader.onload = () => {
      form.setFieldValue('ssh_key_text', reader.result as string);
      message.success('SSH key loaded');
    };
    reader.readAsText(file);
    return false;
  };

  const buildProbeConfig = (values: any) => {
    switch (values.probe_type) {
      case 'MODE_HTTP':
        return JSON.stringify({ url: values.http_url || '' });
      case 'MODE_SSH':
        return JSON.stringify({
          user: values.ssh_user || 'root',
          key_path: values.ssh_key_path || '',
          key_text: values.ssh_key_text || '',
          port: Number(values.ssh_port || 22),
        });
      case 'MODE_IPERF':
        return JSON.stringify({ port: Number(values.iperf_port || 5201) });
      default:
        return '';
    }
  };

  const onSubmit = async () => {
    const values = await form.validateFields();
    const payload: Target = {
      id: editing?.id,
      name: values.name,
      address: values.address,
      desc: values.desc || '',
      enabled: values.enabled ?? true,
      probe_type: values.probe_type,
      probe_config: buildProbeConfig(values),
    };
    await saveTarget(payload);
    setOpen(false);
    refresh();
  };

  return (
    <Card className="page-card" title={t('targets.title')} extra={<Button icon={<PlusOutlined />} onClick={onCreate}>{t('targets.newTarget')}</Button>}>
      <Table rowKey="id" loading={loading} dataSource={data} columns={columns} />

      <Modal
        title={editing ? t('targets.editTarget') : t('targets.newTarget')}
        open={open}
        onOk={onSubmit}
        onCancel={() => setOpen(false)}
        destroyOnClose
      >
        <Form layout="vertical" form={form} preserve={false}>
          <Form.Item name="name" label={t('targets.name')} rules={[{ required: true }]}>
            <Input placeholder="Hong Kong VPS" />
          </Form.Item>
          <Form.Item name="address" label={t('targets.hostIp')} rules={[{ required: true }]}>
            <Input placeholder="1.2.3.4" />
          </Form.Item>
          <Form.Item name="probe_type" label={t('targets.probeType')} rules={[{ required: true }]}>
            <Select options={probeOptions} />
          </Form.Item>
          <Form.Item shouldUpdate={(prev, cur) => prev.probe_type !== cur.probe_type}>
            {({ getFieldValue }) => {
              const mode = getFieldValue('probe_type');
              if (mode === 'MODE_HTTP') {
                return (
                  <Form.Item name="http_url" label={t('targets.httpUrl')} rules={[{ required: true }]}>
                    <Input placeholder="https://example.com/test.zip" />
                  </Form.Item>
                );
              }
              if (mode === 'MODE_SSH') {
                return (
                  <>
                    <Form.Item name="ssh_user" label={t('targets.sshUser')} rules={[{ required: true }]}>
                      <Input placeholder="root" />
                    </Form.Item>
                    <Form.Item name="ssh_port" label={t('targets.sshPort')}>
                      <Input placeholder="22" />
                    </Form.Item>
                    <Form.Item name="ssh_key_path" label={t('targets.sshKeyPath')}>
                      <Input placeholder="/root/.ssh/id_rsa" />
                    </Form.Item>
                    <Form.Item name="ssh_key_text" label={t('targets.sshKeyText')}>
                      <Input.TextArea rows={4} placeholder="Paste private key content" />
                    </Form.Item>
                    <Upload beforeUpload={handleUpload} showUploadList={false}>
                      <Button icon={<UploadOutlined />}>{t('targets.uploadKey')}</Button>
                    </Upload>
                  </>
                );
              }
              if (mode === 'MODE_IPERF') {
                return (
                  <Form.Item name="iperf_port" label={t('targets.iperfPort')} rules={[{ required: true }]}>
                    <Input placeholder="5201" />
                  </Form.Item>
                );
              }
              return null;
            }}
          </Form.Item>
          <Form.Item name="desc" label={t('targets.description')}>
            <Input.TextArea rows={3} />
          </Form.Item>
        </Form>
      </Modal>
    </Card>
  );
};

export default Targets;
