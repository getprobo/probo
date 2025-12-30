/**
 * @generated SignedSource<<1123080dba735de72e2440f5dec008b5>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DeleteEvidenceInput = {
  evidenceId: string;
};
export type MeasureEvidencesTabDeleteMutation$variables = {
  connections: ReadonlyArray<string>;
  input: DeleteEvidenceInput;
};
export type MeasureEvidencesTabDeleteMutation$data = {
  readonly deleteEvidence: {
    readonly deletedEvidenceId: string;
  };
};
export type MeasureEvidencesTabDeleteMutation = {
  response: MeasureEvidencesTabDeleteMutation$data;
  variables: MeasureEvidencesTabDeleteMutation$variables;
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
  "name": "deletedEvidenceId",
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
    "name": "MeasureEvidencesTabDeleteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteEvidencePayload",
        "kind": "LinkedField",
        "name": "deleteEvidence",
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
    "name": "MeasureEvidencesTabDeleteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteEvidencePayload",
        "kind": "LinkedField",
        "name": "deleteEvidence",
        "plural": false,
        "selections": [
          (v3/*: any*/),
          {
            "alias": null,
            "args": null,
            "filters": null,
            "handle": "deleteEdge",
            "key": "",
            "kind": "ScalarHandle",
            "name": "deletedEvidenceId",
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
    "cacheID": "7c9ea7e14d3a342d9108026da877a854",
    "id": null,
    "metadata": {},
    "name": "MeasureEvidencesTabDeleteMutation",
    "operationKind": "mutation",
    "text": "mutation MeasureEvidencesTabDeleteMutation(\n  $input: DeleteEvidenceInput!\n) {\n  deleteEvidence(input: $input) {\n    deletedEvidenceId\n  }\n}\n"
  }
};
})();

(node as any).hash = "5b78978395d1b8a165a078f60f6aa001";

export default node;
