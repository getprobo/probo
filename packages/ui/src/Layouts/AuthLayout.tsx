import logo from "../assets/android-chrome-512x512.png";
import { useTranslate } from "@probo/i18n";
import { Outlet } from "react-router";

export function AuthLayout() {
    const { __ } = useTranslate();
    return (
        <div className="grid grid-cols-1 lg:grid-cols-2 min-h-screen text-txt-primary">
            <div className="bg-level-0 flex flex-col items-center justify-center">
                <div className="max-w-112">
                    <Outlet />
                </div>
            </div>
            <div className="hidden lg:flex bg-dialog text-invert text-5xl font-bold flex flex-col items-center justify-center p-8 text-txt-primary lg:p-10">
                <div className="flex flex-col 2xl:flex-row-reverse items-center justify-center gap-4">
                    <img src={logo} alt="Probo logo" className="h-auto w-96" />
                    <span>
                        {__("Navigate compliance with confidence thanks to")}
                        <span className="text-txt-accent"> probo</span>
                    </span>
                </div>
            </div>
        </div>
    );
}
