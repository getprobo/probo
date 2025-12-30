/**
 * @generated SignedSource<<748bbadd485b1e131e583c119f980e91>>
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
  readonly acceptedAt: any | null | undefined;
  readonly createdAt: any;
  readonly email: any;
  readonly expiresAt: any;
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
    }
  ],
  "type": "Invitation",
  "abstractKey": null
};

(node as any).hash = "772414270b8ce2fed6c7b5fe97082c17";

export default node;
