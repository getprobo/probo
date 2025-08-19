/**
 * @generated SignedSource<<05658205222e432a115cecc9c8f34d05>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
export type SnapshotsType = "ASSETS" | "COMPLIANCE_REGISTRIES" | "DATA" | "NON_CONFORMITY_REGISTRIES" | "RISKS" | "VENDORS";
import { FragmentRefs } from "relay-runtime";
export type SnapshotsPageFragment$data = {
  readonly snapshots: {
    readonly __id: string;
    readonly edges: ReadonlyArray<{
      readonly node: {
        readonly createdAt: any;
        readonly description: string | null | undefined;
        readonly id: string;
        readonly name: string;
        readonly type: SnapshotsType;
      };
    }>;
  };
  readonly " $fragmentType": "SnapshotsPageFragment";
};
export type SnapshotsPageFragment$key = {
  readonly " $data"?: SnapshotsPageFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"SnapshotsPageFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": {
    "connection": [
      {
        "count": null,
        "cursor": null,
        "direction": "forward",
        "path": [
          "snapshots"
        ]
      }
    ]
  },
  "name": "SnapshotsPageFragment",
  "selections": [
    {
      "alias": "snapshots",
      "args": null,
      "concreteType": "SnapshotConnection",
      "kind": "LinkedField",
      "name": "__SnapshotsGraphListQuery__snapshots_connection",
      "plural": false,
      "selections": [
        {
          "alias": null,
          "args": null,
          "concreteType": "SnapshotEdge",
          "kind": "LinkedField",
          "name": "edges",
          "plural": true,
          "selections": [
            {
              "alias": null,
              "args": null,
              "concreteType": "Snapshot",
              "kind": "LinkedField",
              "name": "node",
              "plural": false,
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
                },
                {
                  "alias": null,
                  "args": null,
                  "kind": "ScalarField",
                  "name": "__typename",
                  "storageKey": null
                }
              ],
              "storageKey": null
            },
            {
              "alias": null,
              "args": null,
              "kind": "ScalarField",
              "name": "cursor",
              "storageKey": null
            }
          ],
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "concreteType": "PageInfo",
          "kind": "LinkedField",
          "name": "pageInfo",
          "plural": false,
          "selections": [
            {
              "alias": null,
              "args": null,
              "kind": "ScalarField",
              "name": "endCursor",
              "storageKey": null
            },
            {
              "alias": null,
              "args": null,
              "kind": "ScalarField",
              "name": "hasNextPage",
              "storageKey": null
            }
          ],
          "storageKey": null
        },
        {
          "kind": "ClientExtension",
          "selections": [
            {
              "alias": null,
              "args": null,
              "kind": "ScalarField",
              "name": "__id",
              "storageKey": null
            }
          ]
        }
      ],
      "storageKey": null
    }
  ],
  "type": "Organization",
  "abstractKey": null
};

(node as any).hash = "700dede55fd00f2b57dbbb6ed2e2613d";

export default node;
