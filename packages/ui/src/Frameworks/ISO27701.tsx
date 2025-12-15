import { useLottieRef } from "./useLottieRef";
import json from "./ISO27701.json";
import { AnimationStyle } from "./AnimationStyle";

export function ISO27701() {
    const ref = useLottieRef(json);

    return (
        <div ref={ref}>
            <AnimationStyle />
        </div>
    );
}
