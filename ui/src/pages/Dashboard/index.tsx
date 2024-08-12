import { Col, Row, Typography } from "antd";
import DataCard from "../../components/DataCard";
import { CaretDownOutlined, CaretUpOutlined } from "@ant-design/icons";
import DashboardTable from "../../components/DashboardTable";
import { useQuery } from "@tanstack/react-query";
import { getResources, Resource } from "../../api/resources";
import { Approval, getApprovals } from "../../api/approvals";
import { getStats, Stat } from "../../api/stats";
import {
  Bar,
  BarChart,
  ResponsiveContainer,
  XAxis,
  YAxis,
  Tooltip as ChartTooltip,
} from "recharts";

const { Text } = Typography;

export default function Dashboard() {
  const resources = useQuery<Resource[]>({
    queryKey: ["resources"],
    queryFn: getResources,
  });
  const stats = useQuery<Stat[]>({ queryKey: ["stats"], queryFn: getStats });
  const approvals = useQuery<Approval[]>({
    queryKey: ["approvals"],
    queryFn: getApprovals,
  });

  const totalUpdatesThisPeriod = stats.data?.reduce(
    (acc, curr) => acc + curr.updates,
    0
  );
  const approvalsPending = approvals.data?.filter(
    ({ rejected, archived, votesReceived, votesRequired }) =>
      !rejected && !archived && votesReceived < votesRequired
  ).length;
  const approvalsApproved = approvals.data?.filter(
    ({ rejected, votesReceived, votesRequired }) =>
      !rejected && votesReceived >= votesRequired
  ).length;
  const approvalsRejected = approvals.data?.filter(
    ({ rejected }) => rejected
  ).length;

  return (
    <>
      <Row
        style={{
          backgroundColor: "#141414",
          padding: 16,
          paddingLeft: 0,
          paddingRight: 0,
          marginLeft: 8,
          marginRight: 8,
        }}
        gutter={16}
      >
        <Col span={6}>
          <DataCard
            data={resources.data?.length}
            title="Cluster Resources"
            tooltip="Cluster resources managed by Quilla"
            footer={
              <Text type="secondary">
                Managed by Quilla:{" "}
                {
                  resources.data?.filter(
                    (resource: Resource) => resource.policy !== "nil policy"
                  ).length
                }
              </Text>
            }
          />
        </Col>
        <Col span={6}>
          <DataCard
            data={resources.data?.reduce(
              (acc, curr) => acc + curr.status.replicas,
              0
            )}
            title="Total pods in cluster"
            // @ts-ignore
            extraDisplay={
              <Text>
                Healthy{" "}
                <Text strong>
                  {resources.data?.reduce(
                    (acc, curr) => acc + curr.status.availableReplicas,
                    0
                  )}
                </Text>{" "}
                <CaretUpOutlined style={{ color: "green" }} /> Unavailable{" "}
                <Text strong>
                  {resources.data?.reduce(
                    (acc, curr) => acc + curr.status.unavailableReplica,
                    0
                  )}
                </Text>{" "}
                <CaretDownOutlined style={{ color: "red" }} />
              </Text>
            }
            footer={<Text type="secondary">Percent up: 100%</Text>}
          />
        </Col>
        <Col span={6}>
          <DataCard
            data={totalUpdatesThisPeriod}
            title="Updates"
            tooltip="Weekly Updates"
            extraDisplay={
              <ResponsiveContainer height={50}>
                <BarChart
                  data={stats.data
                    ?.sort(
                      (a, b) =>
                        new Date(a.date).getTime() - new Date(b.date).getTime()
                    )
                    .map((stat) => ({
                      x: stat.date,
                      y: stat.updates,
                    }))}
                >
                  <XAxis hide />
                  <YAxis hide />
                  <ChartTooltip
                    cursor={{ fill: "#303030" }}
                    contentStyle={{
                      backgroundColor: "#141414",
                      borderColor: "#303030",
                      borderRadius: 4,
                    }}
                    labelFormatter={(_, payload) => {
                      if (!payload || payload.length === 0) return <></>;
                      return <Text>{payload[0].payload.x}</Text>;
                    }}
                    formatter={(value) => {
                      return [value];
                    }}
                  />
                  <Bar dataKey="y" fill="#1668dc" />
                  <Bar dataKey="x" hide />
                </BarChart>
              </ResponsiveContainer>
            }
            footer={
              <Text type="secondary">
                {/* @ts-ignore */}
                Average {totalUpdatesThisPeriod / 4} updates per week
              </Text>
            }
          />
        </Col>
        <Col span={6}>
          <DataCard
            data={approvalsPending}
            title="Pending Approvals"
            tooltip="Current approvals waiting for action"
            footer={
              <Text type="secondary">
                A <Text strong>{approvalsApproved}</Text>{" "}
                <CaretUpOutlined style={{ color: "green" }} /> R{" "}
                <Text strong>{approvalsRejected}</Text>{" "}
                <CaretDownOutlined style={{ color: "red" }} />
              </Text>
            }
            extraDisplay={
              <ResponsiveContainer height={50}>
                <BarChart
                  data={stats.data
                    ?.sort(
                      (a, b) =>
                        new Date(a.date).getTime() - new Date(b.date).getTime()
                    )
                    .map((stat) => ({
                      x: stat.date,
                      y: stat.approved,
                    }))}
                >
                  <XAxis hide />
                  <YAxis hide />
                  <ChartTooltip
                    cursor={{ fill: "#303030" }}
                    contentStyle={{
                      backgroundColor: "#141414",
                      borderColor: "#303030",
                      borderRadius: 4,
                    }}
                    labelFormatter={(_, payload) => {
                      if (!payload || payload.length === 0) return <></>;
                      return <Text>{payload[0].payload.x}</Text>;
                    }}
                    formatter={(value) => {
                      return [value];
                    }}
                  />
                  <Bar dataKey="y" fill="#1668dc" />
                  <Bar dataKey="x" hide />
                </BarChart>
              </ResponsiveContainer>
            }
          />
        </Col>
      </Row>
      <DashboardTable />
    </>
  );
}
