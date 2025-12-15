import { useLottieRef } from "./useLottieRef";
import json from "./CASA.json";
import { AnimationStyle } from "./AnimationStyle";

export function CASA() {
    const ref = useLottieRef(json);

    return (
        <div ref={ref}>
            <AnimationStyle />
        </div>
    );
}
