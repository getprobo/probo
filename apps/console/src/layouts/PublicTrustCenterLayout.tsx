import { Outlet } from "react-router";
import { Logo } from "@probo/ui";
import type { ReactNode } from "react";

type Props = {
  organizationName: string;
  organizationLogo?: string | null;
  children?: ReactNode;
};

export function PublicTrustCenterLayout({ organizationName, organizationLogo, children }: Props) {
  return (
    <div className="min-h-screen bg-tertiary">
      <header className="bg-surface border-b border-border-solid">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between h-16">
            <div className="flex items-center space-x-4">
              {organizationLogo ? (
                <img
                  src={organizationLogo}
                  alt={organizationName}
                  className="h-8 w-8 rounded"
                />
              ) : (
                <Logo className="h-8 w-8" />
              )}
              <div>
                <h1 className="text-lg font-semibold text-txt-primary">
                  {organizationName}
                </h1>
                <p className="text-sm text-txt-secondary">Trust Center</p>
              </div>
            </div>
            <div className="text-sm text-txt-tertiary">
              <a
                href="https://www.getprobo.com/"
                target="_blank"
                rel="noopener noreferrer"
                className="hover:text-txt-secondary transition-colors flex items-center space-x-1"
              >
                <img
                  src="/favicons/favicon.ico"
                  alt="Probo"
                  className="h-4 w-4"
                />
                <span>Powered by Probo</span>
              </a>
            </div>
          </div>
        </div>
      </header>

      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {children || <Outlet />}
      </main>
    </div>
  );
}
