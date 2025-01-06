import { useCallback, useEffect } from "react";
import { useNavigate } from "react-router-dom";

const baseURL = "/api/v1";

interface AuthInfo {
  type: "guest" | "user";
  user?: {
    id: number;
    mail: string;
    name: string;
  };
}

export const useAuth = () => {
  const navigate = useNavigate();

  const getUrlWS = useCallback(() => {
    return `${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}${baseURL}`;
  }, []);

  const checkAuth = useCallback(async () => {
    try {
      const response = await fetch(`${baseURL}/info`);
      if (!response.ok) throw new Error("Failed to check auth status");

      const data = await response.json();
      return data.data as AuthInfo;
    } catch (error) {
      console.error("Auth check failed:", error);
      return { type: "guest" } as AuthInfo;
    }
  }, []);

  const login = useCallback(async (email: string, password: string) => {
    const response = await fetch(`${baseURL}/auth/login`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        mail: email,
        password: password,
      }),
    });

    if (!response.ok) {
      if (response.status === 401) {
        throw new Error("Invalid email or password");
      }
      throw new Error("Login failed");
    }

    const authInfo = await checkAuth();
    if (authInfo.type === "guest") {
      throw new Error("Login failed");
    }
  }, [checkAuth]);

  const handleOAuthPopupResponse = useCallback(async (popup: Window, checkInterval: ReturnType<typeof setInterval>) => {
    try {
      const currentUrl = popup.location.href;
      const url = new URL(currentUrl);
      
      if (url.pathname.includes('/auth/login-callback')) {
        clearInterval(checkInterval);
        popup.close();

        try {
          const responseText = popup.document.body.textContent || '';
          console.log(responseText);
          const data = JSON.parse(responseText);
          
          if (data.status === 'success') {
            return { success: true };
          }
          
          return { 
            success: false, 
            error: data.error?.message || 'Authentication failed'
          };
        } catch (error) {
          return { 
            success: false, 
            error: 'Failed to process authentication response'
          };
        }
      }
      
      return null; // Continue waiting for response
    } catch (error) {
      // If there's an error accessing location (window on different domain)
      return null; // Continue waiting for response
    }
  }, []);

  const loginWithGoogle = useCallback(async () => {
    const width = 500;
    const height = 600;
    const left = window.screenX + (window.outerWidth - width) / 2;
    const top = window.screenY + (window.outerHeight - height) / 2;
    
    const popup = window.open(
      `${baseURL}/auth/authorize?provider=google`,
      'Google Sign In',
      `width=${width},height=${height},left=${left},top=${top}`
    );

    if (!popup) {
      throw new Error("Popup blocked. Please allow popups for this site.");
    }

    return new Promise<void>((resolve, reject) => {
      const checkInterval = setInterval(async () => {
        try {
          if (popup.closed) {
            clearInterval(checkInterval);
            reject(new Error("Authentication cancelled by user"));
            return;
          }

          const result = await handleOAuthPopupResponse(popup, checkInterval);
          if (result !== null) {
            if (result.success) {
              navigate('/chat/new');
              resolve();
            } else {
              reject(new Error(result.error));
            }
          }
        } catch (error) {
          clearInterval(checkInterval);
          reject(error);
        }
      }, 500);

      setTimeout(() => {
        clearInterval(checkInterval);
        popup.close();
        reject(new Error("Authentication timeout"));
      }, 300000);
    });
  }, [handleOAuthPopupResponse, navigate]);

  const loginWithGithub = useCallback(async () => {
    const width = 500;
    const height = 600;
    const left = window.screenX + (window.outerWidth - width) / 2;
    const top = window.screenY + (window.outerHeight - height) / 2;
    
    const popup = window.open(
      `${baseURL}/auth/authorize?provider=github`,
      'Github Sign In',
      `width=${width},height=${height},left=${left},top=${top}`
    );

    if (!popup) {
      throw new Error("Popup blocked. Please allow popups for this site.");
    }

    return new Promise<void>((resolve, reject) => {
      const checkInterval = setInterval(async () => {
        try {
          // If popup closed
          if (popup.closed) {
            clearInterval(checkInterval);
            const authInfo = await checkAuth();
            if (authInfo.type === "user") {
              resolve();
            } else {
              reject(new Error("Authentication failed"));
            }
            return;
          }

          // Check authentication result
          const success = await handleOAuthPopupResponse(popup, checkInterval);
          if (success) {
            resolve();
          }
        } catch (error) {
          clearInterval(checkInterval);
          popup.close();
          reject(error);
        }
      }, 500);

      setTimeout(() => {
        clearInterval(checkInterval);
        popup.close();
        reject(new Error("Authentication timeout"));
      }, 300000);
    });
  }, [checkAuth, handleOAuthPopupResponse]);

  useEffect(() => {
    const handleGraphQLError = async (response: Response) => {
      if (response.status === 401) {
        navigate("/login");
      }
    };

    const originalFetch = window.fetch;
    window.fetch = async (...args) => {
      const response = await originalFetch(...args);
      
      if (args[0] === `${baseURL}/graphql`) {
        await handleGraphQLError(response.clone());
      }
      
      return response;
    };

    return () => {
      window.fetch = originalFetch;
    };
  }, [navigate]);

  return { 
    login, 
    checkAuth, 
    getUrlWS,
    loginWithGoogle,
    loginWithGithub
  };
}; 