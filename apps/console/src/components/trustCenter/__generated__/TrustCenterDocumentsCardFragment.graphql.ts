/**
 * @generated SignedSource<<711a0b67f3e93814ae3297a2578373c6>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
export type DocumentStatus = "DRAFT" | "PUBLISHED";
export type DocumentType = "ISMS" | "OTHER" | "POLICY";
import { FragmentRefs } from "relay-runtime";
export type TrustCenterDocumentsCardFragment$data = {
  readonly createdAt: any;
  readonly documentType: DocumentType;
  readonly id: string;
  readonly showOnTrustCenter: boolean;
  readonly title: string;
  readonly versions: {
    readonly edges: ReadonlyArray<{
      readonly node: {
        readonly id: string;
        readonly status: DocumentStatus;
      };
    }>;
  };
  readonly " $fragmentType": "TrustCenterDocumentsCardFragment";
};
export type TrustCenterDocumentsCardFragment$key = {
  readonly " $data"?: TrustCenterDocumentsCardFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"TrustCenterDocumentsCardFragment">;
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
  "name": "TrustCenterDocumentsCardFragment",
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
      "name": "createdAt",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "documentType",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "showOnTrustCenter",
      "storageKey": null
    },
    {
      "alias": null,
      "args": [
        {
          "kind": "Literal",
          "name": "first",
          "value": 1
        }
      ],
      "concreteType": "DocumentVersionConnection",
      "kind": "LinkedField",
      "name": "versions",
      "plural": false,
      "selections": [
        {
          "alias": null,
          "args": null,
          "concreteType": "DocumentVersionEdge",
          "kind": "LinkedField",
          "name": "edges",
          "plural": true,
          "selections": [
            {
              "alias": null,
              "args": null,
              "concreteType": "DocumentVersion",
              "kind": "LinkedField",
              "name": "node",
              "plural": false,
              "selections": [
                (v0/*: any*/),
                {
                  "alias": null,
                  "args": null,
                  "kind": "ScalarField",
                  "name": "status",
                  "storageKey": null
                }
              ],
              "storageKey": null
            }
          ],
          "storageKey": null
        }
      ],
      "storageKey": "versions(first:1)"
    }
  ],
  "type": "Document",
  "abstractKey": null
};
})();

(node as any).hash = "88d5e5e3a2d9f4b730428efcea644f25";

export default node;
