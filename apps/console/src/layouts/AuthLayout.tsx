import { Toaster } from "@/components/ui/toaster";
import type {PropsWithChildren} from "react";

export default function AuthLayout({ children }: PropsWithChildren) {
  return (
    <div className="flex min-h-screen flex-col">
      <div className="flex flex-1">
        <div className="flex flex-1 flex-col bg-subtle-bg/40">
          <div className="flex flex-1 items-center justify-center">
            <main className="w-full max-w-md p-6">{children}</main>
          </div>
        </div>
        <div className="hidden flex-1 bg-linear-to-br from-gray-900 to-gray-950 lg:block">
          <div className="flex h-full items-center justify-center p-8">
            <div className="relative h-full w-full">
              <div className="absolute right-0 top-1/4 z-10">
                <img
                  src="/assets/android-chrome-512x512.png"
                  alt="Probo Mascot"
                  className="h-auto w-96"
                />
              </div>
              <div className="absolute left-8 top-1/3 z-0 text-white">
                <h1 className="text-4xl font-bold leading-tight">
                  Navigate compliance with
                  <br />
                  confidence thanks to{" "}
                  <span className="text-accent">probo</span>
                </h1>
              </div>
            </div>
          </div>
        </div>
      </div>
      <Toaster />
    </div>
  );
}
