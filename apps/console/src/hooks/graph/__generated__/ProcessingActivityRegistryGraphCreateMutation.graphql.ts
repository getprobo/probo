/**
 * @generated SignedSource<<8372cc0960f7642aa8619dbc091dc4de>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type ProcessingActivityRegistryDataProtectionImpactAssessment = "NEEDED" | "NOT_NEEDED";
export type ProcessingActivityRegistryLawfulBasis = "CONSENT" | "CONTRACTUAL_NECESSITY" | "LEGAL_OBLIGATION" | "LEGITIMATE_INTEREST" | "PUBLIC_TASK" | "VITAL_INTERESTS";
export type ProcessingActivityRegistrySpecialOrCriminalData = "NO" | "POSSIBLE" | "YES";
export type ProcessingActivityRegistryTransferImpactAssessment = "NEEDED" | "NOT_NEEDED";
export type ProcessingActivityRegistryTransferSafeguards = "ADEQUACY_DECISION" | "BINDING_CORPORATE_RULES" | "CERTIFICATION_MECHANISMS" | "CODES_OF_CONDUCT" | "DEROGATIONS" | "STANDARD_CONTRACTUAL_CLAUSES";
export type CreateProcessingActivityRegistryInput = {
  auditId: string;
  consentEvidenceLink?: string | null | undefined;
  dataProtectionImpactAssessment: ProcessingActivityRegistryDataProtectionImpactAssessment;
  dataSubjectCategory?: string | null | undefined;
  internationalTransfers: boolean;
  lawfulBasis: ProcessingActivityRegistryLawfulBasis;
  location?: string | null | undefined;
  name: string;
  organizationId: string;
  personalDataCategory?: string | null | undefined;
  purpose?: string | null | undefined;
  recipients?: string | null | undefined;
  retentionPeriod?: string | null | undefined;
  securityMeasures?: string | null | undefined;
  specialOrCriminalData: ProcessingActivityRegistrySpecialOrCriminalData;
  transferImpactAssessment: ProcessingActivityRegistryTransferImpactAssessment;
  transferSafeguards: ProcessingActivityRegistryTransferSafeguards;
};
export type ProcessingActivityRegistryGraphCreateMutation$variables = {
  connections: ReadonlyArray<string>;
  input: CreateProcessingActivityRegistryInput;
};
export type ProcessingActivityRegistryGraphCreateMutation$data = {
  readonly createProcessingActivityRegistry: {
    readonly processingActivityRegistryEdge: {
      readonly node: {
        readonly audit: {
          readonly framework: {
            readonly name: string;
          };
          readonly id: string;
          readonly name: string | null | undefined;
        };
        readonly consentEvidenceLink: string | null | undefined;
        readonly createdAt: any;
        readonly dataProtectionImpactAssessment: ProcessingActivityRegistryDataProtectionImpactAssessment;
        readonly dataSubjectCategory: string | null | undefined;
        readonly id: string;
        readonly internationalTransfers: boolean;
        readonly lawfulBasis: ProcessingActivityRegistryLawfulBasis;
        readonly location: string | null | undefined;
        readonly name: string;
        readonly personalDataCategory: string | null | undefined;
        readonly purpose: string | null | undefined;
        readonly recipients: string | null | undefined;
        readonly retentionPeriod: string | null | undefined;
        readonly securityMeasures: string | null | undefined;
        readonly specialOrCriminalData: ProcessingActivityRegistrySpecialOrCriminalData;
        readonly transferImpactAssessment: ProcessingActivityRegistryTransferImpactAssessment;
        readonly transferSafeguards: ProcessingActivityRegistryTransferSafeguards;
      };
    };
  };
};
export type ProcessingActivityRegistryGraphCreateMutation = {
  response: ProcessingActivityRegistryGraphCreateMutation$data;
  variables: ProcessingActivityRegistryGraphCreateMutation$variables;
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
},
v4 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "name",
  "storageKey": null
},
v5 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "purpose",
  "storageKey": null
},
v6 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "dataSubjectCategory",
  "storageKey": null
},
v7 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "personalDataCategory",
  "storageKey": null
},
v8 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "specialOrCriminalData",
  "storageKey": null
},
v9 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "consentEvidenceLink",
  "storageKey": null
},
v10 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "lawfulBasis",
  "storageKey": null
},
v11 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "recipients",
  "storageKey": null
},
v12 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "location",
  "storageKey": null
},
v13 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "internationalTransfers",
  "storageKey": null
},
v14 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "transferSafeguards",
  "storageKey": null
},
v15 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "retentionPeriod",
  "storageKey": null
},
v16 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "securityMeasures",
  "storageKey": null
},
v17 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "dataProtectionImpactAssessment",
  "storageKey": null
},
v18 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "transferImpactAssessment",
  "storageKey": null
},
v19 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "createdAt",
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
    "name": "ProcessingActivityRegistryGraphCreateMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateProcessingActivityRegistryPayload",
        "kind": "LinkedField",
        "name": "createProcessingActivityRegistry",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "ProcessingActivityRegistryEdge",
            "kind": "LinkedField",
            "name": "processingActivityRegistryEdge",
            "plural": false,
            "selections": [
              {
                "alias": null,
                "args": null,
                "concreteType": "ProcessingActivityRegistry",
                "kind": "LinkedField",
                "name": "node",
                "plural": false,
                "selections": [
                  (v3/*: any*/),
                  (v4/*: any*/),
                  (v5/*: any*/),
                  (v6/*: any*/),
                  (v7/*: any*/),
                  (v8/*: any*/),
                  (v9/*: any*/),
                  (v10/*: any*/),
                  (v11/*: any*/),
                  (v12/*: any*/),
                  (v13/*: any*/),
                  (v14/*: any*/),
                  (v15/*: any*/),
                  (v16/*: any*/),
                  (v17/*: any*/),
                  (v18/*: any*/),
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "Audit",
                    "kind": "LinkedField",
                    "name": "audit",
                    "plural": false,
                    "selections": [
                      (v3/*: any*/),
                      (v4/*: any*/),
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "Framework",
                        "kind": "LinkedField",
                        "name": "framework",
                        "plural": false,
                        "selections": [
                          (v4/*: any*/)
                        ],
                        "storageKey": null
                      }
                    ],
                    "storageKey": null
                  },
                  (v19/*: any*/)
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
    "name": "ProcessingActivityRegistryGraphCreateMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateProcessingActivityRegistryPayload",
        "kind": "LinkedField",
        "name": "createProcessingActivityRegistry",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "ProcessingActivityRegistryEdge",
            "kind": "LinkedField",
            "name": "processingActivityRegistryEdge",
            "plural": false,
            "selections": [
              {
                "alias": null,
                "args": null,
                "concreteType": "ProcessingActivityRegistry",
                "kind": "LinkedField",
                "name": "node",
                "plural": false,
                "selections": [
                  (v3/*: any*/),
                  (v4/*: any*/),
                  (v5/*: any*/),
                  (v6/*: any*/),
                  (v7/*: any*/),
                  (v8/*: any*/),
                  (v9/*: any*/),
                  (v10/*: any*/),
                  (v11/*: any*/),
                  (v12/*: any*/),
                  (v13/*: any*/),
                  (v14/*: any*/),
                  (v15/*: any*/),
                  (v16/*: any*/),
                  (v17/*: any*/),
                  (v18/*: any*/),
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "Audit",
                    "kind": "LinkedField",
                    "name": "audit",
                    "plural": false,
                    "selections": [
                      (v3/*: any*/),
                      (v4/*: any*/),
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "Framework",
                        "kind": "LinkedField",
                        "name": "framework",
                        "plural": false,
                        "selections": [
                          (v4/*: any*/),
                          (v3/*: any*/)
                        ],
                        "storageKey": null
                      }
                    ],
                    "storageKey": null
                  },
                  (v19/*: any*/)
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
            "name": "processingActivityRegistryEdge",
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
    "cacheID": "9942999c0282d64a6250774f8a26f1f9",
    "id": null,
    "metadata": {},
    "name": "ProcessingActivityRegistryGraphCreateMutation",
    "operationKind": "mutation",
    "text": "mutation ProcessingActivityRegistryGraphCreateMutation(\n  $input: CreateProcessingActivityRegistryInput!\n) {\n  createProcessingActivityRegistry(input: $input) {\n    processingActivityRegistryEdge {\n      node {\n        id\n        name\n        purpose\n        dataSubjectCategory\n        personalDataCategory\n        specialOrCriminalData\n        consentEvidenceLink\n        lawfulBasis\n        recipients\n        location\n        internationalTransfers\n        transferSafeguards\n        retentionPeriod\n        securityMeasures\n        dataProtectionImpactAssessment\n        transferImpactAssessment\n        audit {\n          id\n          name\n          framework {\n            name\n            id\n          }\n        }\n        createdAt\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "7563c6519bb6f538e3c211322a5a638f";

export default node;
