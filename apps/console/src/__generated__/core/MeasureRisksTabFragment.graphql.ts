/**
 * @generated SignedSource<<4f367882dc2cd7bb8a8605d9192cd7f5>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type MeasureRisksTabFragment$data = {
  readonly id: string;
  readonly risks: {
    readonly __id: string;
    readonly edges: ReadonlyArray<{
      readonly node: {
        readonly canCreateMeasureMapping: boolean;
        readonly canDeleteMeasureMapping: boolean;
        readonly id: string;
        readonly " $fragmentSpreads": FragmentRefs<"LinkedRisksCardFragment">;
      };
    }>;
  };
  readonly " $fragmentType": "MeasureRisksTabFragment";
};
export type MeasureRisksTabFragment$key = {
  readonly " $data"?: MeasureRisksTabFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"MeasureRisksTabFragment">;
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
          "risks"
        ]
      }
    ]
  },
  "name": "MeasureRisksTabFragment",
  "selections": [
    (v0/*: any*/),
    {
      "alias": "risks",
      "args": null,
      "concreteType": "RiskConnection",
      "kind": "LinkedField",
      "name": "__Measure__risks_connection",
      "plural": false,
      "selections": [
        {
          "alias": null,
          "args": null,
          "concreteType": "RiskEdge",
          "kind": "LinkedField",
          "name": "edges",
          "plural": true,
          "selections": [
            {
              "alias": null,
              "args": null,
              "concreteType": "Risk",
              "kind": "LinkedField",
              "name": "node",
              "plural": false,
              "selections": [
                (v0/*: any*/),
                {
                  "alias": "canCreateMeasureMapping",
                  "args": [
                    {
                      "kind": "Literal",
                      "name": "action",
                      "value": "core:risk:create-measure-mapping"
                    }
                  ],
                  "kind": "ScalarField",
                  "name": "permission",
                  "storageKey": "permission(action:\"core:risk:create-measure-mapping\")"
                },
                {
                  "alias": "canDeleteMeasureMapping",
                  "args": [
                    {
                      "kind": "Literal",
                      "name": "action",
                      "value": "core:risk:delete-measure-mapping"
                    }
                  ],
                  "kind": "ScalarField",
                  "name": "permission",
                  "storageKey": "permission(action:\"core:risk:delete-measure-mapping\")"
                },
                {
                  "args": null,
                  "kind": "FragmentSpread",
                  "name": "LinkedRisksCardFragment"
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
  "type": "Measure",
  "abstractKey": null
};
})();

(node as any).hash = "a837a187026f5c75c67efcb4831d0b76";

export default node;
