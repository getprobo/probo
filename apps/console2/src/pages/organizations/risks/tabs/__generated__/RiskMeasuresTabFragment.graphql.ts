/**
 * @generated SignedSource<<83a07ae53aaa5005db49f5ea8b4479c6>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
export type MeasureState = "IMPLEMENTED" | "IN_PROGRESS" | "NOT_APPLICABLE" | "NOT_STARTED" | "%future added value";
import { FragmentRefs } from "relay-runtime";
export type RiskMeasuresTabFragment$data = {
  readonly id: string;
  readonly measures: {
    readonly __id: string;
    readonly edges: ReadonlyArray<{
      readonly node: {
        readonly category: string;
        readonly createdAt: any;
        readonly description: string;
        readonly id: string;
        readonly name: string;
        readonly state: MeasureState;
      };
    }>;
  };
  readonly " $fragmentType": "RiskMeasuresTabFragment";
};
export type RiskMeasuresTabFragment$key = {
  readonly " $data"?: RiskMeasuresTabFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"RiskMeasuresTabFragment">;
};

const node: ReaderFragment = (function(){
var v0 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
};
return {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": {
    "connection": [
      {
        "count": null,
        "cursor": null,
        "direction": "forward",
        "path": [
          "measures"
        ]
      }
    ]
  },
  "name": "RiskMeasuresTabFragment",
  "selections": [
    (v0/*: any*/),
    {
      "alias": "measures",
      "args": null,
      "concreteType": "MeasureConnection",
      "kind": "LinkedField",
      "name": "__Risk__measures_connection",
      "plural": false,
      "selections": [
        {
          "alias": null,
          "args": null,
          "concreteType": "MeasureEdge",
          "kind": "LinkedField",
          "name": "edges",
          "plural": true,
          "selections": [
            {
              "alias": null,
              "args": null,
              "concreteType": "Measure",
              "kind": "LinkedField",
              "name": "node",
              "plural": false,
              "selections": [
                (v0/*: any*/),
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
                  "name": "category",
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
                  "name": "state",
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
  "type": "Risk",
  "abstractKey": null
};
})();

(node as any).hash = "e04c5962a40b48aade0d1e4a31f0c583";

export default node;
