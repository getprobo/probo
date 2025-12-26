/**
 * @generated SignedSource<<0fa501d76c400538c62a44425ebe4b11>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type InvitationListItem_permissionsFragment$data = {
  readonly canDeleteInvitation: boolean;
  readonly " $fragmentType": "InvitationListItem_permissionsFragment";
};
export type InvitationListItem_permissionsFragment$key = {
  readonly " $data"?: InvitationListItem_permissionsFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"InvitationListItem_permissionsFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [
    {
      "defaultValue": null,
      "kind": "LocalArgument",
      "name": "organizationId"
    }
  ],
  "kind": "Fragment",
  "metadata": null,
  "name": "InvitationListItem_permissionsFragment",
  "selections": [
    {
      "alias": "canDeleteInvitation",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "iam:invitation:delete"
        },
        {
          "kind": "Variable",
          "name": "id",
          "variableName": "organizationId"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": null
    }
  ],
  "type": "Identity",
  "abstractKey": null
};

(node as any).hash = "e9e75ae0f5a3755a777acf16e5024ee0";

export default node;
