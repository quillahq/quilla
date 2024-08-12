import { LockOutlined, UserOutlined } from "@ant-design/icons";
import { Flex, Input, Form, Card, Button } from "antd";
import MSLogo from "../../mslogo.png";
import { useMsal } from "@azure/msal-react";
import { InteractionStatus } from "@azure/msal-browser";
import useConfig from "../../hooks/useConfig";
import { useMutation } from "@tanstack/react-query";
import { login, setAuthToken } from "../../api/auth";
import { useNavigate } from "react-router-dom";

export default function SignIn() {
  const config = useConfig();
  const { instance, inProgress } = useMsal();
  const navigate = useNavigate();

  const loginBasic = useMutation({
    mutationFn: login,
    onSuccess: (token) => {
      setAuthToken(token);
      console.log("back #3")
      navigate("/");
    },
  });
  const onFinishAAD = () => {
    if (inProgress === InteractionStatus.None) {
      instance.loginRedirect({
        redirectUri:
          typeof config?.aad.auth.redirectUri === "string"
            ? config?.aad.auth.redirectUri || ""
            : "",
        scopes: [],
      });
    }
  };

  const onFinishBasic = (values: { username: string; password: string }) => {
    loginBasic.mutate({
      username: values.username,
      password: values.password,
    });
  };

  return (
    <Flex
      style={{ width: "100%", height: "100%" }}
      vertical
      align="center"
      justify="center"
    >
      <Card title="Sign In">
        {config?.basicAuth && (
          <Form name="auth" onFinish={onFinishBasic}>
            <Form.Item
              name="username"
              rules={[{ required: true, message: "Input your username!" }]}
            >
              <Input prefix={<UserOutlined />} placeholder="Username" />
            </Form.Item>
            <Form.Item
              name="password"
              rules={[{ required: true, message: "Input your password!" }]}
            >
              <Input
                prefix={<LockOutlined />}
                type="password"
                placeholder="Password"
              />
            </Form.Item>
            <Form.Item>
              <Button block type="primary" htmlType="submit">
                Log in
              </Button>
            </Form.Item>
          </Form>
        )}

        {config?.aad && (
          <Form name="auth" onFinish={onFinishAAD}>
            <Form.Item>
              <Button
                block
                type="primary"
                htmlType="submit"
                icon={<img alt="Microsoft Logo" src={MSLogo} />}
              >
                Continue with Azure AD
              </Button>
            </Form.Item>
          </Form>
        )}
      </Card>
    </Flex>
  );
}
