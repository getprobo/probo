/**
 * @generated SignedSource<<dd129ccf57ef541ce24672d4b55d708f>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type UploadMeasureEvidenceInput = {
  file: any;
  measureId: string;
};
export type MeasureEvidencesTabUploadMutation$variables = {
  connections: ReadonlyArray<string>;
  input: UploadMeasureEvidenceInput;
};
export type MeasureEvidencesTabUploadMutation$data = {
  readonly uploadMeasureEvidence: {
    readonly evidenceEdge: {
      readonly node: {
        readonly id: string;
        readonly " $fragmentSpreads": FragmentRefs<"MeasureEvidencesTabFragment_evidence">;
      };
    };
  };
};
export type MeasureEvidencesTabUploadMutation = {
  response: MeasureEvidencesTabUploadMutation$data;
  variables: MeasureEvidencesTabUploadMutation$variables;
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
    "name": "MeasureEvidencesTabUploadMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "UploadMeasureEvidencePayload",
        "kind": "LinkedField",
        "name": "uploadMeasureEvidence",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "EvidenceEdge",
            "kind": "LinkedField",
            "name": "evidenceEdge",
            "plural": false,
            "selections": [
              {
                "alias": null,
                "args": null,
                "concreteType": "Evidence",
                "kind": "LinkedField",
                "name": "node",
                "plural": false,
                "selections": [
                  (v3/*: any*/),
                  {
                    "args": null,
                    "kind": "FragmentSpread",
                    "name": "MeasureEvidencesTabFragment_evidence"
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
    "name": "MeasureEvidencesTabUploadMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "UploadMeasureEvidencePayload",
        "kind": "LinkedField",
        "name": "uploadMeasureEvidence",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "EvidenceEdge",
            "kind": "LinkedField",
            "name": "evidenceEdge",
            "plural": false,
            "selections": [
              {
                "alias": null,
                "args": null,
                "concreteType": "Evidence",
                "kind": "LinkedField",
                "name": "node",
                "plural": false,
                "selections": [
                  (v3/*: any*/),
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "filename",
                    "storageKey": null
                  },
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "size",
                    "storageKey": null
                  },
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
                    "name": "createdAt",
                    "storageKey": null
                  },
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "fileUrl",
                    "storageKey": null
                  },
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "mimeType",
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
            "handle": "appendEdge",
            "key": "",
            "kind": "LinkedHandle",
            "name": "evidenceEdge",
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
    "cacheID": "facbff238f3ee10c220a5c80883836da",
    "id": null,
    "metadata": {},
    "name": "MeasureEvidencesTabUploadMutation",
    "operationKind": "mutation",
    "text": "mutation MeasureEvidencesTabUploadMutation(\n  $input: UploadMeasureEvidenceInput!\n) {\n  uploadMeasureEvidence(input: $input) {\n    evidenceEdge {\n      node {\n        id\n        ...MeasureEvidencesTabFragment_evidence\n      }\n    }\n  }\n}\n\nfragment MeasureEvidencesTabFragment_evidence on Evidence {\n  id\n  filename\n  size\n  type\n  createdAt\n  fileUrl\n  mimeType\n}\n"
  }
};
})();

(node as any).hash = "37283c7d6f25ed8c768ba9e0f93ebafb";

export default node;
