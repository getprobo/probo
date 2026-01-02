/**
 * @generated SignedSource<<2271889b21f5a1f41c2e285a7a9ee1ca>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type BusinessImpact = "CRITICAL" | "HIGH" | "LOW" | "MEDIUM";
export type DataSensitivity = "CRITICAL" | "HIGH" | "LOW" | "MEDIUM" | "NONE";
export type CreateVendorRiskAssessmentInput = {
  businessImpact: BusinessImpact;
  dataSensitivity: DataSensitivity;
  expiresAt: string;
  notes?: string | null | undefined;
  vendorId: string;
};
export type CreateRiskAssessmentDialogMutation$variables = {
  connections: ReadonlyArray<string>;
  input: CreateVendorRiskAssessmentInput;
};
export type CreateRiskAssessmentDialogMutation$data = {
  readonly createVendorRiskAssessment: {
    readonly vendorRiskAssessmentEdge: {
      readonly node: {
        readonly " $fragmentSpreads": FragmentRefs<"VendorRiskAssessmentTabFragment_assessment">;
      };
    };
  };
};
export type CreateRiskAssessmentDialogMutation = {
  response: CreateRiskAssessmentDialogMutation$data;
  variables: CreateRiskAssessmentDialogMutation$variables;
};

const node: ConcreteRequest = (function(){
var v0 = {
  "defaultValue": null,
  "kind": "LocalArgument",
  "name": "connections"
},
v1 = {
  "defaultValue": null,
  "kind": "LocalArgument",
  "name": "input"
},
v2 = [
  {
    "kind": "Variable",
    "name": "input",
    "variableName": "input"
  }
];
return {
  "fragment": {
    "argumentDefinitions": [
      (v0/*: any*/),
      (v1/*: any*/)
    ],
    "kind": "Fragment",
    "metadata": null,
    "name": "CreateRiskAssessmentDialogMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateVendorRiskAssessmentPayload",
        "kind": "LinkedField",
        "name": "createVendorRiskAssessment",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "VendorRiskAssessmentEdge",
            "kind": "LinkedField",
            "name": "vendorRiskAssessmentEdge",
            "plural": false,
            "selections": [
              {
                "alias": null,
                "args": null,
                "concreteType": "VendorRiskAssessment",
                "kind": "LinkedField",
                "name": "node",
                "plural": false,
                "selections": [
                  {
                    "args": null,
                    "kind": "FragmentSpread",
                    "name": "VendorRiskAssessmentTabFragment_assessment"
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
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": [
      (v1/*: any*/),
      (v0/*: any*/)
    ],
    "kind": "Operation",
    "name": "CreateRiskAssessmentDialogMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateVendorRiskAssessmentPayload",
        "kind": "LinkedField",
        "name": "createVendorRiskAssessment",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "VendorRiskAssessmentEdge",
            "kind": "LinkedField",
            "name": "vendorRiskAssessmentEdge",
            "plural": false,
            "selections": [
              {
                "alias": null,
                "args": null,
                "concreteType": "VendorRiskAssessment",
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
                    "name": "createdAt",
                    "storageKey": null
                  },
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "expiresAt",
                    "storageKey": null
                  },
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "dataSensitivity",
                    "storageKey": null
                  },
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "businessImpact",
                    "storageKey": null
                  },
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "notes",
                    "storageKey": null
                  }
                ],
                "storageKey": null
              }
            ],
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "filters": null,
            "handle": "prependEdge",
            "key": "",
            "kind": "LinkedHandle",
            "name": "vendorRiskAssessmentEdge",
            "handleArgs": [
              {
                "kind": "Variable",
                "name": "connections",
                "variableName": "connections"
              }
            ]
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "1bd928ac46edc24d7b7cda3e5d7e2cb2",
    "id": null,
    "metadata": {},
    "name": "CreateRiskAssessmentDialogMutation",
    "operationKind": "mutation",
    "text": "mutation CreateRiskAssessmentDialogMutation(\n  $input: CreateVendorRiskAssessmentInput!\n) {\n  createVendorRiskAssessment(input: $input) {\n    vendorRiskAssessmentEdge {\n      node {\n        ...VendorRiskAssessmentTabFragment_assessment\n        id\n      }\n    }\n  }\n}\n\nfragment VendorRiskAssessmentTabFragment_assessment on VendorRiskAssessment {\n  id\n  createdAt\n  expiresAt\n  dataSensitivity\n  businessImpact\n  notes\n}\n"
  }
};
})();

(node as any).hash = "1e6753dbe2856b5a1809c15a81eb9861";

export default node;
