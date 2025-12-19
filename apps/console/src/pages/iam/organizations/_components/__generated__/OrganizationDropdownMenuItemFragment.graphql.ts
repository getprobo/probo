/**
 * @generated SignedSource<<ef66c9eab961d5c3c28669629b82cfd6>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type OrganizationDropdownMenuItemFragment$data = {
  readonly id: string;
  readonly lastSession: {
    readonly expiresAt: any;
    readonly id: string;
  } | null | undefined;
  readonly organization: {
    readonly logoUrl: string | null | undefined;
    readonly name: string;
  };
  readonly " $fragmentType": "OrganizationDropdownMenuItemFragment";
};
export type OrganizationDropdownMenuItemFragment$key = {
  readonly " $data"?: OrganizationDropdownMenuItemFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"OrganizationDropdownMenuItemFragment">;
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
  "name": "OrganizationDropdownMenuItemFragment",
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

(node as any).hash = "6be7f721fd63d6d758b479e58c68b754";

export default node;
