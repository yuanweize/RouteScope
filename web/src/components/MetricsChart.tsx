import React from 'react';
import ReactECharts from 'echarts-for-react';
import { Tabs } from '@arco-design/web-react';

const { TabPane } = Tabs;

interface MetricsChartProps {
    history: any[]; // MonitorRecord[]
}

const MetricsChart: React.FC<MetricsChartProps> = ({ history }) => {

    const getOption = (metric: 'latency' | 'loss' | 'speed') => {
        const times = history.map(h => new Date(h.CreatedAt).toLocaleTimeString());

        let seriesData = [];
        let yAxisName = '';
        let type = 'line';
        let color = '#165dff';

        if (metric === 'latency') {
            seriesData = history.map(h => h.LatencyMs);
            yAxisName = 'ms';
            color = '#165dff';
        } else if (metric === 'loss') {
            seriesData = history.map(h => h.PacketLoss);
            yAxisName = '%';
            color = '#ff7d00';
            type = 'area';
        } else {
            seriesData = history.map(h => h.SpeedDown);
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
                <ReactECharts option={getOption('latency')} style={{ height: 300 }} notMerge={true} />
            </TabPane>
            <TabPane key="loss" title="Packet Loss">
                <ReactECharts option={getOption('loss')} style={{ height: 300 }} notMerge={true} />
            </TabPane>
            <TabPane key="speed" title="Bandwidth">
                <ReactECharts option={getOption('speed')} style={{ height: 300 }} notMerge={true} />
            </TabPane>
        </Tabs>
    );
};

export default MetricsChart;
