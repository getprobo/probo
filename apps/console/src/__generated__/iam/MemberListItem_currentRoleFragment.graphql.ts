/**
 * @generated SignedSource<<89f39cf7aa4cd7702b62f549ed8defdd>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
export type MembershipRole = "ADMIN" | "AUDITOR" | "EMPLOYEE" | "OWNER" | "VIEWER";
import { FragmentRefs } from "relay-runtime";
export type MemberListItem_currentRoleFragment$data = {
  readonly viewerMembership: {
    readonly role: MembershipRole;
  };
  readonly " $fragmentType": "MemberListItem_currentRoleFragment";
};
export type MemberListItem_currentRoleFragment$key = {
  readonly " $data"?: MemberListItem_currentRoleFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"MemberListItem_currentRoleFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "MemberListItem_currentRoleFragment",
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

(node as any).hash = "9e7e3eefcb6f2f4398b0703a9342548b";

export default node;
