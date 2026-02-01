import React from 'react';
import { Card, Table, Tag, Typography } from '@arco-design/web-react';
import { useAppContext } from '../../utils/appContext';

const Targets: React.FC = () => {
  const { targets } = useAppContext();

  return (
    <Card title="Targets" bordered={false}>
      <Table
        data={targets}
        pagination={false}
        columns={[
          {
            title: 'Name',
            dataIndex: 'name',
          },
          {
            title: 'Address',
            dataIndex: 'address',
          },
          {
            title: 'Probe Type',
            dataIndex: 'probe_type',
            render: (val) => <Tag color="blue">{val}</Tag>
          },
          {
            title: 'Status',
            dataIndex: 'enabled',
            render: (val) => val ? <Tag color="green">Enabled</Tag> : <Tag color="red">Disabled</Tag>
          },
          {
            title: 'Description',
            dataIndex: 'desc',
            render: (val) => <Typography.Text type="secondary">{val}</Typography.Text>
          }
        ]}
      />
    </Card>
  );
};

export default Targets;
