import { Button, Popconfirm, Progress, Space, TableProps, Tooltip } from "antd";
import DataTable from "../../components/DataTable";
import TripleDataDisplay from "../../components/TripleDataDisplay";
import {
  approvalIsComplete,
  Approval as ApprovalType,
  ApprovalVotePayload,
  getApprovals,
  voteApproval,
} from "../../api/approvals";
import {
  useMutation,
  UseMutationResult,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";
import {
  DatabaseOutlined,
  DeleteOutlined,
  DislikeOutlined,
  LikeOutlined,
} from "@ant-design/icons";
import CountDown from "../../components/CountDown";
import { AnyObject } from "antd/es/_util/type";

const getProgress = (approval: ApprovalType) => {
  if (approval.votesReceived === 0) return 0;
  return (approval.votesReceived * 100) / approval.votesRequired;
};

const columns: (
  approvalVote: UseMutationResult<any, Error, ApprovalVotePayload, unknown>
) => TableProps<ApprovalType>["columns"] = (
  approvalVote: UseMutationResult<any, Error, ApprovalVotePayload, unknown>
) => [
  {
    title: "Last Activity",
    dataIndex: "updatedAt",
    key: "updatedAt",
  },
  {
    title: "Provider",
    dataIndex: "provider",
    key: "provider",
  },
  {
    title: "Identifier",
    dataIndex: "identifier",
    key: "identifier",
  },
  {
    title: "Votes",
    render: (_, approval) =>
      `${approval.votesReceived} / ${approval.votesRequired}`,
  },
  {
    title: "Delta",
    render: (_, approval) =>
      `${approval.currentVersion} -> ${approval.newVersion}`,
  },
  {
    title: "Status",
    render: (_, approval) => {
      var text = "";
      if (approval.rejected) {
        text = "Rejected";
      } else if (approval.votesReceived === approval.votesRequired) {
        text = "Complete";
      } else if (approval.archived) {
        text = "Archived";
      } else {
        text = "Collecting...";
      }

      return (
        <>
          {text}
          <Progress
            percent={
              approval.votesReceived < approval.votesRequired &&
              !approval.rejected
                ? getProgress(approval)
                : 100
            }
            {...((!approval.archived ||
              approval.votesReceived === approval.votesRequired) && {
              status: approval.rejected ? "exception" : "success",
            } || {
              status: "active"
            })}
          />
        </>
      );
    },
  },
  {
    title: "Expires In",
    dataIndex: "deadline",
    key: "deadline",
    render: (_, approval) =>
      approvalIsComplete(approval) ? (
        "-"
      ) : (
        <CountDown date={new Date(approval.deadline)} />
      ),
  },
  {
    title: "Action",
    render: (_, approval) => (
      <Space>
        <Popconfirm
          title="Approve"
          description="Are you sure you want approve this release?"
          onConfirm={() =>
            approvalVote.mutate({
              id: approval.id,
              identifier: approval.identifier,
              action: "approve",
            })
          }
        >
          <Button type="primary">
            <LikeOutlined />
          </Button>
        </Popconfirm>
        <Popconfirm
          title="Reject"
          description="Are you sure you want reject this release?"
          onConfirm={() =>
            approvalVote.mutate({
              id: approval.id,
              identifier: approval.identifier,
              action: "reject",
            })
          }
        >
          <Button danger>
            <DislikeOutlined />
          </Button>
        </Popconfirm>
        <Tooltip title="Archive approval">
          <Popconfirm
            title="Archive"
            description="Are you sure you want archive this release?"
            onConfirm={() =>
              approvalVote.mutate({
                id: approval.id,
                identifier: approval.identifier,
                action: "archive",
              })
            }
          >
            <Button type="primary">
              <DatabaseOutlined />
            </Button>
          </Popconfirm>
        </Tooltip>

        <Tooltip title="Delete approval request">
          <Popconfirm
            title="Delete"
            description="Are you sure you want delete this release?"
            onConfirm={() =>
              approvalVote.mutate({
                id: approval.id,
                identifier: approval.identifier,
                action: "delete",
              })
            }
          >
            <Button type="primary">
              <DeleteOutlined />
            </Button>
          </Popconfirm>
        </Tooltip>
      </Space>
    ),
  },
];

export default function Approval() {
  const queryClient = useQueryClient();
  const approval = useQuery<ApprovalType[]>({
    queryKey: ["approval"],
    queryFn: getApprovals,
  });

  const approvalVote = useMutation({
    mutationFn: voteApproval,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["approval"] });
    },
  });

  return (
    <>
      <TripleDataDisplay
        firstTitle="Pending"
        firstData={
          approval.data
            ?.filter(
              ({ rejected, archived, votesReceived, votesRequired }) =>
                !rejected && !archived && votesReceived < votesRequired
            )
            .length.toString() || "0"
        }
        secondTitle="Approved"
        secondData={
          approval.data
            ?.filter(
              ({ rejected, votesReceived, votesRequired }) =>
                !rejected && votesReceived >= votesRequired
            )
            .length.toString() || "0"
        }
        thirdTitle="Rejected"
        thirdData={
          approval.data?.filter(({ rejected }) => rejected).length.toString() ||
          "0"
        }
      />
      <DataTable
        search
        title="Approvals"
        columns={columns(approvalVote)}
        dataSource={approval.data}
        extraComponent={
          <>
            <Button
              type="primary"
              ghost
              style={{ color: "green", borderColor: "green" }}
            >
              Approve
            </Button>
            <Button danger>Reject</Button>
          </>
        }
        rowSelection={{
          type: "checkbox",
          onChange: (
            selectedRowKeys: React.Key[],
            selectedRows: AnyObject[]
          ) => {},
        }}
      />
    </>
  );
}
