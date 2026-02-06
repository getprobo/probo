/**
 * @generated SignedSource<<a74d82c45ea0f57774194b3acb02227e>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
export type ProfileKind = "CONTRACTOR" | "EMPLOYEE" | "SERVICE_ACCOUNT";
import { FragmentRefs } from "relay-runtime";
export type UserFormFragment$data = {
  readonly additionalEmailAddresses: ReadonlyArray<string>;
  readonly canUpdate: boolean;
  readonly contractEndDate: string | null | undefined;
  readonly contractStartDate: string | null | undefined;
  readonly fullName: string;
  readonly id: string;
  readonly kind: ProfileKind;
  readonly position: string | null | undefined;
  readonly " $fragmentType": "UserFormFragment";
};
export type UserFormFragment$key = {
  readonly " $data"?: UserFormFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"UserFormFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "UserFormFragment",
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
  "type": "MembershipProfile",
  "abstractKey": null
};

(node as any).hash = "b7c1eb9f3703338d5d0fb4c5411bf2eb";

export default node;
