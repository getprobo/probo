/**
 * @generated SignedSource<<602c62c8c2e8d6d5adb976196f46b65d>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
export type ReportState = "COMPLETED" | "IN_PROGRESS" | "NOT_STARTED" | "OUTDATED" | "REJECTED";
import { FragmentRefs } from "relay-runtime";
export type LinkedReportsCardFragment$data = {
  readonly createdAt: string;
  readonly framework: {
    readonly id: string;
    readonly name: string;
  };
  readonly id: string;
  readonly name: string | null | undefined;
  readonly state: ReportState;
  readonly validFrom: string | null | undefined;
  readonly validUntil: string | null | undefined;
  readonly " $fragmentType": "LinkedReportsCardFragment";
};
export type LinkedReportsCardFragment$key = {
  readonly " $data"?: LinkedReportsCardFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"LinkedReportsCardFragment">;
};

const node: ReaderFragment = (function(){
var v0 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
},
v1 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "name",
  "storageKey": null
};
return {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "LinkedReportsCardFragment",
  "selections": [
    (v0/*: any*/),
    (v1/*: any*/),
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
      "name": "state",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "validFrom",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "validUntil",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "concreteType": "Framework",
      "kind": "LinkedField",
      "name": "framework",
      "plural": false,
      "selections": [
        (v0/*: any*/),
        (v1/*: any*/)
      ],
      "storageKey": null
    }
  ],
  "type": "Report",
  "abstractKey": null
};
})();

(node as any).hash = "f4dfb5be5df05d90dec031eb3702fe1e";

export default node;
