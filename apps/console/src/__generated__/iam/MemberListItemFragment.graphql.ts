/**
 * @generated SignedSource<<770fb773083cc4f819594e268d9eb0cc>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
export type MembershipRole = "ADMIN" | "AUDITOR" | "EMPLOYEE" | "OWNER" | "VIEWER";
export type MembershipSource = "MANUAL" | "SAML" | "SCIM";
export type MembershipState = "ACTIVE" | "INACTIVE";
export type ProfileKind = "CONTRACTOR" | "EMPLOYEE" | "SERVICE_ACCOUNT";
import { FragmentRefs } from "relay-runtime";
export type MemberListItemFragment$data = {
  readonly canDelete: boolean;
  readonly canUpdate: boolean;
  readonly createdAt: string;
  readonly id: string;
  readonly identity: {
    readonly email: string;
  };
  readonly profile: {
    readonly fullName: string;
    readonly id: string;
    readonly kind: ProfileKind;
    readonly position: string | null | undefined;
  };
  readonly role: MembershipRole;
  readonly source: MembershipSource;
  readonly state: MembershipState;
  readonly " $fragmentType": "MemberListItemFragment";
};
export type MemberListItemFragment$key = {
  readonly " $data"?: MemberListItemFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"MemberListItemFragment">;
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
  "name": "MemberListItemFragment",
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
      "name": "source",
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
      "kind": "RequiredField",
      "field": {
        "alias": null,
        "args": null,
        "concreteType": "MembershipProfile",
        "kind": "LinkedField",
        "name": "profile",
        "plural": false,
        "selections": [
          (v0/*: any*/),
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "fullName",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "kind",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "position",
            "storageKey": null
          }
        ],
        "storageKey": null
      },
      "action": "THROW"
    },
    {
      "kind": "RequiredField",
      "field": {
        "alias": null,
        "args": null,
        "concreteType": "Identity",
        "kind": "LinkedField",
        "name": "identity",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "email",
            "storageKey": null
          }
        ],
        "storageKey": null
      },
      "action": "THROW"
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "createdAt",
      "storageKey": null
    },
    {
      "alias": "canUpdate",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "iam:membership:update"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"iam:membership:update\")"
    },
    {
      "alias": "canDelete",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "iam:membership:delete"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"iam:membership:delete\")"
    }
  ],
  "type": "Membership",
  "abstractKey": null
};
})();

(node as any).hash = "378a4cdc212c16aecae368d41d80739e";

export default node;
