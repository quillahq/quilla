import { EllipsisOutlined, QuestionCircleOutlined } from "@ant-design/icons";
import { Button, Dropdown, Menu, Space } from "antd";
import { Header as HeaderLayout } from "antd/es/layout/layout";
import { useNavigate } from "react-router-dom";

const navMap = ["/", "/tracked-images", "/approvals", "/audit-logs"];

export default function Header() {
  const navigate = useNavigate();
  const defaultPathIndex = navMap
    .findIndex((path) => window.location.pathname === path)
    .toString();

  return (
    <HeaderLayout
      style={{
        display: "flex",
        alignItems: "center",
        backgroundColor: "#141414",
        borderBottom: "1px solid rgba(253, 253, 253, 0.12)",
      }}
    >
      <Menu
        mode="horizontal"
        style={{ flex: 1, minWidth: 0 }}
        onClick={(info) => navigate(navMap[Number(info.key)])}
        defaultSelectedKeys={[
          defaultPathIndex === "-1" ? "0" : defaultPathIndex,
        ]}
        items={[
          {
            key: 0,
            label: "Dashboard",
          },
          {
            key: 1,
            label: "Tracked Images",
          },
          {
            key: 2,
            label: "Approvals",
          },
          {
            key: 3,
            label: "Audits",
          },
        ]}
      />
      <Space>
        <Button icon={<QuestionCircleOutlined />} type="text" />
        <Dropdown
          placement="topRight"
          menu={{
            items: [
              {
                key: "1",
                label: "Logout",
              },
            ],
            onClick: (info) => {
              localStorage.clear();
              navigate("/login");
            },
          }}
        >
          <Button icon={<EllipsisOutlined />} type="text" />
        </Dropdown>
      </Space>
    </HeaderLayout>
  );
}
