import React, { useState } from 'react';
import { Form, Input, Button, Card, Typography, Message } from '@arco-design/web-react';
import { IconLock, IconUser } from '@arco-design/web-react/icon';
import { setupAdmin } from '../../api';
import { useNavigate } from 'react-router-dom';

const Setup: React.FC = () => {
    const [loading, setLoading] = useState(false);
    const navigate = useNavigate();

    const onFinish = async (values: any) => {
        setLoading(true);
        try {
            await setupAdmin(values);
            Message.success('Setup successful! Please login.');
            navigate('/login');
        } catch (e: any) {
            console.error(e);
        } finally {
            setLoading(false);
        }
    };

    return (
        <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh', background: 'var(--color-fill-1)' }}>
            <Card style={{ width: 400, borderRadius: 12, boxShadow: '0 4px 20px var(--color-fill-3)' }}>
                <div style={{ textAlign: 'center', marginBottom: 24 }}>
                    <Typography.Title heading={3}>Initial Setup</Typography.Title>
                    <Typography.Text type="secondary">Create your administrator account</Typography.Text>
                </div>
                <Form layout="vertical" onSubmit={onFinish} autoComplete="off">
                    <Form.Item label="Admin Username" field="username" rules={[{ required: true }]}>
                        <Input prefix={<IconUser />} placeholder="Username" />
                    </Form.Item>
                    <Form.Item label="Admin Password" field="password" rules={[{ required: true, min: 6 }]}>
                        <Input.Password prefix={<IconLock />} placeholder="Password" />
                    </Form.Item>
                    <Form.Item style={{ marginTop: 24 }}>
                        <Button type="primary" htmlType="submit" long loading={loading}>
                            Complete Setup
                        </Button>
                    </Form.Item>
                </Form>
            </Card>
        </div>
    );
};

export default Setup;
