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

export function slugify(str: string): string {
    return str
        .normalize("NFD") // split an accented letter in the base letter and the acent
        .replace(/[\u0300-\u036f]/g, "") // remove all previously split accents
        .toLowerCase()
        .trim()
        .replace(/[^a-z0-9 ]/g, "") // remove all chars not letters, numbers and spaces (to be replaced)
        .replace(/\s+/g, "-");
}
