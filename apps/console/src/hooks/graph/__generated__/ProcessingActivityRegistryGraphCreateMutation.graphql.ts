/**
 * @generated SignedSource<<0eb58c6c3650563cef060921b0d7f65c>>
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
  transferSafeguards?: ProcessingActivityRegistryTransferSafeguards | null | undefined;
};
export type ProcessingActivityRegistryGraphCreateMutation$variables = {
  connections: ReadonlyArray<string>;
  input: CreateProcessingActivityRegistryInput;
};
export type ProcessingActivityRegistryGraphCreateMutation$data = {
  readonly createProcessingActivityRegistry: {
    readonly processingActivityRegistryEdge: {
      readonly node: {
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
        readonly transferSafeguards: ProcessingActivityRegistryTransferSafeguards | null | undefined;
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
          "name": "name",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "purpose",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "dataSubjectCategory",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "personalDataCategory",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "specialOrCriminalData",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "consentEvidenceLink",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "lawfulBasis",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "recipients",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "location",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "internationalTransfers",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "transferSafeguards",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "retentionPeriod",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "securityMeasures",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "dataProtectionImpactAssessment",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "transferImpactAssessment",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "createdAt",
          "storageKey": null
        }
      ],
      "storageKey": null
    }
  ],
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
          (v3/*: any*/)
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
          (v3/*: any*/),
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
    "cacheID": "22dc5c717c1e1de0a4bc8399a8e4f2eb",
    "id": null,
    "metadata": {},
    "name": "ProcessingActivityRegistryGraphCreateMutation",
    "operationKind": "mutation",
    "text": "mutation ProcessingActivityRegistryGraphCreateMutation(\n  $input: CreateProcessingActivityRegistryInput!\n) {\n  createProcessingActivityRegistry(input: $input) {\n    processingActivityRegistryEdge {\n      node {\n        id\n        name\n        purpose\n        dataSubjectCategory\n        personalDataCategory\n        specialOrCriminalData\n        consentEvidenceLink\n        lawfulBasis\n        recipients\n        location\n        internationalTransfers\n        transferSafeguards\n        retentionPeriod\n        securityMeasures\n        dataProtectionImpactAssessment\n        transferImpactAssessment\n        createdAt\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "000dcbeea412bc2b3863a37f58135df3";

export default node;
