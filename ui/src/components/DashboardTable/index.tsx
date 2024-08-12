import {
  Button,
  Dropdown,
  Flex,
  MenuProps,
  Space,
  Switch,
  TableProps,
  Tag,
  Tooltip,
  Typography,
} from "antd";
import {
  DownOutlined,
  LinkOutlined,
  PauseOutlined,
  SyncOutlined,
  UpOutlined,
} from "@ant-design/icons";
import { getResources, Resource } from "../../api/resources";
import {
  useMutation,
  UseMutationResult,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";
import { PolicyPayload, setPolicy } from "../../api/policy";
import { setTracking, TrackingPayload } from "../../api/tracking";
import { ApprovalPayload, setApproval } from "../../api/approvals";
import { useState } from "react";
import DataTable from "../DataTable";

const Text = Typography.Text;

const dropdownMenu: MenuProps["items"] = [
  {
    label: "patch",
    key: "1",
  },
  {
    label: "minor",
    key: "2",
  },
  {
    label: "major",
    key: "3",
  },
  {
    label: "all",
    key: "4",
  },
  {
    label: "force",
    key: "5",
  },
  {
    label: "glob",
    key: "6",
  },
  {
    label: "regexp",
    key: "7",
  },
];

const columns: (
  setPolicy: UseMutationResult<any, Error, PolicyPayload, unknown>,
  setTracking: UseMutationResult<any, Error, TrackingPayload, unknown>,
  setApproval: UseMutationResult<any, Error, ApprovalPayload, unknown>
) => TableProps<Resource>["columns"] = (
  setPolicy: UseMutationResult<any, Error, PolicyPayload, unknown>,
  setTracking: UseMutationResult<any, Error, TrackingPayload, unknown>,
  setApproval: UseMutationResult<any, Error, ApprovalPayload, unknown>
) => [
  {
    title: "Namespace",
    dataIndex: "namespace",
    key: "namespace",
  },
  {
    title: "Name",
    dataIndex: "identifier",
    key: "identifier",
  },
  {
    title: "Pods",
    dataIndex: "status",
    key: "status",
    render: (_, { status }) => (
      <>
        <Text>
          <Text
            style={{
              color:
                status.availableReplicas === status.replicas ? "green" : "red",
            }}
          >
            •
          </Text>{" "}
          {status.availableReplicas}/{status.replicas}
        </Text>
      </>
    ),
  },
  {
    title: "Policy",
    dataIndex: "policy",
    key: "policy",
    render: (_, { policy }) => (
      <Text>
        <Text style={{ color: policy !== "nil policy" ? "green" : "gray" }}>
          •
        </Text>{" "}
        {policy === "nil policy" ? "none" : policy}
      </Text>
    ),
  },
  {
    title: "Required Approvals",
    dataIndex: "",
    key: "",
    render: (_, resource) => (
      <Text>
        {
          // @ts-ignore
          resource.annotations["quilla.sh/approvals"]
            ? // @ts-ignore
              resource.annotations["quilla.sh/approvals"]
            : "-"
        }
      </Text>
    ),
  },
  {
    title: "Images",
    dataIndex: "images",
    key: "images",
    render: (_, { images }) => (
      <Flex vertical style={{ width: "min-content" }}>
        {images.map((image, index) => (
          <Tag key={index}>{image}</Tag>
        ))}
      </Flex>
    ),
  },
  {
    title: "Quilla Labels & Annotations",
    dataIndex: "labels",
    key: "labels",
    // @ts-ignore
    render: (_, { labels, annotations }) => (
      <Flex vertical>
        {Object.keys(labels)
          .filter((label) => label.includes("quilla.sh"))
          .map((label, index) => (
            <Tag key={`l${index}`} style={{ width: "min-content" }}>
              {label}: {labels[label]}
            </Tag>
          ))}
        {Object.keys(annotations)
          .filter((label) => label.includes("quilla.sh"))
          .map((annotation, index) => (
            <Tag key={`a${index}`} style={{ width: "min-content" }}>
              {annotation}: {annotations[annotation]}
            </Tag>
          ))}
      </Flex>
    ),
  },
  {
    title: "Policy & Approvals Control",
    dataIndex: "",
    key: "",
    render: (_, resource) => {
      return (
        <Flex>
          <Space>
            <Button
              size="small"
              disabled={resource.policy === "nil policy"}
              onClick={() =>
                setPolicy.mutate({
                  identifier: resource.identifier,
                  provider: resource.provider,
                  policy: "never",
                })
              }
              icon={<PauseOutlined />}
            >
              Pause
            </Button>
            <Dropdown
              menu={{
                items: dropdownMenu,
                onClick: (info) => {
                  // @ts-ignore
                  const policy = dropdownMenu[Number(info.key) - 1].label;
                  setPolicy.mutate({
                    policy,
                    identifier: resource.identifier,
                    provider: resource.provider,
                  });
                },
              }}
            >
              <Button size="small">
                <Space>
                  Policy
                  <DownOutlined />
                </Space>
              </Button>
            </Dropdown>
            <Flex>
              <Button
                size="small"
                style={{
                  borderRadius: "4px 0px 0px 4px",
                  borderRightWidth: 1,
                  marginRight: -1,
                }}
                onClick={() => {
                  // @ts-ignore
                  const current = resource.annotations["quilla.sh/approvals"];
                  setApproval.mutate({
                    identifier: resource.identifier,
                    provider: resource.provider,
                    votesRequired: current ? Number(current) + 1 : 1,
                  });
                }}
              >
                <UpOutlined />
              </Button>
              <Button
                size="small"
                style={{ borderRadius: "0px 4px 4px 0px" }}
                onClick={() => {
                  // @ts-ignore
                  const current = resource.annotations["quilla.sh/approvals"];
                  setApproval.mutate({
                    identifier: resource.identifier,
                    provider: resource.provider,
                    votesRequired: current > 1 ? Number(current) - 1 : 0,
                  });
                }}
              >
                <DownOutlined />
              </Button>
            </Flex>
            <Tooltip title="Enable or disable active registry polling for the images (defaults to polling every minute)">
              <Switch
                onChange={() =>
                  setTracking.mutate({
                    identifier: resource.identifier,
                    provider: resource.provider,
                    trigger:
                      resource.labels["quilla.sh/trigger"] === "poll" ||
                      // @ts-ignore
                      resource.annotations["quilla.sh/trigger"] === "poll"
                        ? "default"
                        : "poll",
                  })
                }
                disabled={resource.policy === "nil policy"}
                checked={
                  resource.labels["quilla.sh/trigger"] === "poll" ||
                  // @ts-ignore
                  resource.annotations["quilla.sh/trigger"] === "poll"
                }
                checkedChildren={<SyncOutlined />}
                unCheckedChildren={<LinkOutlined />}
              />
            </Tooltip>
          </Space>
        </Flex>
      );
    },
  },
];

export default function DashboardTable() {
  const [filterText, setFilterText] = useState("");

  const queryClient = useQueryClient();
  const resources = useQuery<Resource[]>({
    queryKey: ["resources"],
    queryFn: getResources,
  });

  const policySet = useMutation({
    mutationFn: setPolicy,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["resources"] });
    },
  });

  const trackingSet = useMutation({
    mutationFn: setTracking,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["resources"] });
    },
  });

  const approvalSet = useMutation({
    mutationFn: setApproval,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["resources"] });
    },
  });

  return (
    <DataTable
      title="Kubernetes Cluster Resources"
      columns={columns(policySet, trackingSet, approvalSet)}
      dataSource={
        filterText === ""
          ? resources.data
          : resources.data?.filter(
              (resource) =>
                resource.identifier.includes(filterText) ||
                resource.namespace.includes(filterText) ||
                resource.policy.includes(filterText) ||
                resource.provider.includes(filterText) ||
                resource.images.some((image) => image.includes(filterText))
            )
      }
      onRefresh={() =>
        queryClient.invalidateQueries({ queryKey: ["resources"] })
      }
      onChange={(event) => setFilterText(event.currentTarget.value)}
      search
    />
  );
}
