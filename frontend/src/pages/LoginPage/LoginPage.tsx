import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";

import { Button } from "@/components/Button/Button";
import { useAuth } from "@/hooks/useAuth";
import { Icon } from "@/components/Icon/Icon";

import {
  buttonStyles,
  errorMessageStyles,
  formStyles,
  inputGroupStyles,
  inputStyles,
  labelStyles,
  titleStyles,
  wrapperStyles,
  dividerStyles,
  dividerLineStyles,
  dividerTextStyles,
  socialButtonsStyles,
  socialButtonStyles,
} from "./LoginPage.css";

const isValidEmail = (value: string): boolean => {
  if (value.toLowerCase() === 'admin') return true;
  
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  return emailRegex.test(value);
};

type Provider = "google" | "github";

export const LoginPage = () => {
  const navigate = useNavigate();
  const { login, loginWithGoogle, loginWithGithub } = useAuth();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [providers, setProviders] = useState<Provider[]>([]);

  useEffect(() => {
    const fetchProviders = async () => {
      try {
        const response = await fetch("/api/v1/info");
        const data = await response.json();
        setProviders(data.data.providers || []);
      } catch (err) {
        console.error("Failed to fetch providers:", err);
        setProviders([]);
      }
    };

    fetchProviders();
  }, []);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    if (!isValidEmail(email)) {
      setError("Please enter a valid email address or 'admin'");
      return;
    }

    setIsLoading(true);

    try {
      await login(email, password);
      navigate("/chat");
    } catch (err) {
      setError(
        err instanceof Error ? err.message : "An error occurred during login",
      );
    } finally {
      setIsLoading(false);
    }
  };

  const handleEmailChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setEmail(e.target.value);
    if (error && isValidEmail(e.target.value)) {
      setError(null);
    }
  };

  const handleGoogleLogin = async () => {
    setError(null);
    setIsLoading(true);
    try {
      await loginWithGoogle();
      navigate("/chat");
    } catch (err) {
      setError(
        err instanceof Error ? err.message : "An error occurred during Google login"
      );
    } finally {
      setIsLoading(false);
    }
  };

  const handleGithubLogin = async () => {
    setError(null);
    setIsLoading(true);
    try {
      await loginWithGithub();
      navigate("/chat");
    } catch (err) {
      setError(
        err instanceof Error ? err.message : "An error occurred during Github login"
      );
    } finally {
      setIsLoading(false);
    }
  };

  const getProviderButton = (provider: Provider) => {
    switch (provider) {
      case "google":
        return (
          <button
            key="google"
            type="button"
            className={socialButtonStyles}
            onClick={handleGoogleLogin}
            disabled={isLoading}
          >
            <Icon.Google />
            Continue with Google
          </button>
        );
      case "github":
        return (
          <button
            key="github"
            type="button"
            className={socialButtonStyles}
            onClick={handleGithubLogin}
            disabled={isLoading}
          >
            <Icon.Github />
            Continue with Github
          </button>
        );
      default:
        return null;
    }
  };

  return (
    <div className={wrapperStyles}>
      <form className={formStyles} onSubmit={handleSubmit}>
        <h1 className={titleStyles}>PentAGI</h1>

        {providers.length > 0 && (
          <>
            <div className={socialButtonsStyles}>
              {providers?.sort((a, b) => b.localeCompare(a)).map(provider => getProviderButton(provider))}
            </div>

            <div className={dividerStyles}>
              <div className={dividerLineStyles} />
              <span className={dividerTextStyles}>or</span>
              <div className={dividerLineStyles} />
            </div>
          </>
        )}

        <div className={inputGroupStyles}>
          <label className={labelStyles} htmlFor="email">
            Email
          </label>
          <input
            id="email"
            type="text"
            className={inputStyles}
            value={email}
            onChange={handleEmailChange}
            required
            placeholder="Enter your email or 'admin'"
            disabled={isLoading}
          />
        </div>

        <div className={inputGroupStyles}>
          <label className={labelStyles} htmlFor="password">
            Password
          </label>
          <input
            id="password"
            type="password"
            className={inputStyles}
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
            placeholder="Enter your password"
            disabled={isLoading}
          />
        </div>

        {error && <div className={errorMessageStyles}>{error}</div>}

        <div className={buttonStyles}>
          <Button type="submit" disabled={isLoading}>
            {isLoading ? "Signing in..." : "Sign in"}
          </Button>
        </div>
      </form>
    </div>
  );
}; 