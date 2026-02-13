/**
 * @generated SignedSource<<6ee89b2ab0c55ac9498bf9186e78ee90>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
export type ProfileKind = "CONTRACTOR" | "EMPLOYEE" | "SERVICE_ACCOUNT";
import { FragmentRefs } from "relay-runtime";
export type PersonFormFragment$data = {
  readonly additionalEmailAddresses: ReadonlyArray<string>;
  readonly canUpdate: boolean;
  readonly contractEndDate: string | null | undefined;
  readonly contractStartDate: string | null | undefined;
  readonly fullName: string;
  readonly id: string;
  readonly kind: ProfileKind;
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

(node as any).hash = "9ecc41c74271a3202bbde08b63c92b73";

export default node;
