/**
 * @generated SignedSource<<d4c36c6798ea000539ef714721fd514f>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
export type MembershipRole = "ADMIN" | "AUDITOR" | "EMPLOYEE" | "OWNER" | "VIEWER";
import { FragmentRefs } from "relay-runtime";
export type InvitationCardFragment$data = {
  readonly createdAt: string;
  readonly id: string;
  readonly organization: {
    readonly id: string;
    readonly name: string;
  };
  readonly role: MembershipRole;
  readonly " $fragmentType": "InvitationCardFragment";
};
export type InvitationCardFragment$key = {
  readonly " $data"?: InvitationCardFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"InvitationCardFragment">;
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
  "name": "InvitationCardFragment",
  "selections": [
    (v0/*: any*/),
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "role",
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
            "name": "name",
            "storageKey": null
          }
        ],
        "storageKey": null
      },
      "action": "THROW"
    }
  ],
  "type": "Invitation",
  "abstractKey": null
};
})();

(node as any).hash = "57aedeabc62e474c6e5b995af1d31888";

export default node;
