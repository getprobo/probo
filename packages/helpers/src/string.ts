export function sprintf(str: string, ...params: unknown[]): string {
    let i = 0;
    return str.replace(/%(\d+)?s/g, (_, ...args) => {
        if (args[0]) {
            return params[parseInt(args[0]) - 1] as string;
        }
        return params[i++] as string;
    });
}
