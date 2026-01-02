/**
 * @generated SignedSource<<6f67465a1bac4f3cac9c891b9c4cc9a1>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type ProcessingActivitiesPageTIAFragment$data = {
  readonly id: string;
  readonly transferImpactAssessments: {
    readonly __id: string;
    readonly edges: ReadonlyArray<{
      readonly node: {
        readonly createdAt: string;
        readonly dataSubjects: string | null | undefined;
        readonly id: string;
        readonly localLawRisk: string | null | undefined;
        readonly processingActivity: {
          readonly id: string;
          readonly name: string;
        };
        readonly transfer: string | null | undefined;
        readonly updatedAt: string;
      };
    }>;
    readonly pageInfo: {
      readonly endCursor: string | null | undefined;
      readonly hasNextPage: boolean;
    };
    readonly totalCount: number;
  };
  readonly " $fragmentType": "ProcessingActivitiesPageTIAFragment";
};
export type ProcessingActivitiesPageTIAFragment$key = {
  readonly " $data"?: ProcessingActivitiesPageTIAFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"ProcessingActivitiesPageTIAFragment">;
};

import ProcessingActivitiesPageTIARefetchQuery_graphql from './ProcessingActivitiesPageTIARefetchQuery.graphql';

const node: ReaderFragment = (function(){
var v0 = [
  "transferImpactAssessments"
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
      "defaultValue": 10,
      "kind": "LocalArgument",
      "name": "first"
    },
    {
      "defaultValue": null,
      "kind": "LocalArgument",
      "name": "snapshotId"
    }
  ],
  "kind": "Fragment",
  "metadata": {
    "connection": [
      {
        "count": "first",
        "cursor": "after",
        "direction": "forward",
        "path": (v0/*: any*/)
      }
    ],
    "refetch": {
      "connection": {
        "forward": {
          "count": "first",
          "cursor": "after"
        },
        "backward": null,
        "path": (v0/*: any*/)
      },
      "fragmentPathInResult": [
        "node"
      ],
      "operation": ProcessingActivitiesPageTIARefetchQuery_graphql,
      "identifierInfo": {
        "identifierField": "id",
        "identifierQueryVariableName": "id"
      }
    }
  },
  "name": "ProcessingActivitiesPageTIAFragment",
  "selections": [
    (v1/*: any*/),
    {
      "alias": "transferImpactAssessments",
      "args": [
        {
          "fields": [
            {
              "kind": "Variable",
              "name": "snapshotId",
              "variableName": "snapshotId"
            }
          ],
          "kind": "ObjectValue",
          "name": "filter"
        }
      ],
      "concreteType": "TransferImpactAssessmentConnection",
      "kind": "LinkedField",
      "name": "__ProcessingActivitiesPage_transferImpactAssessments_connection",
      "plural": false,
      "selections": [
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "totalCount",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "concreteType": "TransferImpactAssessmentEdge",
          "kind": "LinkedField",
          "name": "edges",
          "plural": true,
          "selections": [
            {
              "alias": null,
              "args": null,
              "concreteType": "TransferImpactAssessment",
              "kind": "LinkedField",
              "name": "node",
              "plural": false,
              "selections": [
                (v1/*: any*/),
                {
                  "alias": null,
                  "args": null,
                  "kind": "ScalarField",
                  "name": "dataSubjects",
                  "storageKey": null
                },
                {
                  "alias": null,
                  "args": null,
                  "kind": "ScalarField",
                  "name": "transfer",
                  "storageKey": null
                },
                {
                  "alias": null,
                  "args": null,
                  "kind": "ScalarField",
                  "name": "localLawRisk",
                  "storageKey": null
                },
                {
                  "alias": null,
                  "args": null,
                  "concreteType": "ProcessingActivity",
                  "kind": "LinkedField",
                  "name": "processingActivity",
                  "plural": false,
                  "selections": [
                    (v1/*: any*/),
                    {
                      "alias": null,
                      "args": null,
                      "kind": "ScalarField",
                      "name": "name",
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
                  "alias": null,
                  "args": null,
                  "kind": "ScalarField",
                  "name": "updatedAt",
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
              "name": "hasNextPage",
              "storageKey": null
            },
            {
              "alias": null,
              "args": null,
              "kind": "ScalarField",
              "name": "endCursor",
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
})();

(node as any).hash = "2c6575fe24698cc6b14d9d7a88cf434b";

export default node;
