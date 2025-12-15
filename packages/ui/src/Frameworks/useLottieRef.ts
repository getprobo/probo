import { useEffect, useRef, useState } from "react";
import lottie, { type AnimationItem } from "lottie-web/build/player/lottie_light";

type Options = {
    threshold?: number;
    once?: boolean;
}

const defaultOptions: Options = { threshold: 0.5, once: false }
// Lottie expects any for animation data
// eslint-disable-next-line @typescript-eslint/no-explicit-any
export function useLottieRef(animationData: any, options: Options = defaultOptions) {
    const ref = useRef<HTMLDivElement>(null);
    const [observed, setObserved] = useState<boolean>(false);
    const animation = useRef<AnimationItem>(null)

    useEffect(() => {
        if (ref.current) {
            const observer = new IntersectionObserver(([entry]) => {
                if (entry.isIntersecting && options.once) {
                    observer.disconnect();
                }
                setObserved(entry.isIntersecting)
            }, {threshold: options.threshold});

            observer.observe(ref.current);

            return () => {
                observer.disconnect();
            };
        }
    });

    useEffect(() => {
        // Load the animation when the element is in the viewport.
        if (observed && !animation.current && ref.current) {
            animation.current = lottie.loadAnimation({
                container: ref.current,
                renderer: "svg",
                loop: false,
                autoplay: true,
                animationData: animationData,
            });
            animation.current.addEventListener("complete", () => {
                if (!animation) {
                    return;
                }
                animation.current?.goToAndPlay(300, true);
            });
            return;
        }

        if (!animation) {
            return;
        }

        if (!observed) {
            animation.current?.stop();
            return;
        }

        if (observed) {
            animation.current?.play();
            return;
        }

        return () => {
            animation.current?.destroy()
        }
    })

    return ref;
}