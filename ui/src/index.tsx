import React from "react";
import ReactDOM from "react-dom/client";
import "./index.css";
import App from "./App";
import reportWebVitals from "./reportWebVitals";
import { ConfigProvider, theme } from "antd";
import { BrowserRouter, Route, Routes } from "react-router-dom";
import SignIn from "./pages/SignIn";
import AuthRoute from "./components/AuthRoute";
import JWT from "./hoc/JWT";
import MSAL from "./hoc/MSAL";
import Dashboard from "./pages/Dashboard";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import TrackedImages from "./pages/TrackedImages";
import Audit from "./pages/Audit";
import Approval from "./pages/Approvals";

const root = ReactDOM.createRoot(
  document.getElementById("root") as HTMLElement
);

const queryClient = new QueryClient();

const RouteHelper = (props: { children: React.ReactNode }) => (
  <AuthRoute>
    <App>{props.children}</App>
  </AuthRoute>
);

const error = console.error;
console.error = (...args: any) => {
  if (/defaultProps/.test(args[0])) return;
  error(...args);
};

root.render(
  <React.StrictMode>
    <ConfigProvider
      theme={{
        algorithm: theme.darkAlgorithm,
      }}
    >
      <QueryClientProvider client={queryClient}>
        <BrowserRouter>
          <MSAL>
            <JWT>
              <Routes>
                <Route path="/login" element={<SignIn />} />
                <Route
                  path="/"
                  element={
                    <RouteHelper>
                      <Dashboard />
                    </RouteHelper>
                  }
                />
                <Route
                  path="/tracked-images"
                  element={
                    <RouteHelper>
                      <TrackedImages />
                    </RouteHelper>
                  }
                />
                <Route
                  path="/audit-logs"
                  element={
                    <RouteHelper>
                      <Audit />
                    </RouteHelper>
                  }
                />
                <Route
                  path="/approvals"
                  element={
                    <RouteHelper>
                      <Approval />
                    </RouteHelper>
                  }
                />
              </Routes>
            </JWT>
          </MSAL>
        </BrowserRouter>
      </QueryClientProvider>
    </ConfigProvider>
  </React.StrictMode>
);

// If you want to start measuring performance in your app, pass a function
// to log results (for example: reportWebVitals(console.log))
// or send to an analytics endpoint. Learn more: https://bit.ly/CRA-vitals
reportWebVitals();
