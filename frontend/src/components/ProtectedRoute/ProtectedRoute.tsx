import { useEffect, useState } from "react";
import { Navigate, Outlet } from "react-router-dom";
import { useAuth } from "@/hooks/useAuth";

export const ProtectedRoute = () => {
  const { checkAuth } = useAuth();
  const [authState, setAuthState] = useState<"loading" | "authenticated" | "guest">(
    "loading",
  );

  useEffect(() => {
    const verify = async () => {
      const authInfo = await checkAuth();
      setAuthState(authInfo.type === "user" ? "authenticated" : "guest");
    };

    verify();
  }, [checkAuth]);

  if (authState === "loading") {
    return null;
  }

  return authState === "authenticated" ? <Outlet /> : <Navigate to="/login" />;
}; 