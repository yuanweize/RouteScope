import React, { useEffect, useState } from 'react';
import { Table, Button, Card, Modal, Form, Input, Switch, Message, Space, Typography, Popconfirm, Tabs, Select, Tag, Grid } from '@arco-design/web-react';
import { IconPlus, IconDelete, IconEdit, IconSafe, IconUnorderedList } from '@arco-design/web-react/icon';
import { getTargets, saveTarget, deleteTarget, updatePassword } from '../../api';
import type { Target } from '../../api';

const { TabPane } = Tabs;
const { Row, Col } = Grid;

const Settings: React.FC = () => {
    const [targets, setTargets] = useState<Target[]>([]);
    const [loading, setLoading] = useState(false);
    const [isModalVisible, setIsModalVisible] = useState(false);
    const [form] = Form.useForm();
    const [passForm] = Form.useForm();
    const [editingId, setEditingId] = useState<number | null>(null);

    const fetchTargets = async () => {
        setLoading(true);
        try {
            const data = await getTargets();
            setTargets(data as any);
        } catch (e) {
            Message.error('Failed to fetch targets');
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchTargets();
    }, []);

    const handleAdd = () => {
        setEditingId(null);
        form.resetFields();
        form.setFieldValue('ProbeMode', 'ICMP');
        setIsModalVisible(true);
    };

    const handleEdit = (record: Target) => {
        setEditingId(record.ID || null);
        form.setFieldsValue(record);
        setIsModalVisible(true);
    };

    const handleDelete = async (id: number) => {
        try {
            await deleteTarget(id);
            Message.success('Target deleted');
            fetchTargets();
        } catch (e) {
            Message.error('Delete failed');
        }
    };

    const handleOk = async () => {
        try {
            const values = await form.validate();
            await saveTarget({ ...values, ID: editingId || undefined });
            Message.success(editingId ? 'Target updated' : 'Target added');
            setIsModalVisible(false);
            fetchTargets();
        } catch (e) {
            // Validation or API error handled by Arco/Intercept
        }
    };

    const handlePassUpdate = async (values: any) => {
        try {
            await updatePassword(values.newPassword);
            Message.success('Password updated successfully');
            passForm.resetFields();
        } catch (e) { console.error(e); }
    };

    const columns = [
        { title: 'Name', dataIndex: 'Name', key: 'Name' },
        { title: 'Address', dataIndex: 'Address', key: 'Address' },
        { title: 'Probe Mode', dataIndex: 'ProbeMode', key: 'ProbeMode', render: (val: string) => <Tag color="arcoblue">{val || 'ICMP'}</Tag> },
        {
            title: 'Status',
            dataIndex: 'Enabled',
            render: (enabled: boolean) => <Switch checked={enabled} disabled />,
        },
        {
            title: 'Actions',
            key: 'actions',
            render: (_: any, record: Target) => (
                <Space>
                    <Button type="text" icon={<IconEdit />} onClick={() => handleEdit(record)}>Edit</Button>
                    <Popconfirm
                        title="Delete this target?"
                        onOk={() => { if (record.ID) handleDelete(record.ID) }}
                    >
                        <Button type="text" status="danger" icon={<IconDelete />}>Delete</Button>
                    </Popconfirm>
                </Space>
            ),
        },
    ];

    return (
        <div>
            <Typography.Title heading={4} style={{ marginBottom: 24 }}>System Settings</Typography.Title>

            <Tabs defaultActiveTab='1' type='card-gutter'>
                <TabPane
                    key='1'
                    title={<span><IconUnorderedList style={{ marginRight: 6 }} />Target Management</span>}
                >
                    <div style={{ marginBottom: 20, display: 'flex', justifyContent: 'flex-end' }}>
                        <Button type="primary" icon={<IconPlus />} onClick={handleAdd}>Add Target</Button>
                    </div>
                    <Card bordered={false}>
                        <Table
                            columns={columns}
                            data={targets}
                            loading={loading}
                            rowKey="ID"
                        />
                    </Card>
                </TabPane>

                <TabPane
                    key='2'
                    title={<span><IconSafe style={{ marginRight: 6 }} />Security</span>}
                >
                    <Card title="Change Password" style={{ maxWidth: 500 }}>
                        <Form form={passForm} layout="vertical" onSubmit={handlePassUpdate}>
                            <Form.Item label="New Password" field="newPassword" rules={[{ required: true, min: 6 }]}>
                                <Input.Password placeholder="Enter new password" />
                            </Form.Item>
                            <Form.Item label="Confirm Password" field="confirm" rules={[
                                { required: true },
                                {
                                    validator: (v, cb) => {
                                        if (v !== passForm.getFieldValue('newPassword')) cb('Passwords do not match');
                                        else cb();
                                    }
                                }
                            ]}>
                                <Input.Password placeholder="Confirm new password" />
                            </Form.Item>
                            <Form.Item>
                                <Button type="primary" htmlType="submit">Update Password</Button>
                            </Form.Item>
                        </Form>
                    </Card>
                </TabPane>
            </Tabs>

            <Modal
                title={editingId ? "Edit Target" : "Add Target"}
                visible={isModalVisible}
                onOk={handleOk}
                onCancel={() => setIsModalVisible(false)}
                autoFocus={false}
                focusLock={true}
                style={{ width: 600 }}
            >
                <Form form={form} layout="vertical">
                    <Row gutter={16}>
                        <Col span={12}>
                            <Form.Item label="Name" field="Name" rules={[{ required: true }]}>
                                <Input placeholder="e.g. HK VPS" />
                            </Form.Item>
                        </Col>
                        <Col span={12}>
                            <Form.Item label="Address (IP/Domain)" field="Address" rules={[{ required: true }]}>
                                <Input placeholder="e.g. 1.2.3.4" />
                            </Form.Item>
                        </Col>
                    </Row>

                    <Row gutter={16}>
                        <Col span={12}>
                            <Form.Item label="Probe Mode" field="ProbeMode" initialValue="ICMP">
                                <Select>
                                    <Select.Option value="ICMP">ICMP/MTR (Default)</Select.Option>
                                    <Select.Option value="SSH">SSH Speed Test</Select.Option>
                                    <Select.Option value="HTTP">HTTP Download</Select.Option>
                                    <Select.Option value="IPERF3">Iperf3 Client</Select.Option>
                                </Select>
                            </Form.Item>
                        </Col>
                        <Col span={12}>
                            <Form.Item label="Probe Config" field="ProbeConfig" tooltip="SSH: 'user:key_path', HTTP: 'url', IPERF: 'port'">
                                <Input placeholder="Config string..." />
                            </Form.Item>
                        </Col>
                    </Row>

                    <Form.Item label="Description" field="Desc">
                        <Input.TextArea placeholder="Details about this target..." />
                    </Form.Item>
                    <Form.Item label="Enabled" field="Enabled" triggerPropName="checked" initialValue={true}>
                        <Switch />
                    </Form.Item>
                </Form>
            </Modal>
        </div>
    );
};

export default Settings;
