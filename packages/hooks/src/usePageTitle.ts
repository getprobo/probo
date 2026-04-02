import { useEffect } from "react";

export function usePageTitle(title: string) {
    useEffect(() => {
        document.title = title + " - Govrly";
        return () => {
            document.title = "Govrly";
        };
    }, [title]);
}
