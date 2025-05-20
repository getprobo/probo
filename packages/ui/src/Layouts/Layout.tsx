import type { PropsWithChildren, ReactNode } from "react";
import { Sidebar } from "../Atoms/Sidebar/Sidebar.tsx";
import { Logo } from "../Atoms/Logo/Logo.tsx";
import { Toasts } from "../Atoms/Toasts/Toasts.tsx";

type Props = PropsWithChildren<{
    header: ReactNode;
    sidebar: ReactNode;
}>;

export function Layout({ header, sidebar, children }: Props) {
    return (
        <div>
            <header className="absolute z-2 left-0 right-0 px-4 flex items-center border-b border-border-solid h-12 bg-level-1">
                <Logo className="w-12 h-5" />
                <svg
                    className="mx-3"
                    width="8"
                    height="18"
                    viewBox="0 0 8 18"
                    fill="none"
                    xmlns="http://www.w3.org/2000/svg"
                >
                    <path d="M1 17L7 1" stroke="#C3C8C2" />
                </svg>
                {header}
            </header>
            <div className="flex h-screen">
                <Sidebar>{sidebar}</Sidebar>
                <main className="p-12 pt-24 max-w-[1200px] w-full mx-auto overflow-y-auto">
                    {children}
                </main>
            </div>
            <Toasts />
        </div>
    );
}
