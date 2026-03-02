/**
 * @generated SignedSource<<3eedb29b73b6e2c450d59f1d1ae0d88a>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type StateOfApplicabilityControlsTabFragment$data = {
  readonly applicabilityStatements: {
    readonly __id: string;
    readonly edges: ReadonlyArray<{
      readonly node: {
        readonly applicability: boolean;
        readonly control: {
          readonly bestPractice: boolean;
          readonly contractual: boolean;
          readonly framework: {
            readonly id: string;
            readonly name: string;
          };
          readonly id: string;
          readonly name: string;
          readonly organization: {
            readonly id: string;
          } | null | undefined;
          readonly regulatory: boolean;
          readonly riskAssessment: boolean;
          readonly sectionTitle: string;
        };
        readonly id: string;
        readonly justification: string;
      };
    }>;
  };
  readonly canCreateApplicabilityStatement: boolean;
  readonly canDeleteApplicabilityStatement: boolean;
  readonly canUpdateApplicabilityStatement: boolean;
  readonly id: string;
  readonly organization: {
    readonly id: string;
  } | null | undefined;
  readonly " $fragmentType": "StateOfApplicabilityControlsTabFragment";
};
export type StateOfApplicabilityControlsTabFragment$key = {
  readonly " $data"?: StateOfApplicabilityControlsTabFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"StateOfApplicabilityControlsTabFragment">;
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
  "concreteType": "Organization",
  "kind": "LinkedField",
  "name": "organization",
  "plural": false,
  "selections": [
    (v0/*: any*/)
  ],
  "storageKey": null
},
v2 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "name",
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
          "applicabilityStatements"
        ]
      }
    ]
  },
  "name": "StateOfApplicabilityControlsTabFragment",
  "selections": [
    (v0/*: any*/),
    (v1/*: any*/),
    {
      "alias": "canCreateApplicabilityStatement",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:applicability-statement:create"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"core:applicability-statement:create\")"
    },
    {
      "alias": "canUpdateApplicabilityStatement",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:applicability-statement:update"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"core:applicability-statement:update\")"
    },
    {
      "alias": "canDeleteApplicabilityStatement",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:applicability-statement:delete"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"core:applicability-statement:delete\")"
    },
    {
      "alias": "applicabilityStatements",
      "args": [
        {
          "kind": "Literal",
          "name": "orderBy",
          "value": {
            "direction": "ASC",
            "field": "CONTROL_SECTION_TITLE"
          }
        }
      ],
      "concreteType": "ApplicabilityStatementConnection",
      "kind": "LinkedField",
      "name": "__StateOfApplicabilityControlsTab_applicabilityStatements_connection",
      "plural": false,
      "selections": [
        {
          "alias": null,
          "args": null,
          "concreteType": "ApplicabilityStatementEdge",
          "kind": "LinkedField",
          "name": "edges",
          "plural": true,
          "selections": [
            {
              "alias": null,
              "args": null,
              "concreteType": "ApplicabilityStatement",
              "kind": "LinkedField",
              "name": "node",
              "plural": false,
              "selections": [
                (v0/*: any*/),
                {
                  "alias": null,
                  "args": null,
                  "kind": "ScalarField",
                  "name": "applicability",
                  "storageKey": null
                },
                {
                  "alias": null,
                  "args": null,
                  "kind": "ScalarField",
                  "name": "justification",
                  "storageKey": null
                },
                {
                  "alias": null,
                  "args": null,
                  "concreteType": "Control",
                  "kind": "LinkedField",
                  "name": "control",
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
                    (v2/*: any*/),
                    {
                      "alias": null,
                      "args": null,
                      "kind": "ScalarField",
                      "name": "bestPractice",
                      "storageKey": null
                    },
                    {
                      "alias": null,
                      "args": null,
                      "kind": "ScalarField",
                      "name": "regulatory",
                      "storageKey": null
                    },
                    {
                      "alias": null,
                      "args": null,
                      "kind": "ScalarField",
                      "name": "contractual",
                      "storageKey": null
                    },
                    {
                      "alias": null,
                      "args": null,
                      "kind": "ScalarField",
                      "name": "riskAssessment",
                      "storageKey": null
                    },
                    {
                      "alias": null,
                      "args": null,
                      "concreteType": "Framework",
                      "kind": "LinkedField",
                      "name": "framework",
                      "plural": false,
                      "selections": [
                        (v0/*: any*/),
                        (v2/*: any*/)
                      ],
                      "storageKey": null
                    },
                    (v1/*: any*/)
                  ],
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
      "storageKey": "__StateOfApplicabilityControlsTab_applicabilityStatements_connection(orderBy:{\"direction\":\"ASC\",\"field\":\"CONTROL_SECTION_TITLE\"})"
    }
  ],
  "type": "StateOfApplicability",
  "abstractKey": null
};
})();

(node as any).hash = "b80671ffb8f18997cb72a9beddd6cad3";

export default node;
