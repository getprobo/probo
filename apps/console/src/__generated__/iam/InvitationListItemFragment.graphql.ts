/**
 * @generated SignedSource<<5b60de19e893c841d45ecd555eecf65a>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
export type InvitationStatus = "ACCEPTED" | "EXPIRED" | "PENDING";
export type MembershipRole = "ADMIN" | "AUDITOR" | "EMPLOYEE" | "OWNER" | "VIEWER";
import { FragmentRefs } from "relay-runtime";
export type InvitationListItemFragment$data = {
  readonly acceptedAt: string | null | undefined;
  readonly canDelete: boolean;
  readonly createdAt: string;
  readonly email: string;
  readonly expiresAt: string;
  readonly fullName: string;
  readonly id: string;
  readonly role: MembershipRole;
  readonly status: InvitationStatus;
  readonly " $fragmentType": "InvitationListItemFragment";
};
export type InvitationListItemFragment$key = {
  readonly " $data"?: InvitationListItemFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"InvitationListItemFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "InvitationListItemFragment",
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
      "name": "email",
      "storageKey": null
    },
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
      "name": "status",
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
      "kind": "ScalarField",
      "name": "expiresAt",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "acceptedAt",
      "storageKey": null
    },
    {
      "alias": "canDelete",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "iam:invitation:delete"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"iam:invitation:delete\")"
    }
  ],
  "type": "Invitation",
  "abstractKey": null
};

(node as any).hash = "c0e21d5dd227cd3fff739dc9f0922657";

export default node;
