/**
 * @generated SignedSource<<de896a2d4f3d1dce06f7b88630afb445>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
export type MembershipRole = "ADMIN" | "AUDITOR" | "EMPLOYEE" | "OWNER" | "VIEWER";
import { FragmentRefs } from "relay-runtime";
export type InviteUserDialog_currentRoleFragment$data = {
  readonly viewerMembership: {
    readonly role: MembershipRole;
  };
  readonly " $fragmentType": "InviteUserDialog_currentRoleFragment";
};
export type InviteUserDialog_currentRoleFragment$key = {
  readonly " $data"?: InviteUserDialog_currentRoleFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"InviteUserDialog_currentRoleFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "InviteUserDialog_currentRoleFragment",
  "selections": [
    {
      "kind": "RequiredField",
      "field": {
        "alias": null,
        "args": null,
        "concreteType": "Membership",
        "kind": "LinkedField",
        "name": "viewerMembership",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "role",
            "storageKey": null
          }
        ],
        "storageKey": null
      },
      "action": "THROW"
    }
  ],
  "type": "Organization",
  "abstractKey": null
};

(node as any).hash = "167528734d4ad650128d5ed8f201c97e";

export default node;
