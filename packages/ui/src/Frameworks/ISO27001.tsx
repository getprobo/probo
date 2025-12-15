import { useLottieRef } from "./useLottieRef";
import json from "./ISO27001.json";
import { AnimationStyle } from "./AnimationStyle";

export function ISO27001() {
    const ref = useLottieRef(json);

    return (
        <div ref={ref}>
            <AnimationStyle />
        </div>
    );
}
