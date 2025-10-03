/**
 * @generated SignedSource<<0fb98c755d9aacac0ef1d508af9e7edc>>
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
  readonly createdAt: any;
  readonly file: {
    readonly fileName: string;
    readonly mimeType: string;
    readonly size: any;
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
    }
  ],
  "type": "Evidence",
  "abstractKey": null
};

(node as any).hash = "441ff1dd2dd62fea27050c5f550b27ea";

export default node;
