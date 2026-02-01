import React from 'react';
import ReactECharts from 'echarts-for-react';
import { Tabs } from '@arco-design/web-react';

const { TabPane } = Tabs;

interface MetricsChartProps {
    history: any[]; // MonitorRecord[]
    isDark?: boolean;
}

const MetricsChart: React.FC<MetricsChartProps> = ({ history, isDark }) => {

    const getOption = (metric: 'latency' | 'loss' | 'speed') => {
        const times = history.map(h => new Date(h.created_at || h.CreatedAt).toLocaleTimeString());

        let seriesData = [];
        let yAxisName = '';
        let type = 'line';
        let color = '#165dff';

        if (metric === 'latency') {
            seriesData = history.map(h => h.latency_ms || h.LatencyMs || 0);
            yAxisName = 'ms';
            color = '#165dff';
        } else if (metric === 'loss') {
            seriesData = history.map(h => h.packet_loss || h.PacketLoss || 0);
            yAxisName = '%';
            color = '#ff7d00';
            type = 'area';
        } else {
            seriesData = history.map(h => h.speed_down || h.SpeedDown || 0);
            yAxisName = 'Mbps';
            color = '#00b42a';
            type = 'bar';
        }

        return {
            backgroundColor: 'transparent',
            tooltip: { trigger: 'axis' },
            grid: { top: 40, bottom: 40, left: 50, right: 20 },
            xAxis: {
                type: 'category',
                data: times,
                axisLine: { lineStyle: { color: 'var(--color-border-3)' } }
            },
            yAxis: {
                type: 'value',
                name: yAxisName,
                splitLine: { lineStyle: { color: 'var(--color-border-2)' } }
            },
            series: [{
                data: seriesData,
                type: metric === 'loss' ? 'line' : type,
                areaStyle: metric === 'loss' ? { opacity: 0.3 } : undefined,
                smooth: true,
                itemStyle: { color: color }
            }]
        };
    };

    return (
        <Tabs defaultActiveTab="latency">
            <TabPane key="latency" title="Latency">
                <ReactECharts option={getOption('latency')} style={{ height: 300 }} notMerge={true} theme={isDark ? 'dark' : 'light'} />
            </TabPane>
            <TabPane key="loss" title="Packet Loss">
                <ReactECharts option={getOption('loss')} style={{ height: 300 }} notMerge={true} theme={isDark ? 'dark' : 'light'} />
            </TabPane>
            <TabPane key="speed" title="Bandwidth">
                <ReactECharts option={getOption('speed')} style={{ height: 300 }} notMerge={true} theme={isDark ? 'dark' : 'light'} />
            </TabPane>
        </Tabs>
    );
};

export default MetricsChart;
