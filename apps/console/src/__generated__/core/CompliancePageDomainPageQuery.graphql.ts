/**
 * @generated SignedSource<<f31f58950034fd90764103d8c1e24f30>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type CompliancePageDomainPageQuery$variables = {
  organizationId: string;
};
export type CompliancePageDomainPageQuery$data = {
  readonly organization: {
    readonly __typename: "Organization";
    readonly canCreateCustomDomain: boolean;
    readonly customDomain: {
      readonly " $fragmentSpreads": FragmentRefs<"CompliancePageDomainCardFragment">;
    } | null | undefined;
  } | {
    // This will never be '%other', but we need some
    // value in case none of the concrete values match.
    readonly __typename: "%other";
  };
};
export type CompliancePageDomainPageQuery = {
  response: CompliancePageDomainPageQuery$data;
  variables: CompliancePageDomainPageQuery$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "organizationId"
  }
],
v1 = [
  {
    "kind": "Variable",
    "name": "id",
    "variableName": "organizationId"
  }
],
v2 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "__typename",
  "storageKey": null
},
v3 = {
  "alias": "canCreateCustomDomain",
  "args": [
    {
      "kind": "Literal",
      "name": "action",
      "value": "core:custom-domain:create"
    }
  ],
  "kind": "ScalarField",
  "name": "permission",
  "storageKey": "permission(action:\"core:custom-domain:create\")"
},
v4 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
};
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "CompliancePageDomainPageQuery",
    "selections": [
      {
        "alias": "organization",
        "args": (v1/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          (v2/*: any*/),
          {
            "kind": "InlineFragment",
            "selections": [
              (v3/*: any*/),
              {
                "alias": null,
                "args": null,
                "concreteType": "CustomDomain",
                "kind": "LinkedField",
                "name": "customDomain",
                "plural": false,
                "selections": [
                  {
                    "args": null,
                    "kind": "FragmentSpread",
                    "name": "CompliancePageDomainCardFragment"
                  }
                ],
                "storageKey": null
              }
            ],
            "type": "Organization",
            "abstractKey": null
          }
        ],
        "storageKey": null
      }
    ],
    "type": "Query",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "CompliancePageDomainPageQuery",
    "selections": [
      {
        "alias": "organization",
        "args": (v1/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          (v2/*: any*/),
          {
            "kind": "InlineFragment",
            "selections": [
              (v3/*: any*/),
              {
                "alias": null,
                "args": null,
                "concreteType": "CustomDomain",
                "kind": "LinkedField",
                "name": "customDomain",
                "plural": false,
                "selections": [
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "domain",
                    "storageKey": null
                  },
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "sslStatus",
                    "storageKey": null
                  },
                  {
                    "alias": "canDelete",
                    "args": [
                      {
                        "kind": "Literal",
                        "name": "action",
                        "value": "core:custom-domain:delete"
                      }
                    ],
                    "kind": "ScalarField",
                    "name": "permission",
                    "storageKey": "permission(action:\"core:custom-domain:delete\")"
                  },
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "DNSRecordInstruction",
                    "kind": "LinkedField",
                    "name": "dnsRecords",
                    "plural": true,
                    "selections": [
                      {
                        "alias": null,
                        "args": null,
                        "kind": "ScalarField",
                        "name": "type",
                        "storageKey": null
                      },
                      {
                        "alias": null,
                        "args": null,
                        "kind": "ScalarField",
                        "name": "name",
                        "storageKey": null
                      },
                      {
                        "alias": null,
                        "args": null,
                        "kind": "ScalarField",
                        "name": "value",
                        "storageKey": null
                      },
                      {
                        "alias": null,
                        "args": null,
                        "kind": "ScalarField",
                        "name": "ttl",
                        "storageKey": null
                      },
                      {
                        "alias": null,
                        "args": null,
                        "kind": "ScalarField",
                        "name": "purpose",
                        "storageKey": null
                      }
                    ],
                    "storageKey": null
                  },
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "sslExpiresAt",
                    "storageKey": null
                  },
                  (v4/*: any*/)
                ],
                "storageKey": null
              }
            ],
            "type": "Organization",
            "abstractKey": null
          },
          (v4/*: any*/)
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "f916852f512fa1e1b06d73c608acad3a",
    "id": null,
    "metadata": {},
    "name": "CompliancePageDomainPageQuery",
    "operationKind": "query",
    "text": "query CompliancePageDomainPageQuery(\n  $organizationId: ID!\n) {\n  organization: node(id: $organizationId) {\n    __typename\n    ... on Organization {\n      canCreateCustomDomain: permission(action: \"core:custom-domain:create\")\n      customDomain {\n        ...CompliancePageDomainCardFragment\n        id\n      }\n    }\n    id\n  }\n}\n\nfragment CompliancePageDomainCardFragment on CustomDomain {\n  domain\n  sslStatus\n  canDelete: permission(action: \"core:custom-domain:delete\")\n  ...CompliancePageDomainDialogFragment\n}\n\nfragment CompliancePageDomainDialogFragment on CustomDomain {\n  sslStatus\n  domain\n  dnsRecords {\n    type\n    name\n    value\n    ttl\n    purpose\n  }\n  sslExpiresAt\n}\n"
  }
};
})();

(node as any).hash = "86eb03a06b0217c125eadf8cf661e384";

export default node;
