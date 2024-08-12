import { Flex, Typography } from "antd";

const Title = Typography.Title

export default function TextDataDisplay(props: { title: string; data: string }) {
    return (<Flex vertical justify="center" align="center">
    <Title style={{ margin: 0 }} disabled level={4}>
      {props.title}
    </Title>
    <Title style={{ margin: 0, paddingTop: 5 }} level={3}>
      {props.data}
    </Title>
  </Flex>)
}