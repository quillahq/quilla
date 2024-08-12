import { TableProps, Tag } from "antd";
import { Audit as AuditType, getAudit } from "../../api/audit";
import DataTable from "../../components/DataTable";
import TripleDataDisplay from "../../components/TripleDataDisplay";
import { useQuery } from "@tanstack/react-query";

const columns: TableProps<AuditType>["columns"] = [
  {
    title: "Time",
    dataIndex: "createdAt",
    key: "createdAt",
    render: (_, audit) =>
      new Date(audit.createdAt).toLocaleDateString(undefined, {
        day: "numeric",
        month: "long",
        year: "numeric",
        hour: "2-digit",
        minute: "2-digit",
        second: "2-digit",
      }),
  },
  {
    title: "Action",
    dataIndex: "action",
    key: "action",
  },
  {
    title: "Resource Kind",
    dataIndex: "resourceKind",
    key: "resourceKind",
  },
  {
    title: "Identifier",
    dataIndex: "identifier",
    key: "identifier",
  },
  {
    title: "Metadata",
    dataIndex: "metadata",
    key: "metadata",
    width: 400,
    render: (_, audit) =>
      Object.keys(audit.metadata).map((key) => (
        <Tag>
          {key}: {audit.metadata[key]}
        </Tag>
      )),
  },
];

export default function Audit() {
  const audit = useQuery<AuditType[]>({
    queryKey: ["audit"],
    queryFn: getAudit,
  });

  return (
    <>
      <TripleDataDisplay
        firstTitle="Last Event"
        firstData={audit.data && audit.data.length > 1 ? new Date(audit.data[0].createdAt).toLocaleDateString(
          undefined,
          {
            day: "numeric",
            month: "long",
            year: "numeric",
            hour: "2-digit",
            minute: "2-digit",
            second: "2-digit",
          }
        ) : "None"}
        secondTitle="Audit Entries"
        secondData={audit.data?.length.toString() || "0"}
        thirdTitle="Registries"
        thirdData="-"
      />
      <DataTable title="Audit Logs" columns={columns} dataSource={audit.data} />
    </>
  );
}
