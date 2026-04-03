"use client";

import {
  createContext,
  useContext,
  useEffect,
  useState,
  ReactNode,
} from "react";
import { api } from "./api";

interface AuthContextType {
  isAuthenticated: boolean;
  isLoading: boolean;
  displayName: string;
  login: (email: string, password: string) => Promise<void>;
  register: (
    email: string,
    password: string,
    displayName: string
  ) => Promise<void>;
  logout: () => void;
}

const AuthContext = createContext<AuthContextType | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [displayName, setDisplayName] = useState("");

  useEffect(() => {
    const token = localStorage.getItem("access_token");
    const name = localStorage.getItem("display_name") || "";
    setIsAuthenticated(!!token);
    setDisplayName(name);
    setIsLoading(false);
  }, []);

  const login = async (email: string, password: string) => {
    await api.login(email, password);
    setIsAuthenticated(true);
  };

  const register = async (
    email: string,
    password: string,
    displayName: string
  ) => {
    await api.register(email, password, displayName);
    localStorage.setItem("display_name", displayName);
    setDisplayName(displayName);
    setIsAuthenticated(true);
  };

  const logout = () => {
    api.clearTokens();
    localStorage.removeItem("display_name");
    setIsAuthenticated(false);
    setDisplayName("");
  };

  return (
    <AuthContext.Provider
      value={{ isAuthenticated, isLoading, displayName, login, register, logout }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
}
