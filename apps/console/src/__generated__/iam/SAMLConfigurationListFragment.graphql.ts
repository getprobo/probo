/**
 * @generated SignedSource<<a9a22d22795f9c09c986748fe19dcccf>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
export type SAMLEnforcementPolicy = "OFF" | "OPTIONAL" | "REQUIRED";
import { FragmentRefs } from "relay-runtime";
export type SAMLConfigurationListFragment$data = {
  readonly samlConfigurations: {
    readonly edges: ReadonlyArray<{
      readonly node: {
        readonly canDelete: boolean;
        readonly canUpdate: boolean;
        readonly domainVerificationToken: string | null | undefined;
        readonly domainVerifiedAt: string | null | undefined;
        readonly emailDomain: string;
        readonly enforcementPolicy: SAMLEnforcementPolicy;
        readonly id: string;
        readonly testLoginUrl: string;
      };
    }>;
  };
  readonly " $fragmentType": "SAMLConfigurationListFragment";
};
export type SAMLConfigurationListFragment$key = {
  readonly " $data"?: SAMLConfigurationListFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"SAMLConfigurationListFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": {
    "connection": [
      {
        "count": null,
        "cursor": null,
        "direction": "forward",
        "path": [
          "samlConfigurations"
        ]
      }
    ]
  },
  "name": "SAMLConfigurationListFragment",
  "selections": [
    {
      "kind": "RequiredField",
      "field": {
        "alias": "samlConfigurations",
        "args": null,
        "concreteType": "SAMLConfigurationConnection",
        "kind": "LinkedField",
        "name": "__SAMLConfigurationListFragment_samlConfigurations_connection",
        "plural": false,
        "selections": [
          {
            "kind": "RequiredField",
            "field": {
              "alias": null,
              "args": null,
              "concreteType": "SAMLConfigurationEdge",
              "kind": "LinkedField",
              "name": "edges",
              "plural": true,
              "selections": [
                {
                  "alias": null,
                  "args": null,
                  "concreteType": "SAMLConfiguration",
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
                      "alias": null,
                      "args": null,
                      "kind": "ScalarField",
                      "name": "emailDomain",
                      "storageKey": null
                    },
                    {
                      "alias": null,
                      "args": null,
                      "kind": "ScalarField",
                      "name": "enforcementPolicy",
                      "storageKey": null
                    },
                    {
                      "alias": null,
                      "args": null,
                      "kind": "ScalarField",
                      "name": "domainVerificationToken",
                      "storageKey": null
                    },
                    {
                      "alias": null,
                      "args": null,
                      "kind": "ScalarField",
                      "name": "domainVerifiedAt",
                      "storageKey": null
                    },
                    {
                      "alias": null,
                      "args": null,
                      "kind": "ScalarField",
                      "name": "testLoginUrl",
                      "storageKey": null
                    },
                    {
                      "alias": "canUpdate",
                      "args": [
                        {
                          "kind": "Literal",
                          "name": "action",
                          "value": "iam:saml-configuration:update"
                        }
                      ],
                      "kind": "ScalarField",
                      "name": "permission",
                      "storageKey": "permission(action:\"iam:saml-configuration:update\")"
                    },
                    {
                      "alias": "canDelete",
                      "args": [
                        {
                          "kind": "Literal",
                          "name": "action",
                          "value": "iam:saml-configuration:delete"
                        }
                      ],
                      "kind": "ScalarField",
                      "name": "permission",
                      "storageKey": "permission(action:\"iam:saml-configuration:delete\")"
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
            "action": "THROW"
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
          }
        ],
        "storageKey": null
      },
      "action": "THROW"
    }
  ],
  "type": "Organization",
  "abstractKey": null
};

(node as any).hash = "3adfb13b2a908c93602b219ef1d0cf50";

export default node;
