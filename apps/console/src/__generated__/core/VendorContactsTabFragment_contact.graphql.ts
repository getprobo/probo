/**
 * @generated SignedSource<<3ada4a1d28f4a4b9849a4a5a7d47443a>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type VendorContactsTabFragment_contact$data = {
  readonly canDelete: boolean;
  readonly canUpdate: boolean;
  readonly email: string | null | undefined;
  readonly fullName: string | null | undefined;
  readonly id: string;
  readonly phone: string | null | undefined;
  readonly role: string | null | undefined;
  readonly " $fragmentType": "VendorContactsTabFragment_contact";
};
export type VendorContactsTabFragment_contact$key = {
  readonly " $data"?: VendorContactsTabFragment_contact$data;
  readonly " $fragmentSpreads": FragmentRefs<"VendorContactsTabFragment_contact">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "VendorContactsTabFragment_contact",
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
      "name": "phone",
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
      "alias": "canUpdate",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:vendor-contact:update"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"core:vendor-contact:update\")"
    },
    {
      "alias": "canDelete",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:vendor-contact:delete"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"core:vendor-contact:delete\")"
    }
  ],
  "type": "VendorContact",
  "abstractKey": null
};

(node as any).hash = "60130c05b2e55237b385e97e51543d6b";

export default node;
