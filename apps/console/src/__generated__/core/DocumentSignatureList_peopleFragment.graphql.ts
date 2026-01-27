/**
 * @generated SignedSource<<3b60ebae6765024147f97c157f6251fe>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type DocumentSignatureList_peopleFragment$data = {
  readonly people: {
    readonly edges: ReadonlyArray<{
      readonly node: {
        readonly id: string;
        readonly " $fragmentSpreads": FragmentRefs<"DocumentSignaturePlaceholder_personFragment">;
      };
    }>;
  };
  readonly " $fragmentSpreads": FragmentRefs<"DocumentSignaturePlaceholder_organizationFragment">;
  readonly " $fragmentType": "DocumentSignatureList_peopleFragment";
};
export type DocumentSignatureList_peopleFragment$key = {
  readonly " $data"?: DocumentSignatureList_peopleFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"DocumentSignatureList_peopleFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [
    {
      "defaultValue": null,
      "kind": "LocalArgument",
      "name": "filter"
    }
  ],
  "kind": "Fragment",
  "metadata": null,
  "name": "DocumentSignatureList_peopleFragment",
  "selections": [
    {
      "args": null,
      "kind": "FragmentSpread",
      "name": "DocumentSignaturePlaceholder_organizationFragment"
    },
    {
      "alias": "people",
      "args": [
        {
          "kind": "Variable",
          "name": "filter",
          "variableName": "filter"
        },
        {
          "kind": "Literal",
          "name": "first",
          "value": 1000
        },
        {
          "kind": "Literal",
          "name": "orderBy",
          "value": {
            "direction": "ASC",
            "field": "FULL_NAME"
          }
        }
      ],
      "concreteType": "PeopleConnection",
      "kind": "LinkedField",
      "name": "peoples",
      "plural": false,
      "selections": [
        {
          "alias": null,
          "args": null,
          "concreteType": "PeopleEdge",
          "kind": "LinkedField",
          "name": "edges",
          "plural": true,
          "selections": [
            {
              "alias": null,
              "args": null,
              "concreteType": "People",
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
                  "args": null,
                  "kind": "FragmentSpread",
                  "name": "DocumentSignaturePlaceholder_personFragment"
                }
              ],
              "storageKey": null
            }
          ],
          "storageKey": null
        }
      ],
      "storageKey": null
    }
  ],
  "type": "Organization",
  "abstractKey": null
};

(node as any).hash = "6c0a2b965d1440c2e0fd8ec9ce3c5e85";

export default node;
