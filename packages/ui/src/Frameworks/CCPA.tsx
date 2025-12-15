import { useLottieRef } from "./useLottieRef";
import json from "./CCPA.json";
import { AnimationStyle } from "./AnimationStyle";

export function CCPA() {
    const ref = useLottieRef(json);

    return (
        <div ref={ref}>
            <AnimationStyle />
        </div>
    );
}
