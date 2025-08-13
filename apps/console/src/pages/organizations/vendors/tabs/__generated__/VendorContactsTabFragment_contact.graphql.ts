/**
 * @generated SignedSource<<ff47c2ce47b00ef35a189dfc482826ec>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type VendorContactsTabFragment_contact$data = {
  readonly createdAt: any;
  readonly email: string | null | undefined;
  readonly id: string;
  readonly name: string | null | undefined;
  readonly phone: string | null | undefined;
  readonly role: string | null | undefined;
  readonly updatedAt: any;
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
      "name": "name",
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

(node as any).hash = "7c5df8f360009bee4a3da3ee68ec2708";

export default node;
