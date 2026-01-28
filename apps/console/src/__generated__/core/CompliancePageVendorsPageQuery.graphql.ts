/**
 * @generated SignedSource<<e0eb06c2450a1c6748294affef817ea1>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type CompliancePageVendorsPageQuery$variables = {
  organizationId: string;
};
export type CompliancePageVendorsPageQuery$data = {
  readonly organization: {
    readonly " $fragmentSpreads": FragmentRefs<"CompliancePageVendorListFragment">;
  };
};
export type CompliancePageVendorsPageQuery = {
  response: CompliancePageVendorsPageQuery$data;
  variables: CompliancePageVendorsPageQuery$variables;
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
  "name": "id",
  "storageKey": null
};
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "CompliancePageVendorsPageQuery",
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
            "args": null,
            "kind": "FragmentSpread",
            "name": "CompliancePageVendorListFragment"
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
    "name": "CompliancePageVendorsPageQuery",
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
                "alias": null,
                "args": [
                  {
                    "kind": "Literal",
                    "name": "first",
                    "value": 100
                  }
                ],
                "concreteType": "VendorConnection",
                "kind": "LinkedField",
                "name": "vendors",
                "plural": false,
                "selections": [
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "VendorEdge",
                    "kind": "LinkedField",
                    "name": "edges",
                    "plural": true,
                    "selections": [
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "Vendor",
                        "kind": "LinkedField",
                        "name": "node",
                        "plural": false,
                        "selections": [
                          (v2/*: any*/),
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "category",
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
                            "name": "showOnTrustCenter",
                            "storageKey": null
                          },
                          {
                            "alias": "canUpdate",
                            "args": [
                              {
                                "kind": "Literal",
                                "name": "action",
                                "value": "core:vendor:update"
                              }
                            ],
                            "kind": "ScalarField",
                            "name": "permission",
                            "storageKey": "permission(action:\"core:vendor:update\")"
                          }
                        ],
                        "storageKey": null
                      }
                    ],
                    "storageKey": null
                  }
                ],
                "storageKey": "vendors(first:100)"
              }
            ],
            "type": "Organization",
            "abstractKey": null
          },
          (v2/*: any*/)
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "83bd7ff13958372665e18d59f9636d55",
    "id": null,
    "metadata": {},
    "name": "CompliancePageVendorsPageQuery",
    "operationKind": "query",
    "text": "query CompliancePageVendorsPageQuery(\n  $organizationId: ID!\n) {\n  organization: node(id: $organizationId) {\n    __typename\n    ...CompliancePageVendorListFragment\n    id\n  }\n}\n\nfragment CompliancePageVendorListFragment on Organization {\n  vendors(first: 100) {\n    edges {\n      node {\n        id\n        ...CompliancePageVendorListItem_vendorFragment\n      }\n    }\n  }\n}\n\nfragment CompliancePageVendorListItem_vendorFragment on Vendor {\n  id\n  category\n  name\n  showOnTrustCenter\n  canUpdate: permission(action: \"core:vendor:update\")\n}\n"
  }
};
})();

(node as any).hash = "991a68cadfe4984bf303a75dcda66a53";

export default node;
