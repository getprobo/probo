/**
 * @generated SignedSource<<f05da6a037d3f282ff9ff6d60e2d684d>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type EmployeeDocumentSignaturePageVersionFragment$data = {
  readonly id: string;
  readonly publishedAt: any | null | undefined;
  readonly signed: boolean;
  readonly version: number;
  readonly " $fragmentType": "EmployeeDocumentSignaturePageVersionFragment";
};
export type EmployeeDocumentSignaturePageVersionFragment$key = {
  readonly " $data"?: EmployeeDocumentSignaturePageVersionFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"EmployeeDocumentSignaturePageVersionFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "EmployeeDocumentSignaturePageVersionFragment",
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
      "name": "version",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "signed",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "publishedAt",
      "storageKey": null
    }
  ],
  "type": "DocumentVersion",
  "abstractKey": null
};

(node as any).hash = "4a85fbbc1bf8b2610f554aa439fd0e95";

export default node;
