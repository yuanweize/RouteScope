import React, { useEffect, useState } from 'react';
import ReactECharts from 'echarts-for-react';
import * as echarts from 'echarts';
import axios from 'axios';

interface MapChartProps {
    target?: string;
    trace?: any;
    isDark?: boolean;
}

const MapChart: React.FC<MapChartProps> = ({ target, trace, isDark }) => {
    const [isLoaded, setIsLoaded] = useState(false);

    useEffect(() => {
        const fetchMap = async () => {
            try {
                const res = await axios.get('https://raw.githubusercontent.com/apache/echarts/master/test/data/map/json/world.json');
                echarts.registerMap('world', res.data);
                setIsLoaded(true);
            } catch (err) { console.error(err); }
        };
        fetchMap();
    }, []);

    const traceData = typeof trace === 'string' ? (() => { try { return JSON.parse(trace); } catch { return null; } })() : trace;
    const hops = traceData?.hops || [];
    const points = hops
        .filter((h: any) => Number.isFinite(h.lon) && Number.isFinite(h.lat) && (h.lon !== 0 || h.lat !== 0))
        .map((h: any) => ({
            name: h.ip,
            value: [h.lon, h.lat],
            latency: h.latency_ms,
        }));

    const lineCoords = points.map((p: any) => p.value);
    const lines = lineCoords.length >= 2 ? [{
        name: target || 'trace',
        coords: lineCoords,
        lineStyle: { color: '#165dff' }
    }] : [];

    const option = {
        backgroundColor: 'transparent',
        tooltip: {
            trigger: 'item',
            formatter: (params: any) => {
                if (params.seriesType === 'effectScatter') {
                    const latency = params.data?.latency ?? 0;
                    return `Node: ${params.name}<br/>Latency: ${latency.toFixed(1)}ms`;
                }
                if (params.data && params.data.name) {
                    return `Target: ${params.data.name}`;
                }
                return params.name;
            }
        },
        geo: {
            map: 'world',
            roam: true,
            silent: false,
            itemStyle: {
                areaColor: isDark ? '#1f2329' : '#e8f3ff',
                borderColor: isDark ? '#2f3338' : '#bcd7ff'
            },
            emphasis: {
                itemStyle: { areaColor: isDark ? '#2a2f36' : '#cfe4ff' }
            }
        },
        series: [
            {
                type: 'lines',
                coordinateSystem: 'geo',
                polyline: true,
                zlevel: 1,
                effect: {
                    show: true,
                    period: 4,
                    trailLength: 0.7,
                    color: '#fff',
                    symbolSize: 3
                },
                lineStyle: {
                    width: 0,
                    curveness: 0.2
                },
                data: lines
            },
            {
                type: 'lines',
                coordinateSystem: 'geo',
                polyline: true,
                zlevel: 2,
                symbol: ['none', 'arrow'],
                symbolSize: 10,
                lineStyle: {
                    width: 2,
                    opacity: 0.6,
                    curveness: 0.2
                },
                data: lines
            },
            {
                type: 'effectScatter',
                coordinateSystem: 'geo',
                zlevel: 3,
                rippleEffect: { brushType: 'stroke' },
                label: { show: true, position: 'right', formatter: '{b}' },
                symbolSize: 10,
                itemStyle: { color: '#165dff' },
                data: points
            }
        ]
    };

    return (
        <div style={{ height: '60vh', minHeight: 600, width: '100%' }}>
            {isLoaded ? (
                <ReactECharts
                    option={option}
                    style={{ height: '100%', width: '100%' }}
                    notMerge={true}
                    theme={isDark ? 'dark' : 'light'}
                />
            ) : 'Loading Map...'}
        </div>
    );
};

export default MapChart;
