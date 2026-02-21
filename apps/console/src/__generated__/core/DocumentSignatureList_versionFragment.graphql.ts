/**
 * @generated SignedSource<<628394c29c246b6018084f14b3c0000b>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type DocumentSignatureList_versionFragment$data = {
  readonly id: string;
  readonly signatures: {
    readonly __id: string;
    readonly edges: ReadonlyArray<{
      readonly node: {
        readonly id: string;
        readonly signedBy: {
          readonly id: string;
        };
        readonly " $fragmentSpreads": FragmentRefs<"DocumentSignatureListItemFragment">;
      };
    }>;
  };
  readonly " $fragmentSpreads": FragmentRefs<"DocumentSignaturePlaceholder_versionFragment">;
  readonly " $fragmentType": "DocumentSignatureList_versionFragment";
};
export type DocumentSignatureList_versionFragment$key = {
  readonly " $data"?: DocumentSignatureList_versionFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"DocumentSignatureList_versionFragment">;
};

import DocumentSignatureListQuery_graphql from './DocumentSignatureListQuery.graphql';

const node: ReaderFragment = (function(){
var v0 = [
  "signatures"
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
      "defaultValue": 1000,
      "kind": "LocalArgument",
      "name": "count"
    },
    {
      "defaultValue": null,
      "kind": "LocalArgument",
      "name": "cursor"
    },
    {
      "defaultValue": null,
      "kind": "LocalArgument",
      "name": "signatureFilter"
    }
  ],
  "kind": "Fragment",
  "metadata": {
    "connection": [
      {
        "count": "count",
        "cursor": "cursor",
        "direction": "forward",
        "path": (v0/*: any*/)
      }
    ],
    "refetch": {
      "connection": {
        "forward": {
          "count": "count",
          "cursor": "cursor"
        },
        "backward": null,
        "path": (v0/*: any*/)
      },
      "fragmentPathInResult": [
        "node"
      ],
      "operation": DocumentSignatureListQuery_graphql,
      "identifierInfo": {
        "identifierField": "id",
        "identifierQueryVariableName": "id"
      }
    }
  },
  "name": "DocumentSignatureList_versionFragment",
  "selections": [
    {
      "args": null,
      "kind": "FragmentSpread",
      "name": "DocumentSignaturePlaceholder_versionFragment"
    },
    {
      "alias": "signatures",
      "args": [
        {
          "kind": "Variable",
          "name": "filter",
          "variableName": "signatureFilter"
        }
      ],
      "concreteType": "DocumentVersionSignatureConnection",
      "kind": "LinkedField",
      "name": "__DocumentSignaturesTab_signatures_connection",
      "plural": false,
      "selections": [
        {
          "alias": null,
          "args": null,
          "concreteType": "DocumentVersionSignatureEdge",
          "kind": "LinkedField",
          "name": "edges",
          "plural": true,
          "selections": [
            {
              "alias": null,
              "args": null,
              "concreteType": "DocumentVersionSignature",
              "kind": "LinkedField",
              "name": "node",
              "plural": false,
              "selections": [
                (v1/*: any*/),
                {
                  "alias": null,
                  "args": null,
                  "concreteType": "Profile",
                  "kind": "LinkedField",
                  "name": "signedBy",
                  "plural": false,
                  "selections": [
                    (v1/*: any*/)
                  ],
                  "storageKey": null
                },
                {
                  "args": null,
                  "kind": "FragmentSpread",
                  "name": "DocumentSignatureListItemFragment"
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
    },
    (v1/*: any*/)
  ],
  "type": "DocumentVersion",
  "abstractKey": null
};
})();

(node as any).hash = "cc402c3b46eacfb04db1186bbd744f67";

export default node;
