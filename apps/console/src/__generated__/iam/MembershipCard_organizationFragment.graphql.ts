/**
 * @generated SignedSource<<b3dda685c5f3688ead238ddd46d12dd3>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type MembershipCard_organizationFragment$data = {
  readonly id: string;
  readonly logoUrl: string | null | undefined;
  readonly name: string;
  readonly " $fragmentType": "MembershipCard_organizationFragment";
};
export type MembershipCard_organizationFragment$key = {
  readonly " $data"?: MembershipCard_organizationFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"MembershipCard_organizationFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "MembershipCard_organizationFragment",
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
      "name": "logoUrl",
      "storageKey": null
    }
  ],
  "type": "Organization",
  "abstractKey": null
};

(node as any).hash = "a02d28730634e510e7e0803d5d1a366f";

export default node;
