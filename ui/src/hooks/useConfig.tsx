import { useQuery } from "@tanstack/react-query";
import { Configuration } from "msal";
import { getConfig } from "../api/config";
type AppConfig = {
  aad: Configuration;
  debug: boolean;
  basicAuth: boolean;
};

const useConfig = () => {
  const configRequest = useQuery<AppConfig>({
    queryFn: getConfig,
    queryKey: ["config"],
  });

  return configRequest.data;
};

export default useConfig;
