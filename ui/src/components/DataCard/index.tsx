import { InfoCircleOutlined } from "@ant-design/icons";
import { Card, Tooltip, Typography } from "antd";

const { Title } = Typography;

export type DataCardProps = {
    tooltip?: string;
    title: string;
    data: number | undefined;
    extraDisplay?: React.ReactNode
    footer: React.ReactNode
}

export default function DataCard(props: DataCardProps) {
  return (
    <Card
      title={<Title style={{ margin: 0}} type="secondary" level={5}>{props.title}</Title>}
      styles={{
        header: {
          borderBottom: 0,
        },
        body: {
          padding: 0,
          paddingTop: 0
        },
      }}
      extra={props.tooltip && <Tooltip title={props.tooltip}><InfoCircleOutlined /></Tooltip>}
    >
      <Title style={{ margin: 0, padding: 24, paddingTop: 0 }}>{props.data}</Title>
      <div style={{ minHeight: 50, paddingLeft: 24, paddingRight: 24 }}>
        {props.extraDisplay}
      </div>
      <div style={{ borderTop: "1px solid #303030", padding: 12, paddingLeft: 24 }}>
        {props.footer}
      </div>
    </Card>
  );
}
