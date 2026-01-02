/**
 * @generated SignedSource<<ea6290653015d098eabae99bb4ff4061>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type VendorContactsTabFragment_contact$data = {
  readonly createdAt: string;
  readonly email: string | null | undefined;
  readonly fullName: string | null | undefined;
  readonly id: string;
  readonly phone: string | null | undefined;
  readonly role: string | null | undefined;
  readonly updatedAt: string;
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
      "name": "updatedAt",
      "storageKey": null
    }
  ],
  "type": "VendorContact",
  "abstractKey": null
};

(node as any).hash = "9c3e5b8b262263249160f5fd3708162a";

export default node;
