export function times<T>(n: number, cb: (i: number) => T): T[] {
    return Array.from({ length: n }, (_, i) => cb(i));
}

export function groupBy<T>(
    arr: T[],
    key: (item: T) => string,
): Record<string, T[]> {
    return arr.reduce(
        (acc, item) => {
            const k = key(item);
            if (!acc[k]) {
                acc[k] = [];
            }
            acc[k].push(item);
            return acc;
        },
        {} as Record<string, T[]>,
    );
}
