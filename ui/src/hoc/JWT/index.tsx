import { useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { clearAuthToken, getAuthToken } from "../../api/auth";

function decodeJwtToken(token: string) {
  const base64Url = token.split(".")[1];
  const base64 = base64Url.replace(/-/g, "+").replace(/_/g, "/");
  const jsonPayload = decodeURIComponent(
    window
      .atob(base64)
      .split("")
      .map(function (c) {
        return "%" + ("00" + c.charCodeAt(0).toString(16)).slice(-2);
      })
      .join("")
  );

  return JSON.parse(jsonPayload);
}

export function verifyJWT(): boolean {
  const token = getAuthToken();
  if (!token) {
    clearAuthToken();
    return false;
  }

  const userData = decodeJwtToken(token);
  const expirationDate = new Date(userData.exp * 1000);
  const currentDate = new Date();
  if (expirationDate < currentDate) {
    clearAuthToken();
    return false;
  }

  return true;
}

export default function JWT(props: { children: React.ReactNode }) {
  const { children } = props;
  const navigate = useNavigate();

  useEffect(() => {
    const valid = verifyJWT();
    if (!valid) {
      navigate("/login");
      return;
    }

    if (window.location.pathname === "/login") {
      console.log("back #1");
      navigate("/");
    }
  }, [navigate]);

  return <>{children}</>;
}
