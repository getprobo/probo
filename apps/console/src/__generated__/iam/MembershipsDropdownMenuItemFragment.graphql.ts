/**
 * @generated SignedSource<<e907a3f8c0610d5deed37b1a47500377>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type MembershipsDropdownMenuItemFragment$data = {
  readonly id: string;
  readonly lastSession: {
    readonly expiresAt: string;
    readonly id: string;
  } | null | undefined;
  readonly organization: {
    readonly id: string;
    readonly logoUrl: string | null | undefined;
    readonly name: string;
  };
  readonly " $fragmentType": "MembershipsDropdownMenuItemFragment";
};
export type MembershipsDropdownMenuItemFragment$key = {
  readonly " $data"?: MembershipsDropdownMenuItemFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"MembershipsDropdownMenuItemFragment">;
};

const node: ReaderFragment = (function(){
var v0 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
};
return {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "MembershipsDropdownMenuItemFragment",
  "selections": [
    (v0/*: any*/),
    {
      "alias": null,
      "args": null,
      "concreteType": "Session",
      "kind": "LinkedField",
      "name": "lastSession",
      "plural": false,
      "selections": [
        (v0/*: any*/),
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "expiresAt",
          "storageKey": null
        }
      ],
      "storageKey": null
    },
    {
      "kind": "RequiredField",
      "field": {
        "alias": null,
        "args": null,
        "concreteType": "Organization",
        "kind": "LinkedField",
        "name": "organization",
        "plural": false,
        "selections": [
          (v0/*: any*/),
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "logoUrl",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "name",
            "storageKey": null
          }
        ],
        "storageKey": null
      },
      "action": "THROW"
    }
  ],
  "type": "Membership",
  "abstractKey": null
};
})();

(node as any).hash = "915a6fb0071b4890010401a56f3affa6";

export default node;
