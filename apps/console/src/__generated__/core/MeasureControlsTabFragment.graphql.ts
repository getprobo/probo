/**
 * @generated SignedSource<<414658fc55873606e711cc4571e9fd9b>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type MeasureControlsTabFragment$data = {
  readonly controls: {
    readonly __id: string;
    readonly edges: ReadonlyArray<{
      readonly node: {
        readonly canCreateMeasureMapping: boolean;
        readonly canDeleteMeasureMapping: boolean;
        readonly id: string;
        readonly " $fragmentSpreads": FragmentRefs<"LinkedControlsCardFragment">;
      };
    }>;
  };
  readonly id: string;
  readonly " $fragmentType": "MeasureControlsTabFragment";
};
export type MeasureControlsTabFragment$key = {
  readonly " $data"?: MeasureControlsTabFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"MeasureControlsTabFragment">;
};

import MeasureControlsTabControlsQuery_graphql from './MeasureControlsTabControlsQuery.graphql';

const node: ReaderFragment = (function(){
var v0 = [
  "controls"
],
v1 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
};
return {
  "argumentDefinitions": [
    {
      "defaultValue": null,
      "kind": "LocalArgument",
      "name": "after"
    },
    {
      "defaultValue": null,
      "kind": "LocalArgument",
      "name": "before"
    },
    {
      "defaultValue": null,
      "kind": "LocalArgument",
      "name": "filter"
    },
    {
      "defaultValue": 20,
      "kind": "LocalArgument",
      "name": "first"
    },
    {
      "defaultValue": null,
      "kind": "LocalArgument",
      "name": "last"
    },
    {
      "defaultValue": null,
      "kind": "LocalArgument",
      "name": "order"
    }
  ],
  "kind": "Fragment",
  "metadata": {
    "connection": [
      {
        "count": null,
        "cursor": null,
        "direction": "bidirectional",
        "path": (v0/*: any*/)
      }
    ],
    "refetch": {
      "connection": {
        "forward": {
          "count": "first",
          "cursor": "after"
        },
        "backward": {
          "count": "last",
          "cursor": "before"
        },
        "path": (v0/*: any*/)
      },
      "fragmentPathInResult": [
        "node"
      ],
      "operation": MeasureControlsTabControlsQuery_graphql,
      "identifierInfo": {
        "identifierField": "id",
        "identifierQueryVariableName": "id"
      }
    }
  },
  "name": "MeasureControlsTabFragment",
  "selections": [
    (v1/*: any*/),
    {
      "alias": "controls",
      "args": [
        {
          "kind": "Variable",
          "name": "filter",
          "variableName": "filter"
        },
        {
          "kind": "Variable",
          "name": "orderBy",
          "variableName": "order"
        }
      ],
      "concreteType": "ControlConnection",
      "kind": "LinkedField",
      "name": "__MeasureControlsTab_controls_connection",
      "plural": false,
      "selections": [
        {
          "alias": null,
          "args": null,
          "concreteType": "ControlEdge",
          "kind": "LinkedField",
          "name": "edges",
          "plural": true,
          "selections": [
            {
              "alias": null,
              "args": null,
              "concreteType": "Control",
              "kind": "LinkedField",
              "name": "node",
              "plural": false,
              "selections": [
                (v1/*: any*/),
                {
                  "alias": "canCreateMeasureMapping",
                  "args": [
                    {
                      "kind": "Literal",
                      "name": "action",
                      "value": "core:control:create-measure-mapping"
                    }
                  ],
                  "kind": "ScalarField",
                  "name": "permission",
                  "storageKey": "permission(action:\"core:control:create-measure-mapping\")"
                },
                {
                  "alias": "canDeleteMeasureMapping",
                  "args": [
                    {
                      "kind": "Literal",
                      "name": "action",
                      "value": "core:control:delete-measure-mapping"
                    }
                  ],
                  "kind": "ScalarField",
                  "name": "permission",
                  "storageKey": "permission(action:\"core:control:delete-measure-mapping\")"
                },
                {
                  "args": null,
                  "kind": "FragmentSpread",
                  "name": "LinkedControlsCardFragment"
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
            },
            {
              "alias": null,
              "args": null,
              "kind": "ScalarField",
              "name": "hasPreviousPage",
              "storageKey": null
            },
            {
              "alias": null,
              "args": null,
              "kind": "ScalarField",
              "name": "startCursor",
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

(node as any).hash = "6cd1d286921188abc2a9fdc1c1e0ed08";

export default node;
