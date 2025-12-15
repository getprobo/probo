import { useLottieRef } from "./useLottieRef";
import json from "./ISO42001.json";
import { AnimationStyle } from "./AnimationStyle";

export function ISO42001() {
    const ref = useLottieRef(json);

    return (
        <div ref={ref}>
            <AnimationStyle />
        </div>
    );
}
