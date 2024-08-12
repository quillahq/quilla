import { Card, Col, Row } from "antd";
import TextDataDisplay from "../TextDataDisplay";

export default function TripleDataDisplay(props: {
    firstTitle: string
    firstData: string
    secondTitle: string
    secondData: string
    thirdTitle: string
    thirdData: string
}) {
    return (<Card style={{ margin: 16, borderRadius: "0px" }}>
        <Row>
          <Col
            span={8}
            style={{ border: "0px solid #303030", borderRightWidth: "1px" }}
          >
            <TextDataDisplay
              title={props.firstTitle}
              data={props.firstData}
            />
          </Col>
          <Col
            span={8}
            style={{ border: "0px solid #303030", borderRightWidth: "1px" }}
          >
            <TextDataDisplay
              title={props.secondTitle}
              data={props.secondData}
            />
          </Col>
          <Col span={8}>
            <TextDataDisplay
              title={props.thirdTitle}
              data={props.thirdData}
            />
          </Col>
        </Row>
      </Card>)
}