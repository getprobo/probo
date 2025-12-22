import { Logo } from "@probo/ui";
import { Outlet } from "react-router";
import { IAMRelayProvider } from "/providers/IAMRelayProvider";

export default function () {
  return (
    <div className="grid grid-cols-1 lg:grid-cols-2 min-h-screen text-txt-primary">
      <div className="bg-level-0 flex flex-col items-center justify-center">
        <div className="w-full max-w-md px-6">
          <IAMRelayProvider>
            <Outlet />
          </IAMRelayProvider>
        </div>
      </div>
      <div className="hidden lg:flex bg-dialog font-bold flex-col items-center justify-center p-8 text-txt-primary lg:p-10">
        <div className="flex flex-col items-center justify-center gap-4">
          <Logo withPicto className="w-[440px]" />
        </div>
      </div>
    </div>
  );
}
