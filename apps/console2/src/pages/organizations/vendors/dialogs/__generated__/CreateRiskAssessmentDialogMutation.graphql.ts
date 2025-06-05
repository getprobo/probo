/**
 * @generated SignedSource<<f877423fd1b767f837337e1b8a917e51>>
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
  assessedBy: string;
  businessImpact: BusinessImpact;
  dataSensitivity: DataSensitivity;
  expiresAt: any;
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
],
v3 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
};
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
                  (v3/*: any*/),
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "assessedAt",
                    "storageKey": null
                  },
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "People",
                    "kind": "LinkedField",
                    "name": "assessedBy",
                    "plural": false,
                    "selections": [
                      (v3/*: any*/),
                      {
                        "alias": null,
                        "args": null,
                        "kind": "ScalarField",
                        "name": "fullName",
                        "storageKey": null
                      }
                    ],
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
    "cacheID": "c55703ef9adb6c6aabf79aafe6cbeb89",
    "id": null,
    "metadata": {},
    "name": "CreateRiskAssessmentDialogMutation",
    "operationKind": "mutation",
    "text": "mutation CreateRiskAssessmentDialogMutation(\n  $input: CreateVendorRiskAssessmentInput!\n) {\n  createVendorRiskAssessment(input: $input) {\n    vendorRiskAssessmentEdge {\n      node {\n        ...VendorRiskAssessmentTabFragment_assessment\n        id\n      }\n    }\n  }\n}\n\nfragment VendorRiskAssessmentTabFragment_assessment on VendorRiskAssessment {\n  id\n  assessedAt\n  assessedBy {\n    id\n    fullName\n  }\n  expiresAt\n  dataSensitivity\n  businessImpact\n  notes\n}\n"
  }
};
})();

(node as any).hash = "1e6753dbe2856b5a1809c15a81eb9861";

export default node;
