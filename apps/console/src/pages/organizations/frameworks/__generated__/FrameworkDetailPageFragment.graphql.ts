/**
 * @generated SignedSource<<c6e55821a322dceae5408ac3898cfa48>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
export type ControlStatus = "EXCLUDED" | "INCLUDED";
import { FragmentRefs } from "relay-runtime";
export type FrameworkDetailPageFragment$data = {
  readonly controls: {
    readonly __id: string;
    readonly edges: ReadonlyArray<{
      readonly node: {
        readonly exclusionJustification: string | null | undefined;
        readonly id: string;
        readonly name: string;
        readonly sectionTitle: string;
        readonly status: ControlStatus;
      };
    }>;
  };
  readonly description: string;
  readonly id: string;
  readonly name: string;
  readonly organization: {
    readonly name: string;
  };
  readonly " $fragmentType": "FrameworkDetailPageFragment";
};
export type FrameworkDetailPageFragment$key = {
  readonly " $data"?: FrameworkDetailPageFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"FrameworkDetailPageFragment">;
};

const node: ReaderFragment = (function(){
var v0 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
},
v1 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "name",
  "storageKey": null
};
return {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "FrameworkDetailPageFragment",
  "selections": [
    (v0/*: any*/),
    (v1/*: any*/),
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
      "concreteType": "Organization",
      "kind": "LinkedField",
      "name": "organization",
      "plural": false,
      "selections": [
        (v1/*: any*/)
      ],
      "storageKey": null
    },
    {
      "alias": null,
      "args": [
        {
          "kind": "Literal",
          "name": "first",
          "value": 250
        },
        {
          "kind": "Literal",
          "name": "orderBy",
          "value": {
            "direction": "ASC",
            "field": "SECTION_TITLE"
          }
        }
      ],
      "concreteType": "ControlConnection",
      "kind": "LinkedField",
      "name": "controls",
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
                (v0/*: any*/),
                {
                  "alias": null,
                  "args": null,
                  "kind": "ScalarField",
                  "name": "sectionTitle",
                  "storageKey": null
                },
                (v1/*: any*/),
                {
                  "alias": null,
                  "args": null,
                  "kind": "ScalarField",
                  "name": "status",
                  "storageKey": null
                },
                {
                  "alias": null,
                  "args": null,
                  "kind": "ScalarField",
                  "name": "exclusionJustification",
                  "storageKey": null
                }
              ],
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
      "storageKey": "controls(first:250,orderBy:{\"direction\":\"ASC\",\"field\":\"SECTION_TITLE\"})"
    }
  ],
  "type": "Framework",
  "abstractKey": null
};
})();

(node as any).hash = "c31fd5c0f675d2524f736d69c4ebebe3";

export default node;
