import { Flex, Spin } from "antd";

export default function Loading() {
  return (
    <Flex
      align="center"
      justify="center"
      style={{ display: "flex", height: "100%", width: "100%" }}
    >
      <Spin />
    </Flex>
  );
}
