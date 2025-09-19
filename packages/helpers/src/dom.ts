export function withViewTransition(fn: () => void) {
    if (!document.startViewTransition) {
        fn();
        return;
    }
    document.startViewTransition(fn);
}

export function downloadFile(url: string | undefined | null, filename: string) {
    if (!url) {
        alert("Cannot download this file, fileUrl is not provided");
        return;
    }
    const link = document.createElement("a");
    link.setAttribute("href", url);
    link.setAttribute("hidden", "hidden");
    link.setAttribute("download", filename);
    link.style.display = "none";
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
}

export function safeOpenUrl(url: string) {
    try {
        const parsedUrl = new URL(url);
        if (parsedUrl.protocol === 'http:' || parsedUrl.protocol === 'https:') {
            window.open(url, '_blank', 'noopener,noreferrer');
        } else {
            console.error('Invalid URL protocol. Only HTTP and HTTPS URLs are allowed:', url);
        }
    } catch (error) {
        console.error('Invalid URL format:', url, error);
    }
}
