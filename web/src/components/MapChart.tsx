import React, { useEffect, useState } from 'react';
import ReactECharts from 'echarts-for-react';
import * as echarts from 'echarts';
import axios from 'axios';

interface MapChartProps {
    target?: string;
}

const MapChart: React.FC<MapChartProps> = ({ target }) => {
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

    // Simulated path data for demonstration
    const allLines = [
        {
            name: 'nas.yuanweize.win',
            coords: [[116.46, 39.92], [114.17, 22.28], [103.85, 1.29], [2.35, 48.85]], // Beijing -> HK -> SG -> Paris
            lineStyle: { color: '#165dff' }
        },
        {
            name: 'nue.eurun.top',
            coords: [[116.46, 39.92], [-122.41, 37.77], [-74.00, 40.71]], // Beijing -> SF -> NY
            lineStyle: { color: '#00b42a' }
        }
    ];

    // Filter based on selected target address
    const filteredLines = target ? allLines.filter(l => l.name === target) : allLines;

    const option = {
        backgroundColor: 'transparent',
        tooltip: {
            trigger: 'item',
            formatter: (params: any) => {
                if (params.seriesType === 'effectScatter') {
                    return `Node: ${params.name}<br/>Latency: ${Math.floor(Math.random() * 50)}ms`;
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
                areaColor: 'var(--color-fill-2)',
                borderColor: 'var(--color-border-2)'
            },
            emphasis: {
                itemStyle: { areaColor: 'var(--color-fill-3)' }
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
                data: filteredLines
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
                data: filteredLines
            },
            {
                type: 'effectScatter',
                coordinateSystem: 'geo',
                zlevel: 3,
                rippleEffect: { brushType: 'stroke' },
                label: { show: true, position: 'right', formatter: '{b}' },
                symbolSize: 10,
                itemStyle: { color: '#165dff' },
                data: [
                    { name: 'Source', value: [116.46, 39.92] },
                    { name: 'Target', value: filteredLines.length > 0 ? filteredLines[0].coords[filteredLines[0].coords.length - 1] : [0, 0] }
                ]
            }
        ]
    };

    return (
        <div style={{ height: '600px', width: '100%' }}>
            {isLoaded ? <ReactECharts option={option} style={{ height: '100%', width: '100%' }} /> : 'Loading Map...'}
        </div>
    );
};

export default MapChart;
