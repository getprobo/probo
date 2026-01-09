/**
 * @generated SignedSource<<b69ceb13c09f597365a10a68e277d535>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
export type EvidenceType = "FILE" | "LINK";
import { FragmentRefs } from "relay-runtime";
export type MeasureEvidencesTabFragment_evidence$data = {
  readonly canDelete: boolean;
  readonly createdAt: string;
  readonly file: {
    readonly fileName: string;
    readonly mimeType: string;
    readonly size: number;
  } | null | undefined;
  readonly id: string;
  readonly type: EvidenceType;
  readonly " $fragmentType": "MeasureEvidencesTabFragment_evidence";
};
export type MeasureEvidencesTabFragment_evidence$key = {
  readonly " $data"?: MeasureEvidencesTabFragment_evidence$data;
  readonly " $fragmentSpreads": FragmentRefs<"MeasureEvidencesTabFragment_evidence">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "MeasureEvidencesTabFragment_evidence",
  "selections": [
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "id",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "concreteType": "File",
      "kind": "LinkedField",
      "name": "file",
      "plural": false,
      "selections": [
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "fileName",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "mimeType",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "size",
          "storageKey": null
        }
      ],
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "type",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "createdAt",
      "storageKey": null
    },
    {
      "alias": "canDelete",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:evidence:delete"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"core:evidence:delete\")"
    }
  ],
  "type": "Evidence",
  "abstractKey": null
};

(node as any).hash = "af214cdfe077ad77326c65f33ce3ca1f";

export default node;
