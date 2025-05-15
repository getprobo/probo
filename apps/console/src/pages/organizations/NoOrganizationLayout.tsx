import { NavLink, Outlet, Route, Routes, type To } from "react-router";
import { AppSidebar } from "@/components/AppSidebar";
import { type ReactNode, Suspense } from "react";
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";
import { SidebarInset, SidebarProvider } from "@/components/ui/sidebar";
import { Toaster } from "@/components/ui/toaster";
import { cn } from "@/lib/utils";
import ErrorBoundary from "@/components/ErrorBoundary";

const BreadCrumbNavLink = ({
  to,
  children,
  className,
}: {
  to: To;
  children?: ReactNode;
  className?: string;
}) => {
  return (
    <BreadcrumbLink asChild>
      <NavLink to={to} end>
        {({ isActive }) => (
          <BreadcrumbPage
            className={cn(
              "text-quaternary",
              isActive && "text-primary",
              className,
            )}
          >
            {children}
          </BreadcrumbPage>
        )}
      </NavLink>
    </BreadcrumbLink>
  );
};

function BreadcrumbHome() {
  return (
    <Breadcrumb>
      <BreadcrumbList>
        <BreadcrumbItem>
          <BreadCrumbNavLink to="/">Home</BreadCrumbNavLink>
        </BreadcrumbItem>
        <Outlet />
      </BreadcrumbList>
    </Breadcrumb>
  );
}

export default function NoOrganizationLayout() {
  return (
    <SidebarProvider>
      <ErrorBoundary>
        <Suspense>
          <AppSidebar className="p-4" />
        </Suspense>
      </ErrorBoundary>
      <SidebarInset>
        <div className="mx-auto w-lg md:w-xl lg:w-3xl xl:w-4xl 2xl:w-5xl p-8">
          <header className="flex shrink-0 items-center gap-2">
            <Routes>
              <Route path="*" element={<BreadcrumbHome />}>
                <Route
                  path="new"
                  element={
                    <>
                      <BreadcrumbSeparator />
                      <BreadcrumbItem>
                        <BreadcrumbPage>New Organization</BreadcrumbPage>
                      </BreadcrumbItem>
                    </>
                  }
                />
              </Route>
            </Routes>
          </header>
          <div className="mt-7 w-full">
            <Outlet />
          </div>
        </div>
        <Toaster />
      </SidebarInset>
    </SidebarProvider>
  );
}
