import React, { useEffect, useMemo, useState } from 'react';
import ReactECharts from 'echarts-for-react';
import * as echarts from 'echarts';
import axios from 'axios';

interface MapChartProps {
  trace?: any;
  isDark: boolean;
}

const MapChart: React.FC<MapChartProps> = ({ trace, isDark }) => {
  const [ready, setReady] = useState(false);

  useEffect(() => {
    const fetchMap = async () => {
      const res = await axios.get('https://raw.githubusercontent.com/apache/echarts/master/test/data/map/json/world.json');
      echarts.registerMap('world', res.data);
      setReady(true);
    };
    fetchMap();
  }, []);

  const traceData = useMemo(() => {
    if (!trace) return null;
    if (typeof trace === 'string') {
      try {
        return JSON.parse(trace);
      } catch {
        return null;
      }
    }
    return trace;
  }, [trace]);

  // High-Precision Mode: Filter out low-precision nodes (country-level only)
  // Only include nodes with city or subdivision precision for accurate map lines
  const highPrecisionPoints = useMemo(() => {
    const hops = traceData?.hops || [];
    return hops
      .filter((h: any) => {
        if (!Number.isFinite(h.lon) || !Number.isFinite(h.lat)) return false;
        if (h.lon === 0 && h.lat === 0) return false;
        const precision = h.geo_precision || '';
        if (precision === 'country' || precision === 'none') return false;
        if (!precision && !h.city && !h.subdiv) return false;
        return true;
      })
      .map((h: any) => ({
        name: h.city || h.subdiv || h.host || h.ip,
        value: [h.lon, h.lat],
        latency: h.latency_last_ms || h.latency_avg_ms || h.latency_ms || 0,
        hop: h.hop,
      }));
  }, [traceData]);

  // All points for scatter display (including low-precision)
  const allPoints = useMemo(() => {
    const hops = traceData?.hops || [];
    return hops
      .filter((h: any) => Number.isFinite(h.lon) && Number.isFinite(h.lat) && (h.lon !== 0 || h.lat !== 0))
      .map((h: any) => ({
        name: h.city || h.subdiv || h.host || h.ip,
        value: [h.lon, h.lat],
        latency: h.latency_last_ms || h.latency_avg_ms || h.latency_ms || 0,
        hop: h.hop,
        precision: h.geo_precision || (h.city ? 'city' : h.subdiv ? 'subdivision' : 'country'),
      }));
  }, [traceData]);

  const colorForLatency = (latency: number) => {
    if (latency > 200) return '#ff4d4f';
    if (latency > 100) return '#faad14';
    return '#52c41a';
  };

  // Build line segments using HIGH PRECISION points only
  const segments = useMemo(() => {
    const segs: any[] = [];
    for (let i = 0; i < highPrecisionPoints.length - 1; i += 1) {
      const curr = highPrecisionPoints[i];
      const next = highPrecisionPoints[i + 1];
      segs.push({
        name: traceData?.target || 'trace',
        coords: [curr.value, next.value],
        lineStyle: { color: colorForLatency(next.latency) },
      });
    }
    return segs;
  }, [highPrecisionPoints, traceData]);

  const option = {
    backgroundColor: 'transparent',
    tooltip: {
      trigger: 'item',
      formatter: (params: any) => {
        if (params.seriesType === 'effectScatter') {
          const latency = params.data?.latency ?? 0;
          const precision = params.data?.precision || '';
          const precisionLabel = precision === 'city' ? '🎯 City' : precision === 'subdivision' ? '📍 Province' : '🌐 Country';
          return `<b>${params.name}</b><br/>Hop: ${params.data?.hop || '-'}<br/>Latency: ${latency.toFixed(1)}ms<br/>Precision: ${precisionLabel}`;
        }
        return params.name;
      },
    },
    geo: {
      map: 'world',
      roam: true,
      itemStyle: {
        areaColor: isDark ? '#1f1f1f' : '#f0f5ff',
        borderColor: isDark ? '#2f2f2f' : '#d6e4ff',
      },
      emphasis: {
        itemStyle: { areaColor: isDark ? '#2a2a2a' : '#d6e4ff' },
      },
    },
    series: [
      {
        type: 'lines',
        coordinateSystem: 'geo',
        polyline: true,
        zlevel: 1,
        effect: { show: true, period: 4, trailLength: 0.7, color: '#fff', symbolSize: 3 },
        lineStyle: { width: 0, curveness: 0.2 },
        data: segments,
      },
      {
        type: 'lines',
        coordinateSystem: 'geo',
        polyline: true,
        zlevel: 2,
        symbol: ['none', 'arrow'],
        symbolSize: 10,
        lineStyle: { width: 2, opacity: 0.7, curveness: 0.2 },
        data: segments,
      },
      {
        type: 'effectScatter',
        coordinateSystem: 'geo',
        zlevel: 3,
        rippleEffect: { brushType: 'stroke' },
        label: {
          show: true,
          position: 'right',
          formatter: (params: any) => (params.data?.precision === 'country' ? '' : params.name),
        },
        symbolSize: (_val: any, params: any) => (params.data?.precision === 'country' ? 6 : 10),
        itemStyle: { color: (params: any) => (params.data?.precision === 'country' ? '#666' : '#1677ff') },
        data: allPoints,
      },
    ],
  };

  return (
    <div className="map-container">
      {ready && <ReactECharts option={option} style={{ height: '100%', width: '100%' }} notMerge={true} theme={isDark ? 'dark' : 'light'} />}
    </div>
  );
};

export default MapChart;
