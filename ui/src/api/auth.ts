export type LoginPayload = {
  username: string;
  password: string;
};

export const login = async ({ username, password }: LoginPayload) => {
  const response = await fetch("/v1/auth/login", {
    method: "POST",
    body: JSON.stringify({
      Username: username,
      Password: password,
    }),
  });
  const json = await response.json();
  return json.token;
};

export const getAuthToken = () => {
  const token = localStorage.getItem("token");
  return token;
};

export const setAuthToken = (token: string) => {
  localStorage.setItem("token", token);
};

export const clearAuthToken = () => {
  localStorage.removeItem("token");
};
