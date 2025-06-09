/**
 * @generated SignedSource<<aa88116272684a0ebbca38e57c798eee>>
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
  readonly fileUrl: string | null | undefined;
  readonly filename: string;
  readonly id: string;
  readonly mimeType: string;
  readonly size: number;
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
      "kind": "ScalarField",
      "name": "filename",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "size",
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
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "fileUrl",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "mimeType",
      "storageKey": null
    }
  ],
  "type": "Evidence",
  "abstractKey": null
};

(node as any).hash = "4fa5951157f130a8ed644151ef5facd6";

export default node;
