import axios from "axios";
import { toast } from "sonner";

export const api = axios.create({
  baseURL: "/v1",
  headers: {
    "Content-Type": "application/json",
  },
});

export function setAuthToken(token: string | null) {
  if (token) {
    localStorage.setItem("fleet_console_token", token);
  } else {
    localStorage.removeItem("fleet_console_token");
  }
}

export function clearAuthToken() {
  localStorage.removeItem("fleet_console_token");
}

export function getAuthToken(): string | null {
  return localStorage.getItem("fleet_console_token");
}

export function setProjectID(projectID: string | null) {
  if (projectID) {
    localStorage.setItem("fleet_console_project", projectID);
  } else {
    localStorage.removeItem("fleet_console_project");
  }
}

export function getProjectID(): string | null {
  return localStorage.getItem("fleet_console_project");
}

api.interceptors.request.use((config) => {
  const token = getAuthToken();
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  const projectID = getProjectID();
  if (projectID) {
    config.headers["X-Fleet-Project"] = projectID;
  }
  return config;
});

let authRedirecting = false;

api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (axios.isCancel(error)) {
      return Promise.reject(error);
    }
    const status = error?.response?.status;
    const message =
      error?.response?.data?.error?.message ||
      error?.message ||
      "Request failed";

    if (status === 401) {
      const isLoginRequest = error?.config?.url === "/console/auth/sign-in";
      if (isLoginRequest) {
        toast.error(message);
      } else if (!authRedirecting) {
        authRedirecting = true;
        toast.error("Session expired. Please sign in again.");
        clearAuthToken();
        setProjectID(null);
        window.location.href = "/console/login";
      }
    } else if (status >= 400) {
      toast.error(message);
    }
    return Promise.reject(error);
  }
);
