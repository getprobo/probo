/**
 * @generated SignedSource<<5193da6ee4b5da0bed510c255c3407b1>>
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
  readonly createdAt: any;
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
    }
  ],
  "type": "Invitation",
  "abstractKey": null
};
})();

(node as any).hash = "1f61b5f07abc69ad33c880bc39cd0e96";

export default node;
