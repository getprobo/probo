/**
 * @generated SignedSource<<3e2b13f2b9a543d8393fdb7a1d7c680d>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type FrameworkDetailPageFragment$data = {
  readonly controls: {
    readonly edges: ReadonlyArray<{
      readonly node: {
        readonly description: string;
        readonly documents: {
          readonly __id: string;
          readonly edges: ReadonlyArray<{
            readonly node: {
              readonly id: string;
              readonly " $fragmentSpreads": FragmentRefs<"LinkedDocumentsCardFragment">;
            };
          }>;
        };
        readonly id: string;
        readonly measures: {
          readonly __id: string;
          readonly edges: ReadonlyArray<{
            readonly node: {
              readonly id: string;
              readonly " $fragmentSpreads": FragmentRefs<"LinkedMeasuresCardFragment">;
            };
          }>;
        };
        readonly name: string;
        readonly sectionTitle: string;
      };
    }>;
  };
  readonly description: string;
  readonly id: string;
  readonly name: string;
  readonly " $fragmentType": "FrameworkDetailPageFragment";
};
export type FrameworkDetailPageFragment$key = {
  readonly " $data"?: FrameworkDetailPageFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"FrameworkDetailPageFragment">;
};

const node: ReaderFragment = (function(){
var v0 = {
  "count": null,
  "cursor": null,
  "direction": "forward",
  "path": null
},
v1 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
},
v2 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "name",
  "storageKey": null
},
v3 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "description",
  "storageKey": null
},
v4 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "__typename",
  "storageKey": null
},
v5 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "cursor",
  "storageKey": null
},
v6 = {
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
v7 = {
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
};
return {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": {
    "connection": [
      (v0/*: any*/),
      (v0/*: any*/)
    ]
  },
  "name": "FrameworkDetailPageFragment",
  "selections": [
    (v1/*: any*/),
    (v2/*: any*/),
    (v3/*: any*/),
    {
      "alias": null,
      "args": [
        {
          "kind": "Literal",
          "name": "first",
          "value": 100
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
                (v1/*: any*/),
                {
                  "alias": null,
                  "args": null,
                  "kind": "ScalarField",
                  "name": "sectionTitle",
                  "storageKey": null
                },
                (v2/*: any*/),
                (v3/*: any*/),
                {
                  "alias": "measures",
                  "args": null,
                  "concreteType": "MeasureConnection",
                  "kind": "LinkedField",
                  "name": "__FrameworkDetailPage_measures_connection",
                  "plural": false,
                  "selections": [
                    {
                      "alias": null,
                      "args": null,
                      "concreteType": "MeasureEdge",
                      "kind": "LinkedField",
                      "name": "edges",
                      "plural": true,
                      "selections": [
                        {
                          "alias": null,
                          "args": null,
                          "concreteType": "Measure",
                          "kind": "LinkedField",
                          "name": "node",
                          "plural": false,
                          "selections": [
                            (v1/*: any*/),
                            {
                              "args": null,
                              "kind": "FragmentSpread",
                              "name": "LinkedMeasuresCardFragment"
                            },
                            (v4/*: any*/)
                          ],
                          "storageKey": null
                        },
                        (v5/*: any*/)
                      ],
                      "storageKey": null
                    },
                    (v6/*: any*/),
                    (v7/*: any*/)
                  ],
                  "storageKey": null
                },
                {
                  "alias": "documents",
                  "args": null,
                  "concreteType": "DocumentConnection",
                  "kind": "LinkedField",
                  "name": "__FrameworkDetailPage_documents_connection",
                  "plural": false,
                  "selections": [
                    {
                      "alias": null,
                      "args": null,
                      "concreteType": "DocumentEdge",
                      "kind": "LinkedField",
                      "name": "edges",
                      "plural": true,
                      "selections": [
                        {
                          "alias": null,
                          "args": null,
                          "concreteType": "Document",
                          "kind": "LinkedField",
                          "name": "node",
                          "plural": false,
                          "selections": [
                            (v1/*: any*/),
                            {
                              "args": null,
                              "kind": "FragmentSpread",
                              "name": "LinkedDocumentsCardFragment"
                            },
                            (v4/*: any*/)
                          ],
                          "storageKey": null
                        },
                        (v5/*: any*/)
                      ],
                      "storageKey": null
                    },
                    (v6/*: any*/),
                    (v7/*: any*/)
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
      "storageKey": "controls(first:100)"
    }
  ],
  "type": "Framework",
  "abstractKey": null
};
})();

(node as any).hash = "f0fe30240833b2da754f040f0c35aaaa";

export default node;
