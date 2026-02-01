import React, { useState } from 'react';
import { Form, Input, Button, Card, Message, Typography } from '@arco-design/web-react';
import { IconLock, IconUser } from '@arco-design/web-react/icon';
import { useNavigate } from 'react-router-dom';
import { login } from '../../api';

const Login: React.FC = () => {
    const [loading, setLoading] = useState(false);
    const navigate = useNavigate();
    const [form] = Form.useForm();

    const handleSubmit = async (values: any) => {
        setLoading(true);
        try {
            const res: any = await login(values.username, values.password);
            localStorage.setItem('token', res.token);
            Message.success('Login Successful');
            navigate('/dashboard');
        } catch (err) {
            console.error(err);
        } finally {
            setLoading(false);
        }
    };

    return (
        <div style={{
            display: 'flex',
            justifyContent: 'center',
            alignItems: 'center',
            height: '100vh',
            background: 'var(--color-fill-1)'
        }}>
            <Card style={{ width: 400, boxShadow: '0 4px 10px var(--color-fill-3)', borderRadius: 12 }}>
                <div style={{ textAlign: 'center', marginBottom: 24 }}>
                    <Typography.Title heading={3}>RouteLens</Typography.Title>
                    <Typography.Text type="secondary">Network Observability Platform</Typography.Text>
                </div>
                <Form form={form} onSubmit={handleSubmit} autoComplete="off" layout="vertical">
                    <Form.Item label="Username" field="username" rules={[{ required: true, message: 'Username is required' }]}>
                        <Input
                            prefix={<IconUser />}
                            placeholder="Username"
                        />
                    </Form.Item>
                    <Form.Item label="Password" field="password" rules={[{ required: true, message: 'Password is required' }]}>
                        <Input.Password
                            prefix={<IconLock />}
                            placeholder="Password"
                            onPressEnter={() => form.submit()}
                        />
                    </Form.Item>
                    <Form.Item style={{ marginTop: 24 }}>
                        <Button type="primary" htmlType="submit" long loading={loading}>
                            Login
                        </Button>
                    </Form.Item>
                </Form>
            </Card>
        </div>
    );
};

export default Login;
