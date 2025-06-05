export function withViewTransition(fn: () => void) {
    if (!document.startViewTransition) {
        fn();
        return;
    }
    document.startViewTransition(fn);
}
