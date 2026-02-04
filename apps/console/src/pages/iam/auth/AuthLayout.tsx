import { Card, Logo } from "@probo/ui";
import { Outlet } from "react-router";

import { IAMRelayProvider } from "#/providers/IAMRelayProvider";

export default function AuthLayout() {
  return (
    <div className="min-h-screen text-txt-primary bg-level-0 flex flex-col items-center justify-center">
      <Card className="w-full max-w-lg px-12 py-8 flex flex-col items-center justify-center">
        <div className="w-full flex flex-col items-center justify-center gap-8">
          <Logo withPicto className="w-[110px]" />
          <div className="w-full border-t border-t-border-mid" />
        </div>
        <IAMRelayProvider>
          <Outlet />
        </IAMRelayProvider>
      </Card>
    </div>
  );
}
