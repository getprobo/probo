/**
 * @generated SignedSource<<ce29629a1c8d489463cc22edbd8b81de>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type PublishDocumentVersionInput = {
  changelog?: string | null | undefined;
  documentId: string;
};
export type DocumentDetailPagePublishMutation$variables = {
  input: PublishDocumentVersionInput;
};
export type DocumentDetailPagePublishMutation$data = {
  readonly publishDocumentVersion: {
    readonly document: {
      readonly id: string;
    };
  };
};
export type DocumentDetailPagePublishMutation = {
  response: DocumentDetailPagePublishMutation$data;
  variables: DocumentDetailPagePublishMutation$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "input"
  }
],
v1 = [
  {
    "alias": null,
    "args": [
      {
        "kind": "Variable",
        "name": "input",
        "variableName": "input"
      }
    ],
    "concreteType": "PublishDocumentVersionPayload",
    "kind": "LinkedField",
    "name": "publishDocumentVersion",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "Document",
        "kind": "LinkedField",
        "name": "document",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "id",
            "storageKey": null
          }
        ],
        "storageKey": null
      }
    ],
    "storageKey": null
  }
];
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "DocumentDetailPagePublishMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "DocumentDetailPagePublishMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "bbee8d20f444cdc0dbb012e8fecfa9c1",
    "id": null,
    "metadata": {},
    "name": "DocumentDetailPagePublishMutation",
    "operationKind": "mutation",
    "text": "mutation DocumentDetailPagePublishMutation(\n  $input: PublishDocumentVersionInput!\n) {\n  publishDocumentVersion(input: $input) {\n    document {\n      id\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "2f5cc9855133b614896a1ac3768e669f";

export default node;
