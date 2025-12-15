import { useLottieRef } from "./useLottieRef";
import json from "./FERPA.json";
import { AnimationStyle } from "./AnimationStyle";

export function FERPA() {
    const ref = useLottieRef(json);

    return (
        <div ref={ref}>
            <AnimationStyle />
        </div>
    );
}
