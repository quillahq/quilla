import { MsalProvider } from "@azure/msal-react";
import Loading from "../../components/Loading";
import {
  Configuration,
  LogLevel,
  PublicClientApplication,
} from "@azure/msal-browser";
import useConfig from "../../hooks/useConfig";
import { useNavigate } from "react-router-dom";
import { useEffect, useState } from "react";

export default function MSAL(props: { children: React.ReactNode }) {
  const { children } = props;
  const config = useConfig();
  const navigate = useNavigate();
  const [loading, setLoading] = useState(true);
  const [pca, setPca] = useState<PublicClientApplication | null>(null);

  useEffect(() => {
    if (config?.aad) {
      const configuration: Configuration = config.aad as Configuration;

      const pca = new PublicClientApplication({
        ...configuration,
        ...(config?.debug && {
          system: {
            loggerOptions: {
              loggerCallback: (level, message, containsPil) => {
                if (containsPil) {
                  return;
                }
                switch (level) {
                  case LogLevel.Error:
                    console.error(message);
                    return;
                  case LogLevel.Info:
                    console.info(message);
                    return;
                  case LogLevel.Verbose:
                    console.debug(message);
                    return;
                  case LogLevel.Warning:
                    console.warn(message);
                    return;
                }
              },
            },
          },
        }),
      });
      pca
        .initialize()
        .then(() => setLoading(false))
        .then(() =>
          pca.handleRedirectPromise().then((result) => {
            if (result && result.idToken) {
              localStorage.setItem("token", result.idToken);
            }

            if (
              localStorage.getItem("token") &&
              window.location.pathname === "/login"
            ) {
              console.log("back #2")
              navigate("/");
            }
          })
        );

      setPca(pca);
    }
  }, [config, config?.aad, config?.debug, navigate]);

  if (config?.aad) {
    if (loading) {
      return <Loading />;
    }

    if (pca) {
      return <MsalProvider instance={pca}>{children}</MsalProvider>;
    }
  }

  if (!config) {
    return <Loading />;
  }

  return <>{children}</>;
}
