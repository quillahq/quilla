import { TableProps } from "antd";
import DataTable from "../../components/DataTable";
import { getTracking, Tracking } from "../../api/tracking";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import TripleDataDisplay from "../../components/TripleDataDisplay";

const columns: TableProps<Tracking>["columns"] = [
  {
    title: "Image Name",
    key: "image",
    dataIndex: "image",
  },
  {
    title: "Provider",
    key: "provider",
    dataIndex: "provider",
  },
  {
    title: "Namespace",
    key: "namespace",
    dataIndex: "namespace",
  },
  {
    title: "Policy",
    key: "policy",
    dataIndex: "policy",
  },
  {
    title: "Trigger",
    key: "trigger",
    dataIndex: "trigger",
  },
];

export default function TrackedImages() {
  const [filterText, setFilterText] = useState("");

  const queryClient = useQueryClient();
  const tracked = useQuery<Tracking[]>({
    queryKey: ["tracked"],
    queryFn: getTracking,
  });

  return (
    <>
      <TripleDataDisplay
        firstTitle="Namespaces"
        firstData={Array.from(
          new Set(tracked.data?.map((t) => t.namespace))
        ).length.toString()}
        secondTitle="Total Images Tracked"
        secondData={tracked.data?.length.toString() || "0"}
        thirdTitle="Registries"
        thirdData={Array.from(
          new Set(tracked.data?.map((t) => t.registry))
        ).length.toString()}
      />
      <DataTable
        title="Tracked Images"
        columns={columns}
        dataSource={
          filterText === ""
            ? tracked.data
            : tracked.data?.filter(
                (track) =>
                  track.image.includes(filterText) ||
                  track.namespace.includes(filterText) ||
                  track.policy.includes(filterText) ||
                  track.trigger.includes(filterText) ||
                  track.pollSchedule.includes(filterText) ||
                  track.registry.includes(filterText)
              )
        }
        onChange={(event) => setFilterText(event.currentTarget.value)}
        onRefresh={() =>
          queryClient.invalidateQueries({ queryKey: ["tracked"] })
        }
        search
      />
    </>
  );
}
