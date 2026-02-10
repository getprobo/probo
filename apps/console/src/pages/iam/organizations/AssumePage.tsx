import { useTranslate } from "@probo/i18n";
import { useLocation, useNavigate } from "react-router";

import { useAssume } from "#/hooks/iam/useAssume";

import AuthLayout from "../auth/AuthLayout";

interface State {
  from: string;
}

export default function AssumePage() {
  const navigate = useNavigate();
  const location = useLocation();
  const state = location.state as State;

  const { __ } = useTranslate();

  useAssume({
    onSuccess: () => void navigate(state.from),
  });

  return (
    <AuthLayout>
      <div className="space-y-6 w-full max-w-md mx-auto pt-8">
        <div className="space-y-2 text-center">
          <h1 className="text-3xl font-bold">{__("Sign in Redirection")}</h1>
          <p className="text-txt-tertiary">
            {__("Redirecting you to your authentication URL…")}
          </p>
        </div>
      </div>
    </AuthLayout>
  );
}
