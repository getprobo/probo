export function AnimationStyle() {
    return (
        <style>{`
@media (prefers-color-scheme: dark) {
    path[fill="rgb(0,0,0)"] { fill: #ffffff; }
    path[fill="rgb(255,255,255)"] { fill: #000000; }
    path[fill="rgb(20,31,18)"] { fill: #c4f042; }
    path[fill="rgb(20,30,18)"] { fill: #c4f042; }
    path[stroke="rgb(0,0,0)"] { stroke: #ffffff; }
    path[stroke="rgb(255,255,255)"] { stroke: #000000; }
    path[stroke="rgb(20,31,18)"] { stroke: #c4f042; }
    path[stroke="rgb(20,30,18)"] { stroke: #c4f042; }
    g[fill="rgb(20,31,18)"] { fill: #c4f042; }
}
`}</style>
    );
}
