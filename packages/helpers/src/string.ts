export function sprintf(str: string, ...params: unknown[]): string {
    let i = 0;
    return str.replace(/%(\d+)?s/g, (_, ...args) => {
        if (args[0]) {
            return params[parseInt(args[0]) - 1] as string;
        }
        return params[i++] as string;
    });
}

export function faviconUrl(url?: string | null): string | null {
    if (!url) {
        return null;
    }

    try {
        const parsedUrl = new URL(url);
        return `https://www.google.com/s2/favicons?domain=${parsedUrl.hostname}&sz=64`;
    } catch (e) {
        return null;
    }
}
