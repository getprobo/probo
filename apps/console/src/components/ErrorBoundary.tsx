import { Component, type ErrorInfo, type ReactNode } from "react";
import { UnAuthenticatedError } from "@/RelayEnvironment";
import { Navigate } from "react-router";
import {ErrorPage} from "../pages/ErrorPage.tsx";

interface ErrorBoundaryProps {
  children: ReactNode;
  fallback?: ReactNode;
}

interface ErrorBoundaryState {
  hasError: boolean;
  error?: Error;
  isUnAuthenticated: boolean;
}

class ErrorBoundary extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
  constructor(props: ErrorBoundaryProps) {
    super(props);
    this.state = { hasError: false, isUnAuthenticated: false };
  }

  static getDerivedStateFromError(error: Error): ErrorBoundaryState {
    if (error instanceof UnAuthenticatedError) {
      return { hasError: true, error, isUnAuthenticated: true };
    }
    return { hasError: true, error, isUnAuthenticated: false };
  }

  componentDidCatch(error: Error, info: ErrorInfo): void {
    console.error("ErrorBoundary caught an error:", error);
    console.error("Error info:", info.componentStack);
  }

  render(): ReactNode {
    if (this.state.isUnAuthenticated) {
      return <Navigate to="/login" replace state={{ authRequired: true }} />;
    }

    if (this.state.hasError) {
      if (this.props.fallback) {
        return this.props.fallback;
      }
      return <ErrorPage error={this.state.error} />;
    }

    return this.props.children;
  }
}

export default ErrorBoundary;
