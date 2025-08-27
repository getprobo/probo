/**
 * @generated SignedSource<<e9fde8de02cc2e3ec97ce64b67818bfe>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
export type SnapshotsType = "ASSETS" | "COMPLIANCE_REGISTRIES" | "DATA" | "NON_CONFORMITY_REGISTRIES" | "RISKS" | "VENDORS";
import { FragmentRefs } from "relay-runtime";
export type LinkedSnapshotsCardFragment$data = {
  readonly createdAt: any;
  readonly description: string | null | undefined;
  readonly id: string;
  readonly name: string;
  readonly type: SnapshotsType;
  readonly " $fragmentType": "LinkedSnapshotsCardFragment";
};
export type LinkedSnapshotsCardFragment$key = {
  readonly " $data"?: LinkedSnapshotsCardFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"LinkedSnapshotsCardFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "LinkedSnapshotsCardFragment",
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
      "name": "description",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "type",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "createdAt",
      "storageKey": null
    }
  ],
  "type": "Snapshot",
  "abstractKey": null
};

(node as any).hash = "b9b682f8e57082c98d8b0460155757fb";

export default node;
