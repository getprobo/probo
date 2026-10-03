const DIMMED_OPACITY = "0.3";
const EMPHASIZED_OPACITY = "1";

type Edge = {
  id: string;
  elements: SVGElement[];
  nodeIDs: [string, string] | null;
};

// Extract Mermaid's logical node ID from either its data attribute or flowchart DOM ID.
function getNodeID(node: SVGElement) {
  // Some Mermaid diagram types expose the logical ID directly.
  const nodeID = node.getAttribute("data-id");
  if (nodeID) {
    return nodeID;
  }

  // Flowcharts encode it as "...-flowchart-<node ID>-<index>".
  const id = node.getAttribute("id");
  return id?.match(/(?:^|-)flowchart-(.+)-\d+$/)?.[1] ?? null;
}

// Resolve an edge's endpoints by matching Mermaid's L_<source>_<target>_<index> ID.
function getEdgeNodeIDs(edgeID: string, nodeIDs: string[]) {
  // Testing known IDs keeps node names containing underscores supported.
  for (const sourceID of nodeIDs) {
    for (const targetID of nodeIDs) {
      if (edgeID.startsWith(`L_${sourceID}_${targetID}_`)) {
        return [sourceID, targetID] as [string, string];
      }
    }
  }

  return null;
}

export function addMermaidInteractions(svg: SVGElement) {
  // Collect nodes and their logical IDs once
  const nodes = [...svg.querySelectorAll<SVGElement>(".node")];
  const nodeIDs = nodes.map(getNodeID).filter((id): id is string => id !== null);
  const edgeElements = [...svg.querySelectorAll<SVGElement>("[data-edge='true']")];
  // Index edges by ID so clicks on an edge label resolve to the rendered path as well.
  const edgesByID = new Map<string, Edge>();
  for (const element of edgeElements) {
    const id = element.getAttribute("data-id");
    if (!id) {
      continue;
    }

    edgesByID.set(id, {
      id,
      elements: [
        element,
        // Edge labels are separate SVG groups and must fade with their path.
        ...svg.querySelectorAll<SVGElement>(`.edgeLabels [data-id='${id}']`),
      ],
      nodeIDs: getEdgeNodeIDs(id, nodeIDs),
    });
  }

  const visualElements = [
    ...nodes,
    ...[...edgesByID.values()].flatMap(edge => edge.elements),
  ];
  // Preserve Mermaid's inline styles so reset restores the exact initial rendering.
  const initialStyles = new Map(
    visualElements.map(element => [element, element.getAttribute("style")]),
  );

  // Restore all interactive elements to their original Mermaid styles.
  function reset() {
    for (const element of visualElements) {
      const initialStyle = initialStyles.get(element);
      if (initialStyle === null) {
        element.removeAttribute("style");
      } else {
        element.setAttribute("style", initialStyle ?? "");
      }
    }
  }

  // Dim the full graph, except for the supplied nodes and edges.
  function emphasize(nodeIDsToEmphasize: Set<string>, edgeIDsToEmphasize: Set<string>) {
    for (const node of nodes) {
      const isEmphasized = nodeIDsToEmphasize.has(getNodeID(node) ?? "");
      node.style.opacity = isEmphasized ? EMPHASIZED_OPACITY : DIMMED_OPACITY;
    }

    for (const edge of edgesByID.values()) {
      const isEmphasized = edgeIDsToEmphasize.has(edge.id);
      for (const element of edge.elements) {
        element.style.opacity = isEmphasized ? EMPHASIZED_OPACITY : DIMMED_OPACITY;
      }
    }
  }

  // Use one SVG-level listener so clicks on nested labels and shapes behave consistently.
  function onClick(event: MouseEvent) {
    if (!(event.target instanceof Element)) {
      reset();
      return;
    }

    const node = event.target.closest(".node");
    if (node instanceof SVGElement && svg.contains(node)) {
      const nodeID = getNodeID(node);
      if (!nodeID) {
        reset();
        return;
      }

      // A node selection includes every directly connected node and its incident edges.
      const associatedEdges = [...edgesByID.values()].filter(edge => edge.nodeIDs?.includes(nodeID));
      emphasize(
        new Set([nodeID, ...associatedEdges.flatMap(edge => edge.nodeIDs ?? [])]),
        new Set(associatedEdges.map(edge => edge.id)),
      );
      return;
    }

    const edgeElement = event.target.closest("[data-edge='true'], .edgeLabels [data-id]");
    const edgeID = edgeElement?.getAttribute("data-id");
    const edge = edgeID ? edgesByID.get(edgeID) : undefined;
    if (edge?.nodeIDs) {
      // An edge selection only keeps that edge and its two endpoint nodes visible.
      emphasize(new Set(edge.nodeIDs), new Set([edge.id]));
      return;
    }

    // Clicking the SVG background clears the active selection.
    reset();
  }

  // Attach the delegated click handler after Mermaid has populated the SVG.
  svg.addEventListener("click", onClick);

  // Let the caller remove the listener and restore the diagram when it unmounts.
  return () => {
    svg.removeEventListener("click", onClick);
    reset();
  };
}
