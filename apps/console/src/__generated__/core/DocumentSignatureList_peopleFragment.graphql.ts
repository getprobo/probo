/**
 * @generated SignedSource<<9d5bb5c52a22efeea7cc554200a7a153>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type DocumentSignatureList_peopleFragment$data = {
  readonly profiles: {
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
      "alias": null,
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
      "concreteType": "ProfileConnection",
      "kind": "LinkedField",
      "name": "profiles",
      "plural": false,
      "selections": [
        {
          "alias": null,
          "args": null,
          "concreteType": "ProfileEdge",
          "kind": "LinkedField",
          "name": "edges",
          "plural": true,
          "selections": [
            {
              "alias": null,
              "args": null,
              "concreteType": "Profile",
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

(node as any).hash = "cf9788ef7f8667bb436ef3776618d3f3";

export default node;
