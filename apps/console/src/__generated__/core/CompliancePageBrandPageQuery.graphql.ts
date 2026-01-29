/**
 * @generated SignedSource<<be7955648260f44b4b42a2b342db2b05>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type CompliancePageBrandPageQuery$variables = {
  organizationId: string;
};
export type CompliancePageBrandPageQuery$data = {
  readonly organization: {
    readonly compliancePage?: {
      readonly darkLogoFileUrl: string | null | undefined;
      readonly logoFileUrl: string | null | undefined;
    } | null | undefined;
  };
};
export type CompliancePageBrandPageQuery = {
  response: CompliancePageBrandPageQuery$data;
  variables: CompliancePageBrandPageQuery$variables;
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
  "name": "logoFileUrl",
  "storageKey": null
},
v3 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "darkLogoFileUrl",
  "storageKey": null
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
    "name": "CompliancePageBrandPageQuery",
    "selections": [
      {
        "alias": "organization",
        "args": (v1/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          {
            "kind": "InlineFragment",
            "selections": [
              {
                "alias": "compliancePage",
                "args": null,
                "concreteType": "TrustCenter",
                "kind": "LinkedField",
                "name": "trustCenter",
                "plural": false,
                "selections": [
                  (v2/*: any*/),
                  (v3/*: any*/)
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
    "name": "CompliancePageBrandPageQuery",
    "selections": [
      {
        "alias": "organization",
        "args": (v1/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "__typename",
            "storageKey": null
          },
          {
            "kind": "InlineFragment",
            "selections": [
              {
                "alias": "compliancePage",
                "args": null,
                "concreteType": "TrustCenter",
                "kind": "LinkedField",
                "name": "trustCenter",
                "plural": false,
                "selections": [
                  (v2/*: any*/),
                  (v3/*: any*/),
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
    "cacheID": "17e4be042ed0e09a4b299d38f33dff32",
    "id": null,
    "metadata": {},
    "name": "CompliancePageBrandPageQuery",
    "operationKind": "query",
    "text": "query CompliancePageBrandPageQuery(\n  $organizationId: ID!\n) {\n  organization: node(id: $organizationId) {\n    __typename\n    ... on Organization {\n      compliancePage: trustCenter {\n        logoFileUrl\n        darkLogoFileUrl\n        id\n      }\n    }\n    id\n  }\n}\n"
  }
};
})();

(node as any).hash = "3364d36a8dd77d39d00d50f451a027a7";

export default node;
