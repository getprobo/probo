/**
 * @generated SignedSource<<681e1d8f8073e0efd781bdf322b893ab>>
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
export type ProcessingActivityRegistryGraphNodeQuery$variables = {
  processingActivityRegistryId: string;
};
export type ProcessingActivityRegistryGraphNodeQuery$data = {
  readonly node: {
    readonly audit?: {
      readonly framework: {
        readonly id: string;
        readonly name: string;
      };
      readonly id: string;
      readonly name: string | null | undefined;
    };
    readonly consentEvidenceLink?: string | null | undefined;
    readonly createdAt?: any;
    readonly dataProtectionImpactAssessment?: ProcessingActivityRegistryDataProtectionImpactAssessment;
    readonly dataSubjectCategory?: string | null | undefined;
    readonly id?: string;
    readonly internationalTransfers?: boolean;
    readonly lawfulBasis?: ProcessingActivityRegistryLawfulBasis;
    readonly location?: string | null | undefined;
    readonly name?: string;
    readonly organization?: {
      readonly id: string;
      readonly name: string;
    };
    readonly personalDataCategory?: string | null | undefined;
    readonly purpose?: string | null | undefined;
    readonly recipients?: string | null | undefined;
    readonly retentionPeriod?: string | null | undefined;
    readonly securityMeasures?: string | null | undefined;
    readonly specialOrCriminalData?: ProcessingActivityRegistrySpecialOrCriminalData;
    readonly transferImpactAssessment?: ProcessingActivityRegistryTransferImpactAssessment;
    readonly transferSafeguards?: ProcessingActivityRegistryTransferSafeguards;
    readonly updatedAt?: any;
  };
};
export type ProcessingActivityRegistryGraphNodeQuery = {
  response: ProcessingActivityRegistryGraphNodeQuery$data;
  variables: ProcessingActivityRegistryGraphNodeQuery$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "processingActivityRegistryId"
  }
],
v1 = [
  {
    "kind": "Variable",
    "name": "id",
    "variableName": "processingActivityRegistryId"
  }
],
v2 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
},
v3 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "name",
  "storageKey": null
},
v4 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "purpose",
  "storageKey": null
},
v5 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "dataSubjectCategory",
  "storageKey": null
},
v6 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "personalDataCategory",
  "storageKey": null
},
v7 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "specialOrCriminalData",
  "storageKey": null
},
v8 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "consentEvidenceLink",
  "storageKey": null
},
v9 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "lawfulBasis",
  "storageKey": null
},
v10 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "recipients",
  "storageKey": null
},
v11 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "location",
  "storageKey": null
},
v12 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "internationalTransfers",
  "storageKey": null
},
v13 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "transferSafeguards",
  "storageKey": null
},
v14 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "retentionPeriod",
  "storageKey": null
},
v15 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "securityMeasures",
  "storageKey": null
},
v16 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "dataProtectionImpactAssessment",
  "storageKey": null
},
v17 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "transferImpactAssessment",
  "storageKey": null
},
v18 = [
  (v2/*: any*/),
  (v3/*: any*/)
],
v19 = {
  "alias": null,
  "args": null,
  "concreteType": "Audit",
  "kind": "LinkedField",
  "name": "audit",
  "plural": false,
  "selections": [
    (v2/*: any*/),
    (v3/*: any*/),
    {
      "alias": null,
      "args": null,
      "concreteType": "Framework",
      "kind": "LinkedField",
      "name": "framework",
      "plural": false,
      "selections": (v18/*: any*/),
      "storageKey": null
    }
  ],
  "storageKey": null
},
v20 = {
  "alias": null,
  "args": null,
  "concreteType": "Organization",
  "kind": "LinkedField",
  "name": "organization",
  "plural": false,
  "selections": (v18/*: any*/),
  "storageKey": null
},
v21 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "createdAt",
  "storageKey": null
},
v22 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "updatedAt",
  "storageKey": null
};
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "ProcessingActivityRegistryGraphNodeQuery",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          {
            "kind": "InlineFragment",
            "selections": [
              (v2/*: any*/),
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
              (v19/*: any*/),
              (v20/*: any*/),
              (v21/*: any*/),
              (v22/*: any*/)
            ],
            "type": "ProcessingActivityRegistry",
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
    "name": "ProcessingActivityRegistryGraphNodeQuery",
    "selections": [
      {
        "alias": null,
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
          (v2/*: any*/),
          {
            "kind": "InlineFragment",
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
              (v19/*: any*/),
              (v20/*: any*/),
              (v21/*: any*/),
              (v22/*: any*/)
            ],
            "type": "ProcessingActivityRegistry",
            "abstractKey": null
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "74eed9b882f3bf3b11af0c15f65ef870",
    "id": null,
    "metadata": {},
    "name": "ProcessingActivityRegistryGraphNodeQuery",
    "operationKind": "query",
    "text": "query ProcessingActivityRegistryGraphNodeQuery(\n  $processingActivityRegistryId: ID!\n) {\n  node(id: $processingActivityRegistryId) {\n    __typename\n    ... on ProcessingActivityRegistry {\n      id\n      name\n      purpose\n      dataSubjectCategory\n      personalDataCategory\n      specialOrCriminalData\n      consentEvidenceLink\n      lawfulBasis\n      recipients\n      location\n      internationalTransfers\n      transferSafeguards\n      retentionPeriod\n      securityMeasures\n      dataProtectionImpactAssessment\n      transferImpactAssessment\n      audit {\n        id\n        name\n        framework {\n          id\n          name\n        }\n      }\n      organization {\n        id\n        name\n      }\n      createdAt\n      updatedAt\n    }\n    id\n  }\n}\n"
  }
};
})();

(node as any).hash = "979f4b05c845c90dd6359d24821e44fd";

export default node;
