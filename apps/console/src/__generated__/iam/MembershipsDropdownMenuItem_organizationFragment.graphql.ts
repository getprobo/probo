/**
 * @generated SignedSource<<a6754adeb9fc43ac3d85ed2e9c2c8aad>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type MembershipsDropdownMenuItem_organizationFragment$data = {
  readonly id: string;
  readonly logoUrl: string | null | undefined;
  readonly name: string;
  readonly " $fragmentType": "MembershipsDropdownMenuItem_organizationFragment";
};
export type MembershipsDropdownMenuItem_organizationFragment$key = {
  readonly " $data"?: MembershipsDropdownMenuItem_organizationFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"MembershipsDropdownMenuItem_organizationFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "MembershipsDropdownMenuItem_organizationFragment",
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

(node as any).hash = "799a0d538f5d667112f4c9ba8c0d3623";

export default node;
