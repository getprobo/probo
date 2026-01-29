/**
 * @generated SignedSource<<042ba763926f3dddfcf4f9d673ae5201>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type ConnectorListFragment$data = {
  readonly scimConfiguration: {
    readonly " $fragmentSpreads": FragmentRefs<"GoogleWorkspaceConnectorFragment">;
  } | null | undefined;
  readonly " $fragmentType": "ConnectorListFragment";
};
export type ConnectorListFragment$key = {
  readonly " $data"?: ConnectorListFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"ConnectorListFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "ConnectorListFragment",
  "selections": [
    {
      "alias": null,
      "args": null,
      "concreteType": "SCIMConfiguration",
      "kind": "LinkedField",
      "name": "scimConfiguration",
      "plural": false,
      "selections": [
        {
          "args": null,
          "kind": "FragmentSpread",
          "name": "GoogleWorkspaceConnectorFragment"
        }
      ],
      "storageKey": null
    }
  ],
  "type": "Organization",
  "abstractKey": null
};

(node as any).hash = "32fd94ab5080bb8deba95ab24b545fc1";

export default node;
