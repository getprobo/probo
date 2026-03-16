/**
 * @generated SignedSource<<bb740374a6287bb7152c60a80b4a3567>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
export type AuditState = "COMPLETED" | "IN_PROGRESS" | "NOT_STARTED" | "OUTDATED" | "REJECTED";
import { FragmentRefs } from "relay-runtime";
export type LinkedAuditsCardFragment$data = {
  readonly framework: {
    readonly id: string;
    readonly name: string;
  };
  readonly id: string;
  readonly name: string | null | undefined;
  readonly state: AuditState;
  readonly " $fragmentType": "LinkedAuditsCardFragment";
};
export type LinkedAuditsCardFragment$key = {
  readonly " $data"?: LinkedAuditsCardFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"LinkedAuditsCardFragment">;
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
  "name": "LinkedAuditsCardFragment",
  "selections": [
    (v0/*: any*/),
    (v1/*: any*/),
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
  "type": "Audit",
  "abstractKey": null
};
})();

(node as any).hash = "8a8cefa204a3afde548d70c072bd3b57";

export default node;
