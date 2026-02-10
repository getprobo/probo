import { useTranslate } from "@probo/i18n";
import { useNavigate, useSearchParams } from "react-router";

import { useAssume } from "#/hooks/iam/useAssume";
import { IAMRelayProvider } from "#/providers/IAMRelayProvider";

import AuthLayout from "../auth/AuthLayout";

function AssumePageInner() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const { __ } = useTranslate();

  useAssume({
    afterAssumePath: searchParams.get("redirect-path") ?? "/",
    onSuccess: () => void navigate(searchParams.get("redirect-path") ?? "/"),
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

export default function AssumePage() {
  return (
    <IAMRelayProvider>
      <AssumePageInner />
    </IAMRelayProvider>
  );
}
