import {
    lazy as reactLazy,
    type LazyExoticComponent,
    type ComponentType,
    type FC,
} from "react";

const DEFAULT_STORAGE_KEY = "reactLazyReloadedFunctions";

class ReloadStorage {
    public readonly storageKey: string;

    constructor(storageKey = DEFAULT_STORAGE_KEY) {
        this.storageKey = storageKey;
    }

    getMap(): Map<string, number> {
        const stored = sessionStorage.getItem(this.storageKey);

        try {
            const parsed: unknown = stored ? JSON.parse(stored) : [];
            return new Map(
                Array.isArray(parsed)
                    ? parsed.filter(
                          (value) =>
                              Array.isArray(value) &&
                              typeof value[0] === "string" &&
                              typeof value[1] === "number",
                      )
                    : [],
            );
        } catch (error) {
            console.error("Error parsing react-lazy storage:", error);
            return new Map();
        }
    }

    addFunction(functionName: string): void {
        const stored = this.getMap();
        stored.set(functionName, (stored.get(functionName) ?? 0) + 1);
        this.save(stored);
    }

    removeFunction(functionName: string): void {
        const stored = this.getMap();
        stored.delete(functionName);
        this.save(stored);
    }

    protected save(stored: Map<string, number>): void {
        sessionStorage.setItem(
            this.storageKey,
            JSON.stringify([...stored.entries()]),
        );
    }
}

type ForceReloadConfig = {
    maxRetries: number;
    storageKey?: string;
};

type LazyConfigInit = {
    forceReload?: false | Partial<ForceReloadConfig>;
    importRetries?: number;
    retryDelay?: number;
    onImportError?: (error: unknown) => void;
};

type LazyConfig = {
    forceReload: ForceReloadConfig;
    importRetries: number;
    retryDelay: number;
    onImportError?: (error: unknown) => void;
};

const createConfig = (init: LazyConfigInit = {}): LazyConfig => ({
    forceReload: {
        maxRetries: 1,
        storageKey: DEFAULT_STORAGE_KEY,
        ...(typeof init.forceReload === "object"
            ? init.forceReload
            : init.forceReload === false
              ? { maxRetries: 0 }
              : {}),
    },
    importRetries:
        typeof init.importRetries === "number"
            ? Math.max(0, init.importRetries)
            : 0,
    retryDelay:
        typeof init.retryDelay === "number"
            ? Math.max(0, init.retryDelay)
            : 300,
    onImportError: init.onImportError,
});

export function createLazy<P = {}>(config: LazyConfigInit = {}) {
    const { forceReload, importRetries, retryDelay, onImportError } =
        createConfig(config);
    const reloadStorage = new ReloadStorage(forceReload.storageKey);

    function enhancedLazy<T = P>(
        importFunction: () => Promise<{ default: ComponentType<T> }>,
    ): LazyExoticComponent<ComponentType<T>> {
        const functionString = importFunction.toString();

        const importWithRetry = async () => {
            let retryCount = 0;

            const tryImport = async (): Promise<{
                default: ComponentType<T>;
            }> => {
                try {
                    return await importFunction();
                } catch (error) {
                    if (onImportError) {
                        onImportError(error);
                    } else {
                        console.error("React-lazy import error:", error);
                    }

                    if (retryCount < importRetries) {
                        retryCount++;

                        if (retryDelay > 0) {
                            await new Promise((resolve) =>
                                setTimeout(resolve, retryDelay),
                            );
                        }

                        return tryImport();
                    }

                    throw error;
                }
            };

            return tryImport();
        };

        return reactLazy(async () => {
            try {
                const component = await importWithRetry();
                if (forceReload.maxRetries > 0) {
                    reloadStorage.removeFunction(functionString);
                }
                return component;
            } catch (error) {
                if (
                    forceReload.maxRetries > 0 &&
                    (reloadStorage.getMap().get(functionString) ?? 0) <
                        forceReload.maxRetries
                ) {
                    reloadStorage.addFunction(functionString);
                    window.location.reload();

                    const EmptyComponent: FC<T> = () => null;
                    return { default: EmptyComponent };
                }
                throw error;
            }
        });
    }

    return enhancedLazy;
}

export const lazy = createLazy({
    forceReload: {
        maxRetries: 3,
        storageKey: DEFAULT_STORAGE_KEY,
    },
    importRetries: 3,
    retryDelay: 300,
});
