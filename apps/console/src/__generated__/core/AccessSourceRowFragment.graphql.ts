/**
 * @generated SignedSource<<46040cd0545a3c52dc40a6cb33db0151>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
export type ConnectorProvider = "GOOGLE_WORKSPACE" | "LINEAR" | "SLACK";
import { FragmentRefs } from "relay-runtime";
export type AccessSourceRowFragment$data = {
  readonly canDelete: boolean;
  readonly connector: {
    readonly provider: ConnectorProvider;
  } | null | undefined;
  readonly connectorId: string | null | undefined;
  readonly createdAt: string;
  readonly id: string;
  readonly name: string;
  readonly " $fragmentType": "AccessSourceRowFragment";
};
export type AccessSourceRowFragment$key = {
  readonly " $data"?: AccessSourceRowFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"AccessSourceRowFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "AccessSourceRowFragment",
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
      "name": "connectorId",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "concreteType": "Connector",
      "kind": "LinkedField",
      "name": "connector",
      "plural": false,
      "selections": [
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "provider",
          "storageKey": null
        }
      ],
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
      "alias": "canDelete",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:access-source:delete"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"core:access-source:delete\")"
    }
  ],
  "type": "AccessSource",
  "abstractKey": null
};

(node as any).hash = "e428b8ba8ab7a89381009ab90798f2a0";

export default node;
