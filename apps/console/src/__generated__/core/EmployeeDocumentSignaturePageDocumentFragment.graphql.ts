/**
 * @generated SignedSource<<30cf9ee1e00a276d2719e8d7d5ea0949>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type EmployeeDocumentSignaturePageDocumentFragment$data = {
  readonly id: string;
  readonly signed: boolean;
  readonly title: string;
  readonly versions: {
    readonly edges: ReadonlyArray<{
      readonly node: {
        readonly id: string;
        readonly " $fragmentSpreads": FragmentRefs<"VersionActionsFragment" | "VersionRowFragment">;
      };
    }>;
  };
  readonly " $fragmentType": "EmployeeDocumentSignaturePageDocumentFragment";
};
export type EmployeeDocumentSignaturePageDocumentFragment$key = {
  readonly " $data"?: EmployeeDocumentSignaturePageDocumentFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"EmployeeDocumentSignaturePageDocumentFragment">;
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
  "metadata": null,
  "name": "EmployeeDocumentSignaturePageDocumentFragment",
  "selections": [
    (v0/*: any*/),
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "title",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "signed",
      "storageKey": null
    },
    {
      "kind": "RequiredField",
      "field": {
        "alias": null,
        "args": [
          {
            "kind": "Literal",
            "name": "first",
            "value": 100
          },
          {
            "kind": "Literal",
            "name": "orderBy",
            "value": {
              "direction": "DESC",
              "field": "CREATED_AT"
            }
          }
        ],
        "concreteType": "DocumentVersionConnection",
        "kind": "LinkedField",
        "name": "versions",
        "plural": false,
        "selections": [
          {
            "kind": "RequiredField",
            "field": {
              "alias": null,
              "args": null,
              "concreteType": "DocumentVersionEdge",
              "kind": "LinkedField",
              "name": "edges",
              "plural": true,
              "selections": [
                {
                  "kind": "RequiredField",
                  "field": {
                    "alias": null,
                    "args": null,
                    "concreteType": "DocumentVersion",
                    "kind": "LinkedField",
                    "name": "node",
                    "plural": false,
                    "selections": [
                      (v0/*: any*/),
                      {
                        "args": null,
                        "kind": "FragmentSpread",
                        "name": "VersionActionsFragment"
                      },
                      {
                        "args": null,
                        "kind": "FragmentSpread",
                        "name": "VersionRowFragment"
                      }
                    ],
                    "storageKey": null
                  },
                  "action": "THROW"
                }
              ],
              "storageKey": null
            },
            "action": "THROW"
          }
        ],
        "storageKey": "versions(first:100,orderBy:{\"direction\":\"DESC\",\"field\":\"CREATED_AT\"})"
      },
      "action": "THROW"
    }
  ],
  "type": "SignableDocument",
  "abstractKey": null
};
})();

(node as any).hash = "aaa20cc1293a2b1ce5d760631c30804c";

export default node;
