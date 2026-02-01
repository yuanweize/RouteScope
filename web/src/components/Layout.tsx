import React, { useEffect } from 'react';
import { Layout, Menu, Button, Space, Typography, Select, Avatar, Dropdown, ConfigProvider } from '@arco-design/web-react';
import { IconDashboard, IconSettings, IconMoonFill, IconSunFill, IconPublic, IconUser } from '@arco-design/web-react/icon';
import { useNavigate, useLocation } from 'react-router-dom';
import '@arco-design/web-react/dist/css/arco.css';
import { AppContextProvider, useAppContext } from '../utils/appContext';

const { Header, Footer, Sider, Content } = Layout;
const MenuItem = Menu.Item;

const AppLayoutInner: React.FC<{ children: React.ReactNode }> = ({ children }) => {
    const navigate = useNavigate();
    const location = useLocation();
    const { isDark, setIsDark, targets, selectedTarget, setSelectedTarget, refreshTargets } = useAppContext();

    useEffect(() => {
        const stored = localStorage.getItem('theme');
        if (stored) {
            setIsDark(stored === 'dark');
        } else {
            const darkModeMediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
            setIsDark(darkModeMediaQuery.matches);
            const listener = (e: MediaQueryListEvent) => setIsDark(e.matches);
            darkModeMediaQuery.addEventListener('change', listener);
            return () => darkModeMediaQuery.removeEventListener('change', listener);
        }
    }, [setIsDark]);

    useEffect(() => {
        refreshTargets();
    }, [refreshTargets]);

    useEffect(() => {
        if (isDark) {
            document.body.setAttribute('arco-theme', 'dark');
            localStorage.setItem('theme', 'dark');
        } else {
            document.body.removeAttribute('arco-theme');
            localStorage.setItem('theme', 'light');
        }
    }, [isDark]);

    const toggleTheme = () => setIsDark(!isDark);

    return (
        <ConfigProvider theme={isDark ? { theme: 'dark' } : undefined}>
            <Layout className="layout-container" style={{ minHeight: '100vh' }}>
                <Sider
                    breakpoint="lg"
                    onBreakpoint={() => { }}
                    collapsible
                    theme={isDark ? 'dark' : 'light'}
                >
                    <div className="logo" style={{ height: 40, margin: '16px', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                        <IconPublic style={{ fontSize: 20, marginRight: 8, color: '#165dff' }} />
                        <Typography.Text bold style={{ color: isDark ? '#fff' : '#000' }}>RouteLens</Typography.Text>
                    </div>
                    <Menu
                        selectedKeys={[location.pathname]}
                        onClickMenuItem={(key) => navigate(key)}
                        style={{ width: '100%' }}
                    >
                        <MenuItem key="/dashboard">
                            <IconDashboard /> Dashboard
                        </MenuItem>
                        <MenuItem key="/targets">
                            <IconPublic /> Targets
                        </MenuItem>
                        <MenuItem key="/settings">
                            <IconSettings /> Settings
                        </MenuItem>
                    </Menu>
                </Sider>
                <Layout>
                    <Header
                        style={{
                            padding: '0 20px',
                            background: isDark ? '#17171a' : '#fff',
                            display: 'flex',
                            alignItems: 'center',
                            justifyContent: 'space-between',
                            borderBottom: '1px solid var(--color-border)'
                        }}
                    >
                        <Space size="large">
                            <Typography.Title heading={5} style={{ margin: 0 }}>RouteLens</Typography.Title>
                            <Select
                                placeholder='Select Target'
                                style={{ width: 260 }}
                                value={selectedTarget}
                                onChange={setSelectedTarget}
                            >
                                {targets.map(t => (
                                    <Select.Option key={t.address} value={t.address}>
                                        {t.name} ({t.address})
                                    </Select.Option>
                                ))}
                            </Select>
                        </Space>
                        <Space>
                            <Button
                                shape="circle"
                                type="secondary"
                                icon={isDark ? <IconSunFill /> : <IconMoonFill />}
                                onClick={toggleTheme}
                            />
                            <Dropdown
                                droplist={
                                    <Menu>
                                        <MenuItem key="profile">Profile</MenuItem>
                                        <MenuItem key="logout">Logout</MenuItem>
                                    </Menu>
                                }
                            >
                                <Avatar size={32} style={{ backgroundColor: '#165dff' }}>
                                    <IconUser />
                                </Avatar>
                            </Dropdown>
                        </Space>
                    </Header>
                    <Content style={{ padding: '24px', background: 'var(--color-fill-1)' }}>
                        {children}
                    </Content>
                    <Footer style={{ textAlign: 'center', padding: 20, background: 'var(--color-bg-1)' }}>
                        RouteLens ©2026 Admin Panel
                    </Footer>
                </Layout>
            </Layout>
        </ConfigProvider>
    );
};

const AppLayout: React.FC<{ children: React.ReactNode }> = ({ children }) => (
    <AppContextProvider>
        <AppLayoutInner>{children}</AppLayoutInner>
    </AppContextProvider>
);

export default AppLayout;
