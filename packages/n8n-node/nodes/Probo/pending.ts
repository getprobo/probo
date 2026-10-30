// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import type { IDataObject } from 'n8n-workflow';
import { personFields, profileSelection, waitingFields } from './output';

const PENDING_PAGE_SIZE = 20;

const signatureFieldsFragment = `
fragment SignatureFields on DocumentVersionSignature {
	id
	state
	requestedAt
	signedBy {
		${profileSelection}
	}
}
`;

const decisionFieldsFragment = `
fragment DecisionFields on DocumentVersionApprovalDecision {
	id
	state
	comment
	createdAt
	approver {
		${profileSelection}
	}
}
`;

export type GraphQLRequest = (query: string, variables: IDataObject) => Promise<IDataObject>;

export interface PendingOptions {
	organizationId: string;
	returnAll: boolean;
	limit: number;
	now: Date;
}

interface Connection<T> {
	edges?: Array<{ node?: T }>;
	pageInfo?: {
		hasNextPage?: boolean;
		endCursor?: string | null;
	};
}

const pendingSignatureDocumentsQuery = `
	query PendingSignatureDocuments($organizationId: ID!, $first: Int, $after: CursorKey, $nestedFirst: Int!) {
		node(id: $organizationId) {
			... on Organization {
				documents(first: $first, after: $after, filter: { status: [ACTIVE] }) {
					edges {
						node {
							id
							title
							versions(first: $nestedFirst, filter: { statuses: [PUBLISHED] }) {
								edges {
									node {
										id
										title
										major
										minor
										status
										signatures(first: $nestedFirst, filter: { states: [REQUESTED] }) {
											edges {
												node {
													...SignatureFields
												}
											}
											pageInfo {
												hasNextPage
												endCursor
											}
										}
									}
								}
								pageInfo {
									hasNextPage
									endCursor
								}
							}
						}
					}
					pageInfo {
						hasNextPage
						endCursor
					}
				}
			}
		}
	}
	${signatureFieldsFragment}
`;

const pendingSignatureVersionsQuery = `
	query PendingSignatureVersions($documentId: ID!, $first: Int, $after: CursorKey, $nestedFirst: Int!) {
		node(id: $documentId) {
			... on Document {
				versions(first: $first, after: $after, filter: { statuses: [PUBLISHED] }) {
					edges {
						node {
							id
							title
							major
							minor
							status
							signatures(first: $nestedFirst, filter: { states: [REQUESTED] }) {
								edges {
									node {
										...SignatureFields
									}
								}
								pageInfo {
									hasNextPage
									endCursor
								}
							}
						}
					}
					pageInfo {
						hasNextPage
						endCursor
					}
				}
			}
		}
	}
	${signatureFieldsFragment}
`;

const pendingSignaturesQuery = `
	query PendingSignatures($documentVersionId: ID!, $first: Int, $after: CursorKey) {
		node(id: $documentVersionId) {
			... on DocumentVersion {
				signatures(first: $first, after: $after, filter: { states: [REQUESTED] }) {
					edges {
						node {
							...SignatureFields
						}
					}
					pageInfo {
						hasNextPage
						endCursor
					}
				}
			}
		}
	}
	${signatureFieldsFragment}
`;

const pendingApprovalDocumentsQuery = `
	query PendingApprovalDocuments($organizationId: ID!, $first: Int, $after: CursorKey, $nestedFirst: Int!) {
		node(id: $organizationId) {
			... on Organization {
				documents(first: $first, after: $after, filter: { status: [ACTIVE] }) {
					edges {
						node {
							id
							title
							versions(first: $nestedFirst, filter: { statuses: [PENDING_APPROVAL] }) {
								edges {
									node {
										id
										title
										major
										minor
										status
										approvalQuorums(first: $nestedFirst) {
											edges {
												node {
													id
													status
													decisions(first: $nestedFirst, filter: { states: [PENDING] }) {
														edges {
															node {
																...DecisionFields
															}
														}
														pageInfo {
															hasNextPage
															endCursor
														}
													}
												}
											}
											pageInfo {
												hasNextPage
												endCursor
											}
										}
									}
								}
								pageInfo {
									hasNextPage
									endCursor
								}
							}
						}
					}
					pageInfo {
						hasNextPage
						endCursor
					}
				}
			}
		}
	}
	${decisionFieldsFragment}
`;

const pendingApprovalVersionsQuery = `
	query PendingApprovalVersions($documentId: ID!, $first: Int, $after: CursorKey, $nestedFirst: Int!) {
		node(id: $documentId) {
			... on Document {
				versions(first: $first, after: $after, filter: { statuses: [PENDING_APPROVAL] }) {
					edges {
						node {
							id
							title
							major
							minor
							status
							approvalQuorums(first: $nestedFirst) {
								edges {
									node {
										id
										status
										decisions(first: $nestedFirst, filter: { states: [PENDING] }) {
											edges {
												node {
													...DecisionFields
												}
											}
											pageInfo {
												hasNextPage
												endCursor
											}
										}
									}
								}
								pageInfo {
									hasNextPage
									endCursor
								}
							}
						}
					}
					pageInfo {
						hasNextPage
						endCursor
					}
				}
			}
		}
	}
	${decisionFieldsFragment}
`;

const pendingApprovalQuorumsQuery = `
	query PendingApprovalQuorums($documentVersionId: ID!, $first: Int, $after: CursorKey, $nestedFirst: Int!) {
		node(id: $documentVersionId) {
			... on DocumentVersion {
				approvalQuorums(first: $first, after: $after) {
					edges {
						node {
							id
							status
							decisions(first: $nestedFirst, filter: { states: [PENDING] }) {
								edges {
									node {
										...DecisionFields
									}
								}
								pageInfo {
									hasNextPage
									endCursor
								}
							}
						}
					}
					pageInfo {
						hasNextPage
						endCursor
					}
				}
			}
		}
	}
	${decisionFieldsFragment}
`;

const pendingApprovalDecisionsQuery = `
	query PendingApprovalDecisions($approvalQuorumId: ID!, $first: Int, $after: CursorKey) {
		node(id: $approvalQuorumId) {
			... on DocumentVersionApprovalQuorum {
				decisions(first: $first, after: $after, filter: { states: [PENDING] }) {
					edges {
						node {
							...DecisionFields
						}
					}
					pageInfo {
						hasNextPage
						endCursor
					}
				}
			}
		}
	}
	${decisionFieldsFragment}
`;

function nodesOf<T>(connection: Connection<T> | undefined): T[] {
	return (connection?.edges ?? [])
		.map((edge) => edge.node)
		.filter((node): node is T => node !== undefined && node !== null);
}

function graphqlNode(response: IDataObject): IDataObject | undefined {
	const data = response.data;
	if (!data || typeof data !== 'object' || Array.isArray(data)) {
		return undefined;
	}
	const node = (data as IDataObject).node;
	if (!node || typeof node !== 'object' || Array.isArray(node)) {
		return undefined;
	}
	return node as IDataObject;
}

function asConnection(value: unknown): Connection<IDataObject> | undefined {
	if (!value || typeof value !== 'object' || Array.isArray(value)) {
		return undefined;
	}
	return value as Connection<IDataObject>;
}

function pageSize(options: PendingOptions, collected: number): number {
	if (options.returnAll) {
		return PENDING_PAGE_SIZE;
	}
	const remaining = Math.floor(options.limit - collected);
	return Math.max(1, Math.min(PENDING_PAGE_SIZE, remaining));
}

async function walkConnection(
	initial: Connection<IDataObject> | undefined,
	more: (after: string) => Promise<Connection<IDataObject> | undefined>,
	visit: (node: IDataObject) => Promise<boolean>,
): Promise<boolean> {
	let connection = initial;
	const seen = new Set<string>();

	while (connection) {
		for (const node of nodesOf(connection)) {
			if (await visit(node)) {
				return true;
			}
		}

		const endCursor = connection.pageInfo?.endCursor ?? undefined;
		if (!connection.pageInfo?.hasNextPage || !endCursor || seen.has(endCursor)) {
			return false;
		}

		seen.add(endCursor);
		connection = await more(endCursor);
	}

	return false;
}

async function eachDocumentPage(
	request: GraphQLRequest,
	query: string,
	options: PendingOptions,
	collected: () => number,
	visit: (document: IDataObject) => Promise<boolean>,
): Promise<void> {
	let after: string | undefined;
	const seen = new Set<string>();

	while (true) {
		const response = await request(query, {
			organizationId: options.organizationId,
			first: PENDING_PAGE_SIZE,
			nestedFirst: pageSize(options, collected()),
			after,
		});
		const documents = asConnection(graphqlNode(response)?.documents);
		for (const document of nodesOf(documents)) {
			if (await visit(document)) {
				return;
			}
		}

		const endCursor = documents?.pageInfo?.endCursor ?? undefined;
		if (!documents?.pageInfo?.hasNextPage || !endCursor || seen.has(endCursor)) {
			return;
		}

		seen.add(endCursor);
		after = endCursor;
	}
}

function reachedLimit(options: PendingOptions, count: number): boolean {
	return !options.returnAll && count >= options.limit;
}

function versionSummary(version: IDataObject): IDataObject {
	return {
		id: version.id,
		title: version.title ?? null,
		major: version.major ?? null,
		minor: version.minor ?? null,
		status: version.status ?? null,
	};
}

function documentSummary(document: IDataObject): IDataObject {
	return {
		id: document.id,
		title: document.title ?? null,
	};
}

export function pendingSignatureRecord(
	document: IDataObject,
	version: IDataObject,
	signature: IDataObject,
	now: Date,
): IDataObject {
	const signedBy = signature.signedBy;
	const person =
		signedBy && typeof signedBy === 'object' && !Array.isArray(signedBy)
			? personFields(signedBy as IDataObject)
			: personFields(undefined);

	return {
		id: signature.id,
		signatureId: signature.id,
		state: signature.state,
		person,
		document: documentSummary(document),
		documentId: document.id,
		versionId: version.id,
		version: versionSummary(version),
		...waitingFields(signature.requestedAt, now),
	};
}

export function pendingApprovalRecord(
	document: IDataObject,
	version: IDataObject,
	quorum: IDataObject,
	decision: IDataObject,
	now: Date,
): IDataObject {
	const approver = decision.approver;
	const person =
		approver && typeof approver === 'object' && !Array.isArray(approver)
			? personFields(approver as IDataObject)
			: personFields(undefined);

	return {
		id: decision.id,
		approvalDecisionId: decision.id,
		state: decision.state,
		comment: decision.comment ?? null,
		person,
		document: documentSummary(document),
		documentId: document.id,
		versionId: version.id,
		version: versionSummary(version),
		quorumId: quorum.id,
		...waitingFields(decision.createdAt, now),
	};
}

export async function collectPendingSignatures(
	request: GraphQLRequest,
	options: PendingOptions,
): Promise<IDataObject[]> {
	const items: IDataObject[] = [];

	await eachDocumentPage(
		request,
		pendingSignatureDocumentsQuery,
		options,
		() => items.length,
		async (document) => {
			return walkConnection(
				asConnection(document.versions),
				async (after) => {
					const size = pageSize(options, items.length);
					const response = await request(pendingSignatureVersionsQuery, {
						documentId: document.id,
						first: size,
						nestedFirst: size,
						after,
					});
					return asConnection(graphqlNode(response)?.versions);
				},
				async (version) => {
					return walkConnection(
						asConnection(version.signatures),
						async (after) => {
							const response = await request(pendingSignaturesQuery, {
								documentVersionId: version.id,
								first: pageSize(options, items.length),
								after,
							});
							return asConnection(graphqlNode(response)?.signatures);
						},
						async (signature) => {
							if (reachedLimit(options, items.length)) {
								return true;
							}
							items.push(pendingSignatureRecord(document, version, signature, options.now));
							return reachedLimit(options, items.length);
						},
					);
				},
			);
		},
	);

	return items;
}

export async function collectPendingApprovals(
	request: GraphQLRequest,
	options: PendingOptions,
): Promise<IDataObject[]> {
	const items: IDataObject[] = [];

	await eachDocumentPage(
		request,
		pendingApprovalDocumentsQuery,
		options,
		() => items.length,
		async (document) => {
			return walkConnection(
				asConnection(document.versions),
				async (after) => {
					const size = pageSize(options, items.length);
					const response = await request(pendingApprovalVersionsQuery, {
						documentId: document.id,
						first: size,
						nestedFirst: size,
						after,
					});
					return asConnection(graphqlNode(response)?.versions);
				},
				async (version) => {
					return walkConnection(
						asConnection(version.approvalQuorums),
						async (after) => {
							const size = pageSize(options, items.length);
							const response = await request(pendingApprovalQuorumsQuery, {
								documentVersionId: version.id,
								first: size,
								nestedFirst: size,
								after,
							});
							return asConnection(graphqlNode(response)?.approvalQuorums);
						},
						async (quorum) => {
							return walkConnection(
								asConnection(quorum.decisions),
								async (after) => {
									const response = await request(pendingApprovalDecisionsQuery, {
										approvalQuorumId: quorum.id,
										first: pageSize(options, items.length),
										after,
									});
									return asConnection(graphqlNode(response)?.decisions);
								},
								async (decision) => {
									if (reachedLimit(options, items.length)) {
										return true;
									}
									items.push(
										pendingApprovalRecord(document, version, quorum, decision, options.now),
									);
									return reachedLimit(options, items.length);
								},
							);
						},
					);
				},
			);
		},
	);

	return items;
}
