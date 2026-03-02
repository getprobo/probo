/**
 * @generated SignedSource<<c712faa6c54f0f1da5895adb10d3f7c0>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type MeasureRisksTabFragment$data = {
  readonly canCreateRiskMeasureMapping: boolean;
  readonly canDeleteRiskMeasureMapping: boolean;
  readonly id: string;
  readonly risks: {
    readonly __id: string;
    readonly edges: ReadonlyArray<{
      readonly node: {
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
      "alias": "canCreateRiskMeasureMapping",
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
      "alias": "canDeleteRiskMeasureMapping",
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

(node as any).hash = "abab542141a748635126f652c0d79761";

export default node;
