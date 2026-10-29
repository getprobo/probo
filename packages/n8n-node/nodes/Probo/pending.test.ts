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

// The n8n package cannot depend on vitest. This file is excluded from the build.
// eslint-disable-next-line @n8n/community-nodes/no-restricted-imports
import { describe, expect, it } from 'vitest';
import type { IDataObject } from 'n8n-workflow';

import { collectPendingApprovals, collectPendingSignatures, type GraphQLRequest } from './pending';

const now = new Date('2026-10-29T15:38:00.000Z');

function signatureNode(id: string) {
	return {
		id,
		state: 'REQUESTED',
		requestedAt: '2026-10-26T15:38:00.000Z',
		signedBy: { id: 'person-1', fullName: 'Ada Lovelace', emailAddress: 'ada@example.com' },
	};
}

describe('collectPendingSignatures', () => {
	it('returns one record per waiting signature, including a follow-up page', async () => {
		const calls: string[] = [];
		const request: GraphQLRequest = async (query, variables) => {
			calls.push(query);
			if (query.includes('query PendingSignatureDocuments')) {
				expect(variables.organizationId).toBe('org-1');
				return {
					data: {
						node: {
							documents: {
								edges: [
									{
										node: {
											id: 'doc-1',
											title: 'Access policy',
											versions: {
												edges: [
													{
														node: {
															id: 'ver-1',
															title: 'Access policy',
															major: 1,
															minor: 0,
															status: 'PUBLISHED',
															signatures: {
																edges: [{ node: signatureNode('sig-1') }],
																pageInfo: { hasNextPage: true, endCursor: 'sig-cursor' },
															},
														},
													},
												],
												pageInfo: { hasNextPage: false },
											},
										},
									},
								],
								pageInfo: { hasNextPage: false },
							},
						},
					},
				};
			}

			expect(variables.after).toBe('sig-cursor');
			expect(variables.documentVersionId).toBe('ver-1');
			return {
				data: {
					node: {
						signatures: {
							edges: [{ node: signatureNode('sig-2') }],
							pageInfo: { hasNextPage: false },
						},
					},
				},
			};
		};

		const items = await collectPendingSignatures(request, {
			organizationId: 'org-1',
			returnAll: true,
			limit: 50,
			now,
		});

		expect(calls.length).toBe(2);
		expect(items.map((item) => item.signatureId)).toEqual(['sig-1', 'sig-2']);
		expect((items[0].person as IDataObject).emailAddress).toBe('ada@example.com');
		expect((items[0].document as IDataObject).title).toBe('Access policy');
		expect(items[0].documentId).toBe('doc-1');
		expect(items[0].versionId).toBe('ver-1');
		expect(items[0].waitingDuration).toBe('P3D');
		expect(items[0].waitingSeconds).toBe(259200);
	});

	it('stops once the requested limit is reached', async () => {
		const request: GraphQLRequest = async () => ({
			data: {
				node: {
					documents: {
						edges: [
							{
								node: {
									id: 'doc-1',
									title: 'Access policy',
									versions: {
										edges: [
											{
												node: {
													id: 'ver-1',
													title: 'Access policy',
													status: 'PUBLISHED',
													signatures: {
														edges: [
															{ node: signatureNode('sig-1') },
															{ node: signatureNode('sig-2') },
														],
														pageInfo: { hasNextPage: true, endCursor: 'more' },
													},
												},
											},
										],
										pageInfo: { hasNextPage: false },
									},
								},
							},
						],
						pageInfo: { hasNextPage: false },
					},
				},
			},
		});

		const items = await collectPendingSignatures(request, {
			organizationId: 'org-1',
			returnAll: false,
			limit: 1,
			now,
		});

		expect(items.map((item) => item.id)).toEqual(['sig-1']);
	});
});

describe('collectPendingApprovals', () => {
	it('returns the approver, document, and how long the decision has been waiting', async () => {
		const request: GraphQLRequest = async (query) => {
			expect(query).toMatch(/PENDING_APPROVAL/);
			expect(query).toMatch(/states: \[PENDING\]/);
			return {
				data: {
					node: {
						documents: {
							edges: [
								{
									node: {
										id: 'doc-1',
										title: 'Incident policy',
										versions: {
											edges: [
												{
													node: {
														id: 'ver-1',
														title: 'Incident policy',
														major: 2,
														minor: 1,
														status: 'PENDING_APPROVAL',
														approvalQuorums: {
															edges: [
																{
																	node: {
																		id: 'quorum-1',
																		status: 'PENDING',
																		decisions: {
																			edges: [
																				{
																					node: {
																						id: 'decision-1',
																						state: 'PENDING',
																						comment: null,
																						createdAt: '2026-10-28T15:38:00.000Z',
																						approver: {
																							id: 'person-2',
																							fullName: 'Grace Hopper',
																							emailAddress: 'grace@example.com',
																						},
																					},
																				},
																			],
																			pageInfo: { hasNextPage: false },
																		},
																	},
																},
															],
															pageInfo: { hasNextPage: false },
														},
													},
												},
											],
											pageInfo: { hasNextPage: false },
										},
									},
								},
							],
							pageInfo: { hasNextPage: false },
						},
					},
				},
			};
		};

		const items = await collectPendingApprovals(request, {
			organizationId: 'org-1',
			returnAll: false,
			limit: 50,
			now,
		});

		expect(items.length).toBe(1);
		expect(items[0].approvalDecisionId).toBe('decision-1');
		expect((items[0].person as IDataObject).emailAddress).toBe('grace@example.com');
		expect(items[0].documentId).toBe('doc-1');
		expect(items[0].versionId).toBe('ver-1');
		expect(items[0].quorumId).toBe('quorum-1');
		expect(items[0].waitingDuration).toBe('P1D');
	});
});
