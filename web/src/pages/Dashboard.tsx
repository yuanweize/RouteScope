import React, { useEffect, useState, useMemo } from 'react';
import { Card, Col, Row, Statistic, Select, Typography, Collapse, Table, Tag, Space, Tooltip } from 'antd';
import { CheckCircleOutlined, CloseCircleOutlined, MinusCircleOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import { useRequest } from 'ahooks';
import { useTranslation } from 'react-i18next';
import { getHistory, getLatestTrace, getTargets } from '../api';
import type { Target } from '../api';
import MapChart from '../components/MapChart';
import MetricsChart from '../components/MetricsChart';
import { useTheme } from '../context/ThemeContext';

interface HopRow {
  key: number;
  hop: number;
  host: string;
  ip: string;
  location: string;
  subdiv: string;
  isp: string;
  loss: number;
  last: number;
  avg: number;
  best: number;
  worst: number;
  asn: string;
  isTimeout: boolean;
  geoPrecision: string;
}

const Dashboard: React.FC = () => {
  const { isDark } = useTheme();
  const { t, i18n } = useTranslation();
  const [selectedTarget, setSelectedTarget] = useState<string>('');
  const [trace, setTrace] = useState<any>(null);

  const { data: targets = [] } = useRequest(getTargets);

  // Find the selected target metadata (for error display)
  const selectedMeta: Target | undefined = useMemo(
    () => targets.find((t: Target) => t.address === selectedTarget),
    [targets, selectedTarget]
  );

  useEffect(() => {
    if (targets.length > 0 && !selectedTarget) {
      setSelectedTarget(targets[0].address);
    }
  }, [targets, selectedTarget]);

  const { data: history = [] } = useRequest(() => getHistory({ target: selectedTarget }), {
    refreshDeps: [selectedTarget],
    ready: !!selectedTarget,
  });

  // Re-fetch trace when language changes for localized location names
  useRequest(
    () => getLatestTrace(selectedTarget, i18n.language),
    {
      refreshDeps: [selectedTarget, i18n.language],
      ready: !!selectedTarget,
      onSuccess: (data) => setTrace(data),
    }
  );

  const avgLatency = history.length
    ? history.reduce((sum: number, h: any) => sum + (h.latency_ms || h.LatencyMs || 0), 0) / history.length
    : 0;
  const avgLoss = history.length
    ? history.reduce((sum: number, h: any) => sum + (h.packet_loss || h.PacketLoss || 0), 0) / history.length
    : 0;
  const lastSpeed = history.length
    ? history[history.length - 1].speed_down || history[history.length - 1].SpeedDown || 0
    : 0;

  // Speed test status indicator
  const isSpeedEnabled = selectedMeta && selectedMeta.probe_type !== 'MODE_ICMP' && selectedMeta.probe_type !== '';
  const hasSpeedError = isSpeedEnabled && selectedMeta?.last_error;
  const speedStatusIcon = useMemo(() => {
    if (!isSpeedEnabled) return null;
    if (lastSpeed > 0) {
      return <CheckCircleOutlined style={{ color: '#52c41a', marginLeft: 8 }} />;
    }
    if (hasSpeedError) {
      return (
        <Tooltip title={selectedMeta?.last_error} color="red">
          <CloseCircleOutlined style={{ color: '#ff4d4f', marginLeft: 8, cursor: 'help' }} />
        </Tooltip>
      );
    }
    return <MinusCircleOutlined style={{ color: '#faad14', marginLeft: 8 }} />;
  }, [isSpeedEnabled, lastSpeed, hasSpeedError, selectedMeta?.last_error]);

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

  // Format location based on current language
  const formatLocation = (hop: any) => {
    const lang = i18n.language;
    if (lang === 'zh-CN' || lang === 'zh') {
      // Chinese: use default fields (already zh-CN or fallback)
      return [hop.city, hop.subdiv, hop.country].filter(Boolean).join(', ');
    } else {
      // English: prefer *_en fields if available
      const city = hop.city_en || hop.city;
      const subdiv = hop.subdiv_en || hop.subdiv;
      const country = hop.country_en || hop.country;
      return [city, subdiv, country].filter(Boolean).join(', ');
    }
  };

  const renderLatencyTag = (value: any, color?: string) => {
    if (typeof value === 'number' && value > 0) {
      return <Tag color={color}>{`${value.toFixed(1)}ms`}</Tag>;
    }
    return <Typography.Text type="secondary" style={{ fontSize: 12 }}>{t('common.na')}</Typography.Text>;
  };

  const renderLoss = (value: any) => {
    if (typeof value === 'number') {
      const color = value > 50 ? '#ff4d4f' : value > 10 ? '#faad14' : undefined;
      return <span style={{ color }}>{value.toFixed(1)}%</span>;
    }
    return <Typography.Text type="secondary">{t('common.na')}</Typography.Text>;
  };

  const hopRows: HopRow[] = (traceData?.hops || []).map((hop: any) => {
    const isTimeout = hop.loss === 100 || hop.ip === '*' || hop.host === '???';
    return {
      key: hop.hop,
      hop: hop.hop,
      host: hop.host || hop.ip,
      ip: hop.ip,
      location: formatLocation(hop),
      subdiv: hop.subdiv || '',
      isp: hop.isp,
      loss: hop.loss,
      last: hop.latency_last_ms,
      avg: hop.latency_avg_ms,
      best: hop.latency_best_ms,
      worst: hop.latency_worst_ms,
      asn: hop.asn,
      isTimeout,
      geoPrecision: hop.geo_precision || '',
    };
  });

  const hopColumns: ColumnsType<HopRow> = [
    { title: t('hopTable.hop'), dataIndex: 'hop', width: 50, align: 'center' },
    {
      title: t('hopTable.ipHost'),
      dataIndex: 'host',
      width: 200,
      ellipsis: true,
      render: (val: string, row: HopRow) => (
        <Typography.Text copyable={{ text: row.ip }} style={{ opacity: row.isTimeout ? 0.4 : 1 }}>
          {val || '???'}
        </Typography.Text>
      ),
    },
    {
      title: t('hopTable.location'),
      width: 220,
      render: (_: any, row: HopRow) => (
        <div style={{ opacity: row.isTimeout ? 0.4 : 1 }}>
          <div>{row.location || '-'}</div>
          {row.isp && <Typography.Text type="secondary" style={{ fontSize: 11 }}>{row.isp}</Typography.Text>}
          {row.geoPrecision === 'country' && <Tag color="orange" style={{ marginLeft: 4, fontSize: 10 }}>{t('hopTable.countryOnly')}</Tag>}
        </div>
      ),
    },
    {
      title: t('hopTable.loss'),
      dataIndex: 'loss',
      width: 80,
      align: 'right',
      render: (val: number, row: HopRow) => (
        <span style={{ opacity: row.isTimeout ? 0.4 : 1 }}>{renderLoss(val)}</span>
      ),
    },
    {
      title: t('hopTable.latency'),
      width: 280,
      render: (_: any, row: HopRow) => (
        <Space style={{ opacity: row.isTimeout ? 0.4 : 1 }}>
          {renderLatencyTag(row.last, 'green')}
          {renderLatencyTag(row.avg, 'blue')}
          {renderLatencyTag(row.best)}
          {renderLatencyTag(row.worst)}
        </Space>
      ),
    },
    {
      title: t('hopTable.asn'),
      dataIndex: 'asn',
      width: 100,
      render: (val: string, row: HopRow) => (
        <Typography.Text type="secondary" style={{ opacity: row.isTimeout ? 0.4 : 1 }}>{val || '-'}</Typography.Text>
      ),
    },
  ];

  return (
    <div>
      <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
        <Col span={24}>
          <Typography.Title level={4} style={{ margin: 0 }}>{t('dashboard.title')}</Typography.Title>
        </Col>
      </Row>

      <Row gutter={[16, 16]}>
        <Col span={8}>
          <Card className="page-card">
            <Statistic title={t('dashboard.avgLatency')} value={avgLatency} suffix="ms" precision={1} />
          </Card>
        </Col>
        <Col span={8}>
          <Card className="page-card">
            <Statistic title={t('dashboard.packetLoss')} value={avgLoss} suffix="%" precision={2} />
          </Card>
        </Col>
        <Col span={8}>
          <Card className="page-card">
            <div style={{ display: 'flex', alignItems: 'center' }}>
              <Statistic
                title={t('dashboard.downlink')}
                value={isSpeedEnabled ? lastSpeed : t('common.na')}
                suffix={isSpeedEnabled ? 'Mbps' : ''}
                precision={1}
              />
              {speedStatusIcon}
            </div>
            {hasSpeedError && (
              <Typography.Text type="danger" style={{ fontSize: 11, display: 'block', marginTop: 4 }}>
                {selectedMeta?.last_error}
              </Typography.Text>
            )}
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
        <Col span={16}>
          <Card
            className="chart-card"
            title={t('dashboard.routeMap')}
            extra={
              <Select
                style={{ width: 240 }}
                value={selectedTarget}
                onChange={setSelectedTarget}
                options={targets.map((tgt: Target) => ({ label: `${tgt.name} (${tgt.address})`, value: tgt.address }))}
              />
            }
          >
            <MapChart trace={trace} isDark={isDark} />
          </Card>
          <Card className="chart-card" style={{ marginTop: 16 }}>
            <Collapse
              defaultActiveKey={['hops']}
              items={[
                {
                  key: 'hops',
                  label: (
                    <Space>
                      <span>{t('dashboard.mtrHopDetails')}</span>
                      {traceData?.truncated ? <Tag color="orange">{t('dashboard.truncated')}</Tag> : null}
                    </Space>
                  ),
                  children: (
                    <Table
                      size="small"
                      dataSource={hopRows}
                      pagination={false}
                      columns={hopColumns}
                      rowClassName={(row: HopRow) => (row.isTimeout ? 'hop-timeout-row' : '')}
                      scroll={{ x: 900 }}
                    />
                  ),
                },
              ]}
            />
          </Card>
        </Col>
        <Col span={8}>
          <Card className="chart-card" title={t('dashboard.historicalMetrics')}>
            <MetricsChart history={history} isDark={isDark} />
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default Dashboard;
