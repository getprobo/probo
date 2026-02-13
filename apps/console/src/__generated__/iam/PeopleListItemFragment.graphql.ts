/**
 * @generated SignedSource<<6e9b0a4fbf6ce65f11f2367f2ec5b6d4>>
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
export type PeopleListItemFragment$data = {
  readonly canUpdate: boolean;
  readonly createdAt: string;
  readonly fullName: string;
  readonly id: string;
  readonly identity: {
    readonly email: string;
  };
  readonly kind: ProfileKind;
  readonly membership: {
    readonly canDelete: boolean;
    readonly canUpdate: boolean;
    readonly id: string;
    readonly role: MembershipRole;
    readonly source: MembershipSource;
    readonly state: MembershipState;
  };
  readonly position: string | null | undefined;
  readonly " $fragmentType": "PeopleListItemFragment";
};
export type PeopleListItemFragment$key = {
  readonly " $data"?: PeopleListItemFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"PeopleListItemFragment">;
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
  "name": "PeopleListItemFragment",
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
    },
    {
      "kind": "RequiredField",
      "field": {
        "alias": null,
        "args": null,
        "concreteType": "Membership",
        "kind": "LinkedField",
        "name": "membership",
        "plural": false,
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
                "value": "iam:membership-profile:delete"
              }
            ],
            "kind": "ScalarField",
            "name": "permission",
            "storageKey": "permission(action:\"iam:membership-profile:delete\")"
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
          "value": "iam:membership-profile:update"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"iam:membership-profile:update\")"
    }
  ],
  "type": "Profile",
  "abstractKey": null
};
})();

(node as any).hash = "89702fbdf02294212269adb22efb65fe";

export default node;
