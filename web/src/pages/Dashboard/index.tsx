import React, { useEffect, useState } from 'react';
import { Card, Grid, Statistic, Button, Typography, Space } from '@arco-design/web-react';
import { IconThunderbolt } from '@arco-design/web-react/icon';
import MapChart from '../../components/MapChart';
import MetricsChart from '../../components/MetricsChart';
import { triggerProbe, getHistory, getLatestTrace } from '../../api';
import { useAppContext } from '../../utils/appContext';

const { Row, Col } = Grid;

const Dashboard: React.FC = () => {
    const [history, setHistory] = useState<any[]>([]);
    const [trace, setTrace] = useState<any>(null);
    const { selectedTarget, targets, isDark } = useAppContext();
    const selectedMeta = targets.find(t => t.address === selectedTarget);
    const isIcmpOnly = selectedMeta?.probe_type === 'MODE_ICMP';

    const fetchHistory = async () => {
        if (!selectedTarget) return;
        try {
            const data = await getHistory({ target: selectedTarget });
            setHistory(data as any[]);
        } catch (e) {
            console.error(e);
        }
    };

    const fetchTrace = async () => {
        if (!selectedTarget) return;
        try {
            const data = await getLatestTrace(selectedTarget);
            setTrace(data);
        } catch (e) {
            console.error(e);
        }
    };

    useEffect(() => {
        fetchHistory();
        fetchTrace();
    }, []);

    useEffect(() => {
        fetchHistory();
        fetchTrace();
        const timer = setInterval(() => {
            fetchHistory();
            fetchTrace();
        }, 10000);
        return () => clearInterval(timer);
    }, [selectedTarget]);

    const handleProbe = async () => {
        await triggerProbe({ target: selectedTarget } as any);
        fetchHistory();
        fetchTrace();
    };

    return (
        <div style={{ minHeight: '80vh' }}>
            <div style={{ marginBottom: 24, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <Space size="large">
                    <Typography.Title heading={4} style={{ margin: 0 }}>Network Observer</Typography.Title>
                    <Typography.Text type="secondary">Target: {selectedTarget || 'N/A'}</Typography.Text>
                </Space>
                <Button type="primary" icon={<IconThunderbolt />} onClick={handleProbe} disabled={!selectedTarget}>
                    Quick Probe
                </Button>
            </div>

            <Row gutter={24} style={{ marginBottom: 24 }}>
                <Col span={6}>
                    <Card bordered={false} hoverable>
                        <Statistic
                            title="Avg Latency"
                            value={history.length > 0 ? (history.reduce((a, b) => a + (b.latency_ms || b.LatencyMs || 0), 0) / history.length) : 0}
                            precision={1}
                            suffix="ms"
                            groupSeparator
                        />
                    </Card>
                </Col>
                <Col span={6}>
                    <Card bordered={false} hoverable>
                        <Statistic
                            title="Packet Loss"
                            value={history.length > 0 ? (history.reduce((a, b) => a + (b.packet_loss || b.PacketLoss || 0), 0) / history.length) : 0}
                            precision={2}
                            suffix="%"
                            style={{ color: 'var(--color-danger-text)' }}
                        />
                    </Card>
                </Col>
                <Col span={6}>
                    <Card bordered={false} hoverable>
                        {isIcmpOnly ? (
                            <Statistic
                                title="Downlink"
                                value="N/A"
                            />
                        ) : (
                            <Statistic
                                title="Downlink"
                                value={history.length > 0 ? (history[history.length - 1].speed_down || history[history.length - 1].SpeedDown || 0) : 0}
                                precision={1}
                                suffix="Mbps"
                            />
                        )}
                    </Card>
                </Col>
                <Col span={6}>
                    <Card bordered={false} hoverable>
                        <Statistic title="Monitoring Nodes" value={targets.length} />
                    </Card>
                </Col>
            </Row>

            <Row gutter={24}>
                <Col span={16}>
                    <Card title="Traffic Path Visualization" bordered={false} extra={<Typography.Text type="secondary">Real-time Path</Typography.Text>}>
                        <MapChart target={selectedTarget} trace={trace} isDark={isDark} />
                    </Card>
                </Col>
                <Col span={8}>
                    <Card title="Connectivity Trends" bordered={false}>
                        <MetricsChart history={history} isDark={isDark} />
                    </Card>
                </Col>
            </Row>
        </div>
    );
};

export default Dashboard;
