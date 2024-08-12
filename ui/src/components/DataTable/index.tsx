import { Button, Card, Flex, Space, Table, Typography } from "antd";
import { AnyObject } from "antd/es/_util/type";
import Search from "antd/es/input/Search";
import { ColumnsType } from "antd/es/table";
import { TableRowSelection } from "antd/es/table/interface";

const Title = Typography.Title;

const TableTitle = (
  props: {
    title: string;
    onRefresh?: () => {};
    onChange?: React.ChangeEventHandler<HTMLInputElement>;
    search?: boolean;
    extraComponent?: React.ReactNode
  }
) => (
  <Flex justify="space-between">
    <Title level={3}>{props.title}</Title>
    <Flex align="center" vertical={false}>
      <Space>
        {props.search && (
          <>
            <Button ghost onClick={props.onRefresh} type="primary">
              Refresh
            </Button>
            {props.extraComponent}
            <Search onChange={props.onChange} />
          </>
        )}
      </Space>
    </Flex>
  </Flex>
);

export { TableTitle };

export default function DataTable<T extends AnyObject>(props: {
  title: string;
  onChange?: React.ChangeEventHandler<HTMLInputElement>;
  onRefresh?: () => {};
  columns?: ColumnsType<T>;
  dataSource?: AnyObject[];
  search?: boolean
  extraComponent?: React.ReactNode
  rowSelection?: TableRowSelection<AnyObject>
}) {
  return (
    <Card
      style={{ margin: 16, borderRadius: "0px" }}
      title={
        <TableTitle
          title={props.title}
          onChange={props.onChange}
          onRefresh={props.onRefresh}
          search={props.search}
          extraComponent={props.extraComponent}
        />
      }
    >
      <Table columns={props.columns} dataSource={props.dataSource} rowSelection={props.rowSelection} />
    </Card>
  );
}
