/**
 * @generated SignedSource<<a8e224db1200fa5aaa6edc17c09e7a35>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
export type MembershipRole = "ADMIN" | "AUDITOR" | "EMPLOYEE" | "OWNER" | "VIEWER";
export type ProfileKind = "CONTRACTOR" | "EMPLOYEE" | "SERVICE_ACCOUNT";
import { FragmentRefs } from "relay-runtime";
export type PersonFormFragment$data = {
  readonly additionalEmailAddresses: ReadonlyArray<string>;
  readonly canUpdate: boolean;
  readonly contractEndDate: string | null | undefined;
  readonly contractStartDate: string | null | undefined;
  readonly emailAddress: string;
  readonly fullName: string;
  readonly id: string;
  readonly kind: ProfileKind;
  readonly membership: {
    readonly role: MembershipRole;
  };
  readonly position: string | null | undefined;
  readonly " $fragmentType": "PersonFormFragment";
};
export type PersonFormFragment$key = {
  readonly " $data"?: PersonFormFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"PersonFormFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "PersonFormFragment",
  "selections": [
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "id",
      "storageKey": null
    },
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
      "name": "emailAddress",
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
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "additionalEmailAddresses",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "contractStartDate",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "contractEndDate",
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

(node as any).hash = "4f99d9b9ed708290c5a7d7ec258a0b1d";

export default node;
